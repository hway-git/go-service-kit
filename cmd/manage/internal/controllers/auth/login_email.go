package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"example.com/go-service-kit/cmd/manage/internal/cache"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/shared/services/email"
	"example.com/go-service-kit/utils/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	emailChallengeTTL = 5 * time.Minute
	emailSendCooldown = time.Minute
	emailMaxAttempts  = 5
)

func (c *AuthController) beginEmailVerification(ctx *gin.Context, admin *model.Admin) {
	address := strings.TrimSpace(admin.Email)
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Address != address {
		response.ApiError(ctx, "账号未配置有效邮箱，请联系管理员", nil)
		return
	}
	if c.mail == nil {
		response.ServerError(ctx, "邮件服务未配置，请联系管理员")
		return
	}

	requestCtx := ctx.Request.Context()
	allowed, err := cache.ReserveEmailLoginSend(requestCtx, admin.Id, emailSendCooldown)
	if err != nil {
		response.ServerError(ctx)
		return
	}
	if !allowed {
		response.ApiError(ctx, "验证码发送过于频繁，请稍后重试", nil)
		return
	}

	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		response.ServerError(ctx)
		return
	}
	challengeID := hex.EncodeToString(id[:])
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		response.ServerError(ctx)
		return
	}
	code := fmt.Sprintf("%06d", value.Int64())
	challenge := cache.EmailLoginChallenge{
		AdminID: admin.Id,
		Version: c.emailAdminVersion(admin),
		Digest:  c.emailDigest("code:" + challengeID + ":" + code),
	}
	if err := cache.SaveEmailLoginChallenge(requestCtx, challengeID, challenge, emailChallengeTTL); err != nil {
		response.ServerError(ctx)
		return
	}
	if err := c.mail.Send(requestCtx, &email.Message{
		To:          parsed.String(),
		Subject:     "后台登录验证码",
		ContentType: "text/plain",
		Body:        fmt.Sprintf("您的后台登录验证码为 %s，%d 分钟内有效。", code, int(emailChallengeTTL/time.Minute)),
	}); err != nil {
		_ = cache.DeleteEmailLoginChallenge(requestCtx, challengeID)
		logger.Warn("admin login verification email failed", zap.Int("admin_id", admin.Id), zap.Error(err))
		response.ServerError(ctx, "验证码发送失败，请稍后重试")
		return
	}
	response.Success(ctx, types.LoginChallengeResponse{
		ChallengeID: challengeID,
		MaskedEmail: maskLoginEmail(address),
		ExpiresIn:   int(emailChallengeTTL.Seconds()),
	})
}

// VerifyEmail 校验一次性邮箱验证码，成功后签发本地 JWT。
// @Summary 验证管理员登录邮箱验证码
// @Tags 管理员认证
// @Accept json
// @Produce json
// @Param body body types.LoginVerifyEmailRequest true "验证参数"
// @Success 200 {object} response.Response{data=types.AdminResponse} "登录成功"
// @Router /auth/login/verify-email [post]
func (c *AuthController) VerifyEmail(ctx *gin.Context) {
	var req types.LoginVerifyEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ApiError(ctx, "无效的邮箱验证码参数", nil)
		return
	}

	challenge, err := cache.ConsumeEmailLoginChallenge(ctx.Request.Context(), req.ChallengeID,
		c.emailDigest("code:"+req.ChallengeID+":"+req.Code), emailMaxAttempts)
	if err != nil {
		response.ServerError(ctx)
		return
	}
	if challenge == nil {
		response.ApiError(ctx, "验证码无效或已过期", nil)
		return
	}

	var admin model.Admin
	if err := (&repository.Admin{}).GetById(int64(challenge.AdminID), &admin); err != nil ||
		admin.Status != model.ADMIN_STATUS_ON || c.emailAdminVersion(&admin) != challenge.Version {
		response.ApiError(ctx, "账号信息已变更，请重新登录", nil)
		return
	}
	if loginRoleAllowed(ctx, &admin) {
		c.completeLogin(ctx, &admin)
	}
}

func (c *AuthController) emailAdminVersion(admin *model.Admin) string {
	return c.emailDigest(fmt.Sprintf("account:%d:%q:%q:%q:%d:%d:%d:%d", admin.Id, strings.TrimSpace(admin.Email),
		admin.Salt, admin.Password, admin.Status, admin.RoleId, admin.IsSuper, admin.UpdatedTs))
}

func (c *AuthController) emailDigest(value string) string {
	mac := hmac.New(sha256.New, c.codeKey)
	_, _ = mac.Write([]byte("admin-login-email:" + value))
	return hex.EncodeToString(mac.Sum(nil))
}

func maskLoginEmail(address string) string {
	separator := strings.LastIndexByte(address, '@')
	local := []rune(address[:separator])
	masked := "***"
	if len(local) > 1 {
		masked = string(local[0]) + masked
	}
	if len(local) >= 4 {
		masked += string(local[len(local)-1])
	}
	return masked + address[separator:]
}
