package types

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
)

// 填充其他字段
type MenuList struct {
	model.Menu
	CreatedTs    int64      `json:"-"`
	UpdatedTs    int64      `json:"-"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
	TypeRemark   string     `json:"type_remark"`
	StatusRemark string     `json:"status_remark"`
	Children     []MenuList `json:"children,omitempty"`
}

// ####################    列表    ####################
type MenuListResPonse struct {
	List []MenuList `json:"list"`
}

// ####################    新增    ####################
type MenuAddRequest struct {
	Title   string `json:"title" binding:"required"`
	TitleEn string `json:"title_en"`
	Pid     int    `json:"pid"`
	Icon    string `json:"icon"`
	Type    int    `json:"type" binding:"required"`
	Sort    int    `json:"sort"`
	Status  int    `json:"status"`
	Remark  string `json:"remark"`
	// 接口
	ApiPath       string `json:"api_path"`       // 接口：路由地址
	ApiMethod     string `json:"api_method"`     // 接口：请求方式：get|post|put|delete
	ApiPermission string `json:"api_permission"` // 兼容旧客户端，实际权限码由服务端生成。
	// 前端
	WebRouterComponent string `json:"web_router_component"`
	WebRouterPath      string `json:"web_router_path"`
	WebRouterName      string `json:"web_router_name"`
	WebRouterRedirect  string `json:"web_router_redirect"`

	IsLink      string `json:"is_link"`       // 是否外链
	IsHide      int    `json:"is_hide"`       // 是否隐藏菜单
	IsAffix     int    `json:"is_affix"`      // 是否固定在某个位置
	IsFull      int    `json:"is_full"`       // 是否全屏打开
	IsKeepAlive int    `json:"is_keep_alive"` // 是否保持存活
}

// ####################    编辑    ####################
type MenuEditRequest struct {
	Id int `json:"id" binding:"required"`
	MenuAddRequest
}

// ####################    删除    ####################
type MenuDeleteRequest struct {
	Id int `json:"id" binding:"required"`
}
