package model

import (
	"time"
)

type AdminLog struct {
	Id            int           `gorm:"column:id;primary_key;comment:审计记录ID" json:"id"`                         // id
	RequestUrl    string        `gorm:"column:request_url;comment:请求路由" json:"request_url"`                     // 接口URL
	RequestMethod string        `gorm:"column:request_method;comment:请求方法" json:"request_method"`               // 请求方式
	RequestIp     string        `gorm:"column:request_ip;comment:客户端IP" json:"request_ip"`                      // 请求IP
	RequestData   string        `gorm:"column:request_data;type:text;comment:脱敏请求数据" json:"request_data"`       // request数据
	ResponseData  string        `gorm:"column:response_data;type:text;comment:脱敏响应数据" json:"response_data"`     // response数据
	Status        int           `gorm:"column:status;comment:响应状态码" json:"status"`                              // 状态码
	Duration      time.Duration `gorm:"column:duration;comment:耗时毫秒数" json:"duration"`                          // 耗时 ms
	Operator      string        `gorm:"column:operator;type:varchar(64);index;comment:操作管理员账号" json:"operator"` // 管理员名称
	CreatedTs     int64         `gorm:"column:created_ts;index;autoCreateTime;comment:创建时间戳" json:"created_ts"` // 添加时间
}

func (*AdminLog) TableName() string {
	return "admin_log"
}
