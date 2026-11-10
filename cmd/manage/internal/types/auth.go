package types

// #########################   登录   ####################
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ##################    登录接口返回体   ########################
type AdminInfo struct {
	AdminId  int    `json:"admin_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	RoleId   int    `json:"role_id"`
	IsSuper  int    `json:"is_super"`
}
type AdminResponse struct {
	Token string    `json:"token"`
	Info  AdminInfo `json:"info"`
}

// #########################   修改密码   ####################
type ResetPwdRequest struct {
	OriPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"new_password_confirmation" binding:"required"`
}

// ####################    管理员验证     ##########
type VerifyRequest struct {
	AuthPwd string `json:"auth_pwd"  binding:"required"` // 管理员密码
}
