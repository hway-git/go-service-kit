package model

type RolePermission struct {
	RoleId int `gorm:"column:role_id;primaryKey;autoIncrement:false;comment:角色ID" json:"role_id"`         // 角色ID
	MenuId int `gorm:"column:menu_id;primaryKey;autoIncrement:false;index;comment:权限节点ID" json:"menu_id"` // 权限节点ID
}

func (*RolePermission) TableName() string {
	return "rbac_role_permission"
}
