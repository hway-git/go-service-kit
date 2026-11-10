package model

type IpWhitelist struct {
	ID       int    `gorm:"column:id;primaryKey;autoIncrement:true;comment:白名单记录ID" json:"id"`
	IP       string `gorm:"column:ip;type:varchar(45);uniqueIndex;comment:白名单IP地址" json:"ip"`
	Remark   string `gorm:"column:remark;comment:备注" json:"remark"`
	Operator string `gorm:"column:operator;comment:操作管理员账号" json:"operator"`
}

func (*IpWhitelist) TableName() string {
	return "ip_whitelist"
}
