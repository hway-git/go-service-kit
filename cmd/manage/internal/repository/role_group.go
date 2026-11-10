package repository

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/utils/tool"
	"fmt"

	"gorm.io/gorm"
)

type RoleGroup struct {
	BaseRepository[model.RoleGroup]
}

// 获取角色 By Name
func (c *RoleGroup) GetByName(name string) (model model.Role) {
	if name != "" {
		mysql.DB.Where("name = ?", name).Find(&model)
	}
	return
}

// 获取角色 By Name
func (c *RoleGroup) GetByNameNoSelf(name string, id int) (model model.Role) {
	if name != "" {
		mysql.DB.Where("name = ? and id <> ?", name, id).Find(&model)
	}
	return
}

// 编辑删除
func (c *RoleGroup) Delete(ids []int) error {
	// 开始事务
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		// 删除角色
		var role model.RoleGroup
		if err := tx.Where("id IN ?", ids).Delete(&role).Error; err != nil {
			return err
		}

		// //删除角色菜单 , 包含子菜单
		// var roleMenu model.RoleMenu
		// if err := tx.Where("role_id IN ?", ids).Delete(&roleMenu).Error; err != nil {
		// 	return err
		// }

		return nil
	})
}

// 获取角色所有分组
func (c *RoleGroup) RoleAllGroupAll() (groups []model.RoleGroup) {
	var roleGroup model.RoleGroup
	var sql string = fmt.Sprintf("select * from %s where `status`=?", roleGroup.TableName())
	mysql.DB.Raw(sql, model.ROLE_GROUP_STATUS_NORMAL).Find(&groups)
	return
}

// 获取角色列表
func (c *RoleGroup) RoleGroupList(params types.RoleGroupListRequest) (list []types.RoleGroupData, total int64, Err error) {
	var roleGroupT model.RoleGroup
	var roleT model.Role
	myDb := mysql.DB.Table(roleGroupT.TableName() + " as rg").
		Select([]string{
			"rg.*",
			"count(r.`id`) as roleCount",
		}).
		Joins("left join " + roleT.TableName() + " as r on r.`group_id`=rg.`id`").
		Group("rg.id")
	if params.Name != "" {
		myDb = myDb.Where("rg.`name` like ?", "%"+params.Name+"%")
	}

	// total
	if err := myDb.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		Err = err
		return
	}
	// list
	offset, size := tool.GetPageInfo(params.Page, params.PageSize)
	myDb.Offset(offset).Limit(size).Scan(&list)
	return
}
