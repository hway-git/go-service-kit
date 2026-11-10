package model

import (
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	Id           int    `gorm:"column:id;comment:管理员ID" json:"id"`                                         // 管理员ID
	Username     string `gorm:"column:username;type:varchar(64);uniqueIndex;comment:登录账号" json:"username"` // 账号
	Nickname     string `gorm:"column:nickname;comment:昵称" json:"nickname"`                                // 昵称
	Password     string `gorm:"column:password;comment:密码摘要" json:"-"`                                     // 密码
	Salt         string `gorm:"column:salt;comment:密码盐" json:"-"`                                          // 盐
	Phone        string `gorm:"column:phone;comment:联系电话" json:"phone"`                                    // 电话
	Email        string `gorm:"column:email;comment:邮箱" json:"email"`                                      // 邮箱
	Avatar       string `gorm:"column:avatar;comment:头像地址" json:"avatar"`                                  // 头像
	GoogleSecret string `gorm:"column:google_secret;comment:二次验证密钥" json:"-"`                              // 谷歌验证器secret
	Status       int    `gorm:"column:status;comment:状态，1正常0禁用" json:"status"`                             // 状态：1-禁用 | 0-正常
	Remark       string `gorm:"column:remark;comment:备注" json:"remark"`                                    // 备注
	RoleId       int    `gorm:"column:role_id;comment:角色ID" json:"role_id"`                                // 角色ID
	IsSuper      int    `gorm:"column:is_super;comment:超级管理员标记" json:"is_super"`                           // 是否超管
	NoDelete     int    `gorm:"column:no_delete;comment:删除保护标记" json:"no_delete"`                          // 是否不可删除
	DeletedTs    int64  `gorm:"column:deleted_ts;comment:删除时间戳" json:"deleted_ts"`                         // 删除时间
	CreatedTs    int64  `gorm:"column:created_ts;autoCreateTime;comment:创建时间戳" json:"created_ts"`
	UpdatedTs    int64  `gorm:"column:updated_ts;autoUpdateTime;comment:更新时间戳" json:"updated_ts"`

	// 内部字段
	Timezone string `gorm:"-" json:"timezone"` // 时区
}

func (*Admin) TableName() string {
	return "admin"
}

// ******************   常量定义    ********************
const (
	// 账户状态
	ADMIN_STATUS_ON     int = 1  // 正常
	ADMIN_STATUS_OFF    int = 0  // 禁用
	ADMIN_STATUS_DElETE int = -1 // 已删除

	// 是否超级管理员
	ADMIN_IS_SUPER_F int = 0 // 否
	ADMIN_IS_SUPER_Y int = 1 // 是

)

var (
	// 账户状态 MAP
	ADMIN_STATUS_MAP = map[int]string{
		ADMIN_STATUS_ON:     "正常",
		ADMIN_STATUS_OFF:    "禁用",
		ADMIN_STATUS_DElETE: "已删除",
	}
	// 是否超级管理员 MAP
	ADMIN_IS_SUPER_MAP = map[int]string{
		ADMIN_IS_SUPER_F: "否",
		ADMIN_IS_SUPER_Y: "是",
	}
)

// ******************   自定义方法    ****************

// 软删除 -- 调用： user := &AppUser{Id: 1}    err := user.SoftDelete(db)
func (r *Admin) SoftDelete(tx *gorm.DB) error {
	now := time.Now().Unix()
	return tx.Model(r).Updates(map[string]interface{}{
		"status":     ADMIN_STATUS_DElETE,
		"deleted_ts": now,
		"updated_ts": now,
	}).Error
}
