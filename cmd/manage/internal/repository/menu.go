package repository

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/db/mysql"

	"gorm.io/gorm"
)

type Menu struct {
	BaseRepository[model.Menu]
}

// 新增菜单 -- 返回ID
func (c *Menu) Add(menu model.Menu) (int, error) {
	if tx := mysql.DB.Create(&menu); tx.Error != nil {
		return 0, tx.Error
	}
	return menu.Id, nil
}

// 编辑删除
func (c *Menu) Delete(ids []int) error {
	// 开始事务
	return mysql.DB.Transaction(func(tx *gorm.DB) error {

		//删除菜单
		var menu model.Menu
		if err := tx.Where("id IN ?", ids).Delete(&menu).Error; err != nil {
			return err
		}

		//删除角色授权菜单ID
		roleMenu := model.RoleMenu{}
		if err := tx.Where("menu_id IN ?", ids).Delete(&roleMenu).Error; err != nil {
			return err
		}

		//删除角色授权菜单ID
		rolePermission := model.RolePermission{}
		if err := tx.Where("menu_id IN ?", ids).Delete(&rolePermission).Error; err != nil {
			return err
		}
		return nil
	})
}

// 获取全部菜单
func (c *Menu) ReadAll() ([]model.Menu, error) {
	rows := make([]model.Menu, 0)
	err := mysql.DB.Order("sort ASC, id ASC").Find(&rows).Error
	return rows, err
}

// 获取全部菜单
func (c *Menu) GetAll() []model.Menu {
	rows, _ := c.ReadAll()
	return rows
}

// 获取全部菜单
func (c *Menu) GetAllMenus() []model.Menu {
	var typeMenus []model.Menu
	menuList := c.GetAll()
	for _, v := range menuList {
		if v.Type == model.MENU_TYPE_MENU {
			typeMenus = append(typeMenus, v)
		}
	}
	return typeMenus
}

// 获取全部节点
func (c *Menu) GetAllNodes() []model.Menu {
	var typeNodes []model.Menu
	menuList := c.GetAll()
	for _, v := range menuList {
		if v.Type == model.MENU_TYPE_NODE {
			typeNodes = append(typeNodes, v)
		}
	}
	return typeNodes
}
