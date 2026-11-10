package middlewares

import (
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"fmt"
	"slices"
	"strings"
	"time"

	"example.com/go-service-kit/cmd/manage/internal/cache"
	"example.com/go-service-kit/cmd/manage/internal/config"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/utils/jwt"
	"example.com/go-service-kit/utils/tool"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
)

// 同一个旧令牌的并发续期只生成一次新令牌。
var adminTokenRefreshGroup singleflight.Group

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := checkToken(c)
		if !ok {
			c.JSON(response.HTTP_INVALID_TOKEN, gin.H{"error": help.T(c, "Login expired")})
			c.Abort()
			return
		}
		if claims.Status != model.ADMIN_STATUS_ON {
			c.JSON(response.HTTP_BAN_USER, gin.H{"error": help.T(c, "Login expired")})
			c.Abort()
			return
		}
		if claims.IsSuper == model.ADMIN_IS_SUPER_Y {
			c.Next()
			return
		}
		var role model.Role
		if err := (&repository.Role{}).GetById(int64(claims.RoleId), &role); err != nil || role.Status != model.ROLE_STATUS_NORMAL {
			c.JSON(response.HTTP_BAN_USER, gin.H{"error": help.T(c, "角色不存在或已禁用")})
			c.Abort()
			return
		}
		if !checkApi(c, claims) {
			c.JSON(response.HTTP_NOT_ALLOWED_API, gin.H{"error": help.T(c, "No API Permission")})
			c.Abort()
			return
		}
		c.Next()
	}
}

func checkToken(c *gin.Context) (*jwt.Claims, bool) {
	tokenString := c.GetHeader("Authorization")
	token, claims, err := jwt.ParseToken(tokenString)
	if token == nil || claims == nil {
		return nil, false
	}
	if err != nil {
		if validation, ok := err.(*jwtgo.ValidationError); !ok || validation.Errors != jwtgo.ValidationErrorExpired {
			return nil, false
		}
	}
	if claims.AdminId <= 0 || claims.Id == "" || claims.IssuedAt <= 0 || claims.ExpiresAt <= 0 {
		return nil, false
	}
	ctx := c.Request.Context()
	revokedAt, err := cache.GetAdminTokenRevokedAt(ctx, int64(claims.AdminId))
	if err != nil || claims.IssuedAt <= revokedAt {
		return nil, false
	}
	invalid, err := cache.ExistInvalidAdminToken(ctx, claims.Id)
	if err != nil || invalid {
		return nil, false
	}
	singleLogin := config.Config.System().Admin_Single_Login_Switch == 1
	if singleLogin {
		lastJTI, err := cache.GetAdminLastJwtToken(ctx, int64(claims.AdminId))
		if err != nil || claims.Id != lastJTI {
			return nil, false
		}
	}
	var current model.Admin
	if err := (&repository.Admin{}).GetById(int64(claims.AdminId), &current); err != nil || current.Status != model.ADMIN_STATUS_ON || !jwt.MatchesCredential(claims, &current) {
		return nil, false
	}
	claims.Username, claims.RoleId, claims.IsSuper, claims.Status = current.Username, current.RoleId, current.IsSuper, current.Status
	if claims.ExpiresAt <= time.Now().Unix() {
		if time.Since(time.Unix(claims.ExpiresAt, 0)) >= jwt.RefreshGrace {
			return nil, false
		}
		refreshKey := fmt.Sprintf("admin-refresh:%d:%s", claims.AdminId, claims.Id)
		value, err, _ := adminTokenRefreshGroup.Do(refreshKey, func() (any, error) {
			current.Timezone = claims.Timezone
			jti := tool.GetUUID()
			next, err := jwt.CreateUserToken(&current, jti)
			if err != nil {
				return "", err
			}
			if singleLogin {
				if err := cache.SetAdminLastJwtToken(ctx, int64(current.Id), jti, jwt.GetAccessExpireDuration()+jwt.RefreshGrace); err != nil {
					return "", err
				}
			}
			// 旧令牌续期后不能再次用于换取新令牌，包括未开启单点登录的情况。
			if err := jwt.InvalidateToken(ctx, tokenString); err != nil {
				return "", err
			}
			return next, nil
		})
		if err != nil {
			return nil, false
		}
		tokenString = "Bearer " + value.(string)
		_, claims, err = jwt.ParseToken(tokenString)
		if err != nil {
			return nil, false
		}
		c.Header("Authorization", value.(string))
		// 后续退出操作须撤销本次实际使用的新令牌。
		c.Request.Header.Set("Authorization", tokenString)
	}
	c.Set("admin", model.Admin{
		Id: claims.AdminId, Username: claims.Username, Status: claims.Status,
		RoleId: claims.RoleId, IsSuper: claims.IsSuper, Timezone: claims.Timezone,
	})
	return claims, true
}

// checkApi 只信任服务端路由和HTTP方法，公共操作使用固定的最小白名单。
func checkApi(c *gin.Context, claims *jwt.Claims) bool {
	path := strings.TrimPrefix(c.FullPath(), "/"+strings.Trim(config.Router_Prefix, "/"))
	switch c.Request.Method + " " + path {
	case "GET /common/init", "POST /auth/logout", "POST /auth/reset-pwd", "POST /auth/verify":
		return true
	}
	permission := strings.ReplaceAll(strings.Trim(path, "/"), "/", ".") + "." + strings.ToLower(c.Request.Method)
	ids, err := (&repository.Role{}).GetPermission(claims.RoleId)
	if err != nil {
		return false
	}
	menus, err := (&repository.Menu{}).ReadAll()
	if err != nil {
		return false
	}
	return apiPermissionAllowed(permission, ids, menus)
}

// apiPermissionAllowed 缺少启用的权限节点时拒绝访问。
func apiPermissionAllowed(permission string, ids []int, menus []model.Menu) bool {
	for _, menu := range menus {
		if menu.ApiPermission == permission && menu.Type == model.MENU_TYPE_NODE && menu.Status == model.MENU_STATUS_ON {
			return slices.Contains(ids, menu.Id)
		}
	}
	return false
}

// SuperOnly 限制权限定义、授权关系和后台安全设置的写操作。
func SuperOnly(c *gin.Context) {
	admin, ok := help.GetLoginAdmin(c)
	if !ok || admin.IsSuper != model.ADMIN_IS_SUPER_Y {
		c.AbortWithStatusJSON(response.HTTP_NOT_ALLOWED_API, gin.H{"error": help.T(c, "No API Permission")})
		return
	}
	c.Next()
}
