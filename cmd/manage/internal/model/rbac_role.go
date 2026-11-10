package model

type Role struct {
	Id        int    `gorm:"column:id;primaryKey;autoIncrement;comment:角色ID" json:"id"`                  // 角色ID
	GroupId   int    `gorm:"column:group_id;not null;index;comment:所属分组ID" json:"group_id"`              // 所属分组
	Name      string `gorm:"column:name;type:varchar(64);not null;uniqueIndex;comment:角色名称" json:"name"` // 角色名称
	Status    int    `gorm:"column:status;not null;comment:状态，0禁用1启用" json:"status"`                     // 状态
	Remark    string `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`          // 备注
	CreatedTs int64  `gorm:"column:created_ts;autoCreateTime;comment:创建时间戳" json:"created_ts"`           // 创建时间
	UpdatedTs int64  `gorm:"column:updated_ts;autoUpdateTime;comment:更新时间戳" json:"updated_ts"`           // 更新时间
	NoDelete  int    `gorm:"column:no_delete;not null;default:0;comment:是否禁止删除" json:"no_delete"`        // 删除保护
}

func (*Role) TableName() string {
	return "rbac_role"
}

// ******************   常量定义    ****************

const (
	// 角色状态
	ROLE_STATUS_NORMAL = 1 // 正常
	ROLE_STATUS_OFF    = 0 // 禁用

	// 分组状态
	ROLE_GROUP_STATUS_NORMAL = 1 // 正常
	ROLE_GROUP_STATUS_OFF    = 0 // 禁用
)

var (
	// 角色状态 MAP
	ROLE_STATUS_MAP = map[int]string{
		ROLE_STATUS_NORMAL: "正常",
		ROLE_STATUS_OFF:    "禁用",
	}

	// 角色状态 MAP
	ROLE_GROUP_STATUS_MAP = map[int]string{
		ROLE_GROUP_STATUS_NORMAL: "正常",
		ROLE_GROUP_STATUS_OFF:    "禁用",
	}
)
