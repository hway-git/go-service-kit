package repository

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/utils/tool"
)

type IpWhitelist struct {
	BaseRepository[model.IpWhitelist]
}

// 查询 IP 白名单
func (c *IpWhitelist) List(params types.IpWhitelistListRequest) (list []model.IpWhitelist, total int64, Err error) {
	var m model.IpWhitelist

	var myDb = mysql.DB.Model(&m)

	// where
	if params.IP != "" {
		myDb = myDb.Where("`ip` = ?", params.IP)
	}

	// total
	if err := myDb.Count(&total).Error; err != nil {
		Err = err
		return
	}

	// list
	offset, size := tool.GetPageInfo(params.Page, params.PageSize)
	Err = myDb.Order("`id` DESC").Offset(offset).Limit(size).Scan(&list).Error
	return
}

// 批量删除IP
func (c *IpWhitelist) BatchDeleteByIds(ids []int) error {
	var m model.IpWhitelist
	if tx := mysql.DB.Where("`id` IN ?", ids).Delete(&m); tx.Error != nil {
		return tx.Error
	}
	return nil
}
