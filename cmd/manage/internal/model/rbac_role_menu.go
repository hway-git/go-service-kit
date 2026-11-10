package model

type RoleMenu struct {
	RoleId int `gorm:"column:role_id;primaryKey;autoIncrement:false;comment:角色ID" json:"role_id"`       // 角色ID
	MenuId int `gorm:"column:menu_id;primaryKey;autoIncrement:false;index;comment:菜单ID" json:"menu_id"` // 菜单ID
}

func (*RoleMenu) TableName() string {
	return "rbac_role_menu"
}
