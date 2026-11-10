package jwt

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"example.com/go-service-kit/cmd/manage/internal/cache"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	jwtSecret  []byte
	ExpireTime = 7 * 24 * time.Hour
	// ExpireTime = 30 * time.Second
)

// RefreshGrace 是已过期令牌仍允许续期的最大窗口。
const RefreshGrace = 2 * time.Hour

// Configure 在注册路由前设置令牌签名密钥。
func Configure(secret string) error {
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET or jwt_secret must contain at least 32 bytes")
	}
	jwtSecret = []byte(secret)
	return nil
}

type Claims struct {
	AdminId    int    `json:"i"`
	Username   string `json:"u"`
	Status     int    `json:"s"`
	RoleId     int    `json:"r"`
	IsSuper    int    `json:"super"`
	Timezone   string `json:"tz"`
	Credential string `json:"credential"`
	jwt.StandardClaims
}

// CreateUserToken
func CreateUserToken(admin *model.Admin, jti string) (string, error) {
	if len(jwtSecret) < 32 {
		return "", fmt.Errorf("JWT signing key is not configured")
	}
	claims := &Claims{
		AdminId:    admin.Id,
		Username:   admin.Username,
		Status:     admin.Status,
		RoleId:     admin.RoleId,
		IsSuper:    admin.IsSuper,
		Timezone:   admin.Timezone,
		Credential: credentialFingerprint(admin),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ExpireTime).Unix(),
			IssuedAt:  time.Now().Unix(),
			Id:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 解析token
func ParseToken(tokenStr string) (*jwt.Token, *Claims, error) {
	if len(jwtSecret) < 32 {
		return nil, nil, fmt.Errorf("JWT signing key is not configured")
	}
	if !strings.HasPrefix(tokenStr, "Bearer ") || len(tokenStr) <= len("Bearer ") {
		return nil, nil, fmt.Errorf("token 为空")
	}
	tokenStr = tokenStr[len("Bearer "):] // 去掉 "Bearer "
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(token *jwt.Token) (i interface{}, err error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected JWT signing method")
			}
			return jwtSecret, nil
		})
	return token, claims, err
}

// credentialFingerprint 使用独立用途前缀派生指纹，不暴露数据库密码摘要。
func credentialFingerprint(admin *model.Admin) string {
	mac := hmac.New(sha256.New, jwtSecret)
	_, _ = fmt.Fprintf(mac, "admin-credential:%d:%s:%s", admin.Id, admin.Salt, admin.Password)
	return hex.EncodeToString(mac.Sum(nil))
}

// MatchesCredential 让任一改密入口都能使旧令牌失效，旧版无指纹令牌须重新登录。
func MatchesCredential(claims *Claims, admin *model.Admin) bool {
	return claims.Credential != "" && hmac.Equal([]byte(claims.Credential), []byte(credentialFingerprint(admin)))
}

// InvalidateToken 撤销指定JWT，保留黑名单至续期窗口结束。
func InvalidateToken(ctx context.Context, tokenStr string) error {
	token, claims, err := ParseToken(tokenStr)
	if token == nil || claims == nil {
		return fmt.Errorf("令牌无效")
	}
	if err != nil {
		if validation, ok := err.(*jwt.ValidationError); !ok || validation.Errors != jwt.ValidationErrorExpired {
			return fmt.Errorf("令牌无效")
		}
	}
	ttl := time.Until(time.Unix(claims.ExpiresAt, 0).Add(RefreshGrace))
	if ttl <= 0 {
		return nil
	}
	return cache.AddInvalidAdminToken(ctx, claims.StandardClaims.Id, ttl)
}

func GetAccessExpireDuration() time.Duration {
	return ExpireTime
}
