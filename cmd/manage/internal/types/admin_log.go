package types

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/shared/types/page"
)

// ##################   管理员日志菜单   ###########################

type AdminLogListRequest struct {
	Username      string `form:"username"`
	RequestUrl    string `form:"request_url"`
	RequestMethod string `form:"request_method"`
	RequestIp     string `form:"request_ip"`
	Status        string `form:"status"`
	page.PageRequest
}
type AdminLogListData struct {
	model.AdminLog
	Username  string `gorm:"column:username" json:"username"`
	CreatedAt string `json:"created_at"`
	Test      string `json:"test,omitempty"`
}
type AdminLogListResponse struct {
	Rows []AdminLogListData `json:"rows"`
	page.PageResponse
}
