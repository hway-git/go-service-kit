package model

type RoleGroup struct {
	Id        int    `gorm:"primary_key;comment:角色分组ID" json:"id"`                             // 类型ID
	Name      string `gorm:"column:name;comment:分组名称" json:"name"`                             // 类型名称
	Status    int    `gorm:"column:status;comment:状态，1正常0禁用" json:"status"`                    // 状态：0-禁用 | 1-正常
	Remark    string `gorm:"column:remark;comment:备注" json:"remark"`                           // 备注
	CreatedTs int64  `gorm:"column:created_ts;autoCreateTime;comment:创建时间戳" json:"created_ts"` // 创建时间
	UpdatedTs int64  `gorm:"column:updated_ts;autoUpdateTime;comment:更新时间戳" json:"updated_ts"` // 更新时间
	NoDelete  int    `gorm:"column:no_delete;comment:删除保护标记" json:"no_delete"`                 // 是否不可删除
}

func (*RoleGroup) TableName() string {
	return "rbac_role_group"
}
