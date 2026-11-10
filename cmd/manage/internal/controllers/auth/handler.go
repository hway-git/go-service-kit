package auth

import (
	"strings"

	"example.com/go-service-kit/cmd/manage/internal/cache"
	"example.com/go-service-kit/cmd/manage/internal/config"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/credential"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/cmd/manage/internal/utils/jwt"
	"example.com/go-service-kit/cmd/manage/internal/utils/validatorx"
	"example.com/go-service-kit/shared/services/email"
	"example.com/go-service-kit/utils/tool"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	mail    *email.Client
	codeKey []byte
}

func NewController(mail *email.Client, secret string) *AuthController {
	return &AuthController{mail: mail, codeKey: []byte(secret)}
}

// 登录
// @Summary 管理员登录
// @Description 账号配置邮箱时返回 challenge_id、masked_email 和 expires_in，验证邮箱后签发 JWT。
// @Tags 管理员认证
// @Accept json
// @Produce json
// @Param body body types.LoginRequest true "请求参数"
// @Success 200 {object} response.Response{data=types.AdminResponse} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 907 {object} response.LegacyError "访问限制"
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req types.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, LoginValidator)
		response.ApiError(ctx, msg, nil)
		return
	}

	var admin model.Admin
	if err := (&repository.Admin{}).GetByUsername(req.Username, &admin); err != nil || admin.Id == 0 {
		response.ApiError(ctx, "账号不存在", nil)
		return
	}

	// 判断用户状态
	if admin.Status != model.ADMIN_STATUS_ON {
		response.ApiError(ctx, "账号被禁用", nil)
		return
	}

	// 登录密码验证
	if ok := credential.Matches(req.Password, admin.Salt, admin.Password); !ok {
		response.ApiError(ctx, "密码错误", nil)
		return
	}

	if !loginRoleAllowed(ctx, &admin) {
		return
	}
	if strings.TrimSpace(admin.Email) != "" {
		c.beginEmailVerification(ctx, &admin)
		return
	}
	c.completeLogin(ctx, &admin)
}

func loginRoleAllowed(ctx *gin.Context, admin *model.Admin) bool {
	if admin.IsSuper != model.ADMIN_IS_SUPER_Y {
		var role model.Role
		if err := (&repository.Role{}).GetById(int64(admin.RoleId), &role); err != nil || role.Status != model.ROLE_STATUS_NORMAL {
			response.ApiError(ctx, "角色不存在或已禁用", nil)
			return false
		}
	}
	return true
}

func (c *AuthController) completeLogin(ctx *gin.Context, admin *model.Admin) {
	admin.Timezone = help.GetTimezone(ctx)

	jti := tool.GetUUID()
	token, err := jwt.CreateUserToken(admin, jti)
	if err != nil {
		response.ApiError(ctx, "登录失败", nil)
		return
	}

	// 开启单点登录，则记录最后一次登录的 token
	if config.Config.System().Admin_Single_Login_Switch == 1 {
		if err := cache.SetAdminLastJwtToken(ctx.Request.Context(), int64(admin.Id), jti, jwt.GetAccessExpireDuration()+jwt.RefreshGrace); err != nil {
			response.ApiError(ctx, "登录失败", nil)
			return
		}
	}

	data := types.AdminResponse{
		Token: token,
		Info: types.AdminInfo{
			AdminId:  admin.Id,
			Username: admin.Username,
			Nickname: admin.Nickname,
			Phone:    admin.Phone,
			Email:    admin.Email,
			Avatar:   admin.Avatar,
			RoleId:   admin.RoleId,
			IsSuper:  admin.IsSuper,
		},
	}
	response.Success(ctx, data)
}

// 修改密码
// @Summary 修改当前管理员密码
// @Description 校验原密码和两次新密码，成功后撤销当前管理员的访问令牌。
// @Tags 管理员认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.ResetPwdRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /auth/reset-pwd [post]
func (c *AuthController) ResetPwd(ctx *gin.Context) {

	authUser, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	// 接收参数
	var req types.ResetPwdRequest
	if err := ctx.ShouldBind(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, ResetPwdValidator)
		response.ApiError(ctx, msg, nil)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.ApiError(ctx, "新密码输入不一致", nil)
		return
	}

	// 指纹随密码变化自动失效，无需在MySQL提交后再补写撤销状态。
	if err := (&repository.Admin{}).ChangePassword(authUser.Id, req.OriPassword, req.NewPassword); err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}
	response.Success(ctx, nil)
}

// 退出登录
// @Summary 管理员退出登录
// @Tags 管理员认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	_, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取用户信息失败", nil)
		return
	}
	if err := jwt.InvalidateToken(ctx.Request.Context(), ctx.GetHeader("Authorization")); err != nil {
		response.ApiError(ctx, "操作失败", nil)
		return
	}
	response.Success(ctx, nil)
}

// 管理员验证密码
// @Summary 验证当前管理员密码
// @Tags 管理员认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.VerifyRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /auth/verify [post]
func (c *AuthController) Verify(ctx *gin.Context) {

	auth, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	if err := (&repository.Admin{}).GetById(int64(auth.Id), &auth); err != nil {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}

	// 接收参数
	var req types.VerifyRequest
	if err := ctx.ShouldBind(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, VerifyValidator)
		response.ApiError(ctx, msg, nil)
		return
	}

	// 校验管理员密码
	if !credential.Matches(req.AuthPwd, auth.Salt, auth.Password) {
		response.ApiError(ctx, "授权码错误", nil)
		return
	}

	response.Success(ctx, nil)
}
