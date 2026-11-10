package types

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/shared/types/page"
)

// ####################    角色分组   ####################
type RoleAllGroupResponse struct {
	List []map[string]any `json:"list" `
}

// ####################    分组列表    ####################
type RoleGroupListRequest struct {
	Name string `json:"name" form:"name"`
	page.PageRequest
}

// 联表取数据实体
type RoleGroupData struct {
	model.RoleGroup
	RoleCount string `gorm:"column:roleCount" json:"roleCount"`
}
type RoleGroupRow struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
	RoleCount string `json:"role_count"`
	NoDelete  int    `json:"no_delete"`
}
type RoleGroupListResponse struct {
	Rows []RoleGroupRow `json:"rows"`
	page.PageResponse
}

// ####################    新增    ####################
type RoleGroupAddRequest struct {
	Name   string `json:"name" binding:"required"`
	Status int    `json:"status"`
	Remark string `json:"remark"`
}

// ####################    编辑    ####################
type RoleGroupEditRequest struct {
	Id int `json:"id" binding:"required"`
	RoleGroupAddRequest
}

// ####################    删除    ####################
type RoleGroupDeleteRequest struct {
	Id int `form:"id" binding:"required"`
}
