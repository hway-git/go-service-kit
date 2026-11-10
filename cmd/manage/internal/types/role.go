package types

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/shared/types/page"
)

// ####################   所有角色    ####################
type RoleAllSingle struct {
	Id   int    `json:"id"`   // 角色ID
	Name string `json:"name"` // 角色名称
}
type RoleAllResponse struct {
	List []RoleAllSingle `json:"list"`
}

// ####################    列表    ####################
type RoleListRequest struct {
	Name string `json:"name" form:"name"`
	page.PageRequest
}
type RoleListData struct { // 联表取数据实体
	model.Role
	GroupName string `gorm:"column:group_name"`
}
type RoleListRow struct {
	Id        int    `json:"id"`
	GroupId   int    `json:"group_id"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"created_at"`
	GroupName string `json:"group_name"`
	NoDelete  int    `json:"no_delete"`
}
type RoleListResponse struct {
	Rows []RoleListRow `json:"rows"`
	page.PageResponse
}

// RoleDetailResponse 描述角色详情的公开字段。
type RoleDetailResponse struct {
	Id        int    `json:"id"`         // 角色ID
	GroupId   int    `json:"group_id"`   // 所属分组ID
	GroupName string `json:"group_name"` // 分组名称
	Name      string `json:"name"`       // 角色名称
	Status    int    `json:"status"`     // 状态，0禁用1启用
	Remark    string `json:"remark"`     // 备注
	NoDelete  int    `json:"no_delete"`  // 是否禁止删除
	CreatedAt string `json:"created_at"` // 创建时间
	UpdatedAt string `json:"updated_at"` // 更新时间
}

// ####################    新增    ####################
type RoleAddRequest struct {
	GroupId int    `json:"group_id" binding:"required,gt=0"`
	Name    string `json:"name" binding:"required,max=64"`
	Status  int    `json:"status" binding:"oneof=0 1" enums:"0,1"`
	Remark  string `json:"remark" binding:"max=255"`
}

// ####################    编辑    ####################
type RoleEditRequest struct {
	Id int `json:"id" binding:"required,gt=0"`
	RoleAddRequest
}

// ####################    删除    ####################
type RoleDeleteRequest struct {
	Id int `form:"id" binding:"required,gt=0"`
}

// ####################    获取  角色权限   ####################
type RoleGetPermissionRequest struct {
	RoleId int `form:"role_id" binding:"required,gt=0"`
}
type RoleGetPermissionList struct {
	Id       int                     `json:"id"`
	Title    string                  `json:"title"`
	Checked  int                     `json:"checked"` // 是否选中，0未选1已选
	Type     int                     `json:"type"`
	Children []RoleGetPermissionList `json:"children"`
}
type RoleGetPermissionResponse struct {
	List                 []RoleGetPermissionList `json:"list"`
	CheckedMenuIds       []int                   `json:"checked_menu_ids"`
	CheckedPermissionIds []int                   `json:"checked_permission_ids"`
}

// ####################    设置  角色权限   ####################
type RoleSetPermissionRequest struct {
	RoleId        int    `json:"role_id" binding:"required,gt=0"`
	MenuIds       string `json:"menu_ids"`       // 用户展示菜单ID,多个用逗号分隔
	PermissionIds string `json:"permission_ids"` // 权限菜单ID,多个用逗号分隔
}
