package auth

const (
	Username_Required = "账号不能为空"
	Password_Required = "密码不能为空"

	OriPassword_Required     = "原密码不能为空"
	NewPassword_Required     = "新密码不能为空"
	ConfirmPassword_Required = "确认密码不能为空"

	AuthPwd_Required = "授权码不能为空"
)

var (
	// 登录验证器
	LoginValidator = map[string]string{
		"Username.required": Username_Required,
		"Password.required": Password_Required,
	}

	// 重置密码验证器
	ResetPwdValidator = map[string]string{
		"OriPassword.required":     OriPassword_Required,
		"NewPassword.required":     Password_Required,
		"ConfirmPassword.required": ConfirmPassword_Required,
	}

	// 登出验证器
	LogoutValidator = map[string]string{}

	// 管理员密码验证
	VerifyValidator = map[string]string{
		"AuthPwd.required": AuthPwd_Required,
	}
)
