package model

import "database/sql/driver"

// PermissionCode stores a menu's empty permission as SQL NULL while keeping its JSON string.
type PermissionCode string

func (code PermissionCode) Value() (driver.Value, error) {
	if code == "" {
		return nil, nil
	}
	return string(code), nil
}

type Menu struct {
	Id                 int            `gorm:"primary_key;column:id;comment:菜单或节点ID" json:"id"`
	Title              string         `gorm:"column:title;comment:名称" json:"title"`             // 名称
	TitleEn            string         `gorm:"column:title_en;comment:英文名称" json:"title_en"`     // 英文名称
	Pid                int            `gorm:"column:pid;index;comment:上级菜单ID" json:"pid"`       // 上级
	Icon               string         `gorm:"column:icon;comment:图标" json:"icon"`               // 图标
	Type               int            `gorm:"column:type;comment:类型，1菜单2权限节点" json:"type"`      // 类型：1-菜单; 2-节点
	ApiPath            string         `gorm:"column:api_path;comment:接口路由" json:"api_path"`     // 接口：路由地址
	ApiMethod          string         `gorm:"column:api_method;comment:接口方法" json:"api_method"` // 接口：请求方式：get|post|put|delete
	ApiPermission      PermissionCode `gorm:"column:api_permission;type:varchar(191);uniqueIndex:ux_rbac_menu_api_permission;comment:接口权限标识" json:"api_permission"`
	WebRouterPath      string         `gorm:"column:web_router_path;comment:前端路由地址" json:"web_router_path"`           // 前端：路由地址
	WebRouterName      string         `gorm:"column:web_router_name;comment:前端路由名称" json:"web_router_name"`           // 前端：路由名称
	WebRouterRedirect  string         `gorm:"column:web_router_redirect;comment:前端重定向地址" json:"web_router_redirect"`  // 前端：路由重定向地址
	WebRouterComponent string         `gorm:"column:web_router_component;comment:前端组件路径" json:"web_router_component"` // 前端：路由组件路径
	Sort               int            `gorm:"column:sort;comment:排序" json:"sort"`                                     // 排序
	Status             int            `gorm:"column:status;comment:状态，1正常2禁用" json:"status"`                          // 是否正常: 0-禁用 | 1-启用
	Remark             string         `gorm:"column:remark;comment:备注" json:"remark"`                                 // 描述
	ActiveMenu         string         `gorm:"column:active_menu;comment:高亮菜单" json:"activeMenu"`                      // 是否外链
	IsLink             string         `gorm:"column:is_link;comment:外链地址" json:"is_link"`                             // 是否外链
	IsHide             int            `gorm:"column:is_hide;comment:隐藏标记" json:"is_hide"`                             // 是否隐藏菜单
	IsFull             int            `gorm:"column:is_full;comment:全屏标记" json:"is_full"`                             // 是否全屏打开
	IsAffix            int            `gorm:"column:is_affix;comment:固定标记" json:"is_affix"`                           // 是否固定在某个位置
	IsKeepAlive        int            `gorm:"column:is_keep_alive;comment:保留页面状态标记" json:"is_keep_alive"`             // 是否保持存活
	CreatedTs          int64          `gorm:"column:created_ts;autoCreateTime;comment:创建时间戳" json:"created_ts"`       //
	UpdatedTs          int64          `gorm:"column:updated_ts;autoUpdateTime;comment:更新时间戳" json:"updated_ts"`       //
	NoDelete           int            `gorm:"column:no_delete;comment:删除保护标记" json:"no_delete"`                       // 是否不可删除
}

func (*Menu) TableName() string {
	return "rbac_menu"
}

// ******************   常量定义    ********************
const (
	// 是否菜单
	MENU_TYPE_MENU int = 1 // 菜单
	MENU_TYPE_NODE int = 2 // 节点
	// 菜单状态
	MENU_STATUS_ON  int = 1 // 正常
	MENU_STATUS_OFF int = 2 // 禁用
)

var (
	// 是否菜单 MAP
	MENU_TYPE_MAP = map[int]string{
		MENU_TYPE_MENU: "菜单",
		MENU_TYPE_NODE: "节点",
	}
	// 菜单状态  MAP
	MENU_STATUS_MAP = map[int]string{
		MENU_STATUS_ON:  "正常",
		MENU_STATUS_OFF: "禁用",
	}
)
