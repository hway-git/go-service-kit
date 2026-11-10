package types

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/shared/types/page"
)

// ##################   接口结构体   ###########################

// ##################   管理员菜单   ###########################

type AdminMenuMeta struct {
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	IsLink      string `json:"isLink"`
	IsHide      bool   `json:"isHide"`
	IsFull      bool   `json:"isFull"`
	IsAffix     bool   `json:"isAffix"`
	IsKeepAlive bool   `json:"isKeepAlive"`
}
type AdminMenu struct {
	Path      string        `json:"path"`                // 前端组件路由
	Name      string        `json:"name"`                // 前端组件名称
	Redirect  string        `json:"redirect,omitempty"`  // 前端组件重定向
	Component string        `json:"component,omitempty"` // 前端组件名称
	Meta      AdminMenuMeta `json:"meta"`
	Children  []AdminMenu   `json:"children,omitempty"`
}

// ##################   管理员列表   ###########################
// ##################   管理员列表   ###########################
type AdminListRequest struct {
	Name string `json:"name" form:"name"`
	page.PageRequest
}
type AdminIndexData struct { // 取数据实体
	model.Admin
	RoleName string `json:"role_name"`
}
type AdminIndexRow struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
	RoleId    int    `json:"role_id"`
	IsSuper   int    `json:"is_super"`
	CreatedAt string `json:"created_at"`
	RoleName  string `json:"role_name"`
	NoDelete  int    `json:"no_delete"`
}
type AdminIndexResponse struct {
	Rows []AdminIndexRow `json:"rows"`
	page.PageResponse
}

// ##################   管理员 新增  ###########################
type AdminAddRequest struct {
	Username string `json:"username"     binding:"required"`
	Nickname string `json:"nickname"`
	Password string `json:"password"     binding:"required"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Remark   string `json:"remark"`
	RoleId   int    `json:"role_id"     binding:"required"`
	Status   int    `json:"status"`
	IsSuper  int    `json:"is_super"`
}

// ##################   管理员 编辑  ###########################
type AdminEditRequest struct {
	Id       int64  `json:"id"           binding:"required"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Remark   string `json:"remark"`
	RoleId   int    `json:"role_id"      binding:"required"`
	Status   int    `json:"status"`
	IsSuper  int    `json:"is_super"`
}

// ##################   管理员 踢下线（清空 refresh 会话）   ###########################
type AdminKickLoginRequest struct {
	AdminId int64 `json:"admin_id" binding:"required"`
}

// ##################   管理员 删除  ###########################
type AdminDeleteRequest struct {
	Id int64 `form:"id" binding:"required"`
}

// ##################   管理员 全部  ###########################
type AdminAllResponse struct {
	List []map[string]any `json:"list" `
}
