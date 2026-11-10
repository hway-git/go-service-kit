package models

type SysConfig struct {
	Id     int    `gorm:"column:id;primary_key;comment:系统配置ID" json:"id"`
	Title  string `gorm:"column:title;comment:配置显示名称" json:"title"`
	Code   string `gorm:"column:code;type:varchar(64);uniqueIndex;comment:配置标识码，确认后禁止修改" json:"code"`
	Value  string `gorm:"column:value;comment:配置值" json:"value"`
	Type   string `gorm:"column:type;comment:配置值类型" json:"type"`
	Desc   string `gorm:"column:desc;comment:配置描述" json:"desc"`
	Sort   int    `gorm:"column:sort;comment:显示排序" json:"sort"`
	Extend string `gorm:"column:extend;comment:扩展属性" json:"extend"`
}

func (*SysConfig) TableName() string {
	return "sys_config"
}

// =============   标识码   ============

const (
	SysConfCode_Admin_IP_White_Switch     = "IP_White_Switch"     // IP 白名单开关
	SysConfCode_Admin_Pwd_Code_Switch     = "Pwd_Code_Switch"     // 后台管理员授权码开关
	SysConfCode_Admin_Single_Login_Switch = "Single_Login_Switch" // 后台管理员授权码开关
)

// 具体值
type SysConfData struct {

	// ========  后台 ============
	Admin_IP_White_Switch     int // IP 白名单开关   0-关 1-开
	Admin_Pwd_Code_Switch     int // 后台管理员授权码开关 0-关 1-开
	Admin_Single_Login_Switch int // 后台管理员授权码开关 0-关 1-开
}
