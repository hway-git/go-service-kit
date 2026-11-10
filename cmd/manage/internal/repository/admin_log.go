package repository

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/utils/tool"
	"fmt"

	"gorm.io/gorm"
)

type AdminLog struct {
	BaseRepository[model.AdminLog]
}

// 获取管理员日志列表
func (c *AdminLog) GetList(params types.AdminLogListRequest) (List []types.AdminLogListData, Total int64, Err error) {
	var adminLogT model.AdminLog
	var myDb = mysql.DB
	myDb = myDb.Table(adminLogT.TableName())
	// 条件查询
	if params.Username != "" {
		myDb = myDb.Where("`operator`=?", params.Username)
	}
	if params.RequestMethod != "" {
		myDb = myDb.Where("`request_method` = ?", params.RequestMethod)
	}
	if params.Status != "" {
		myDb = myDb.Where("`status` = ?", params.Status)
	}
	if params.RequestUrl != "" {
		myDb = myDb.Where("`request_url` like ?", fmt.Sprintf("%%%s%%", params.RequestUrl))
	}
	if params.RequestIp != "" {
		myDb = myDb.Where("`request_ip` like ?", fmt.Sprintf("%%%s%%", params.RequestIp))
	}
	// total
	if err := myDb.Session(&gorm.Session{}).Count(&Total).Error; err != nil {
		Err = err
		return
	}
	// list
	offset, size := tool.GetPageInfo(params.Page, params.PageSize)
	Err = myDb.Order("`id` DESC").Offset(offset).Limit(size).Scan(&List).Error
	return
}
