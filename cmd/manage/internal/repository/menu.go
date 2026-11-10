package repository

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/db/mysql"

	driver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Menu struct {
	BaseRepository[model.Menu]
}

// 新增菜单 -- 返回ID
func (c *Menu) Add(menu model.Menu) (int, error) {
	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		rows, err := lockMenus(tx)
		if err != nil {
			return err
		}
		if err := validateMenuParent(rows, &menu); err != nil {
			return err
		}
		if err := setMenuPermission(&menu); err != nil {
			return err
		}
		return tx.Create(&menu).Error
	})
	var duplicate *driver.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		return 0, fmt.Errorf("权限码已存在")
	}
	return menu.Id, err
}

func lockMenus(tx *gorm.DB) (map[int]model.Menu, error) {
	var menus []model.Menu
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("id ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	rows := make(map[int]model.Menu, len(menus))
	for _, menu := range menus {
		rows[menu.Id] = menu
	}
	return rows, nil
}

func validateMenuParent(rows map[int]model.Menu, menu *model.Menu) error {
	visited := map[int]bool{menu.Id: true}
	for parentID := menu.Pid; parentID != 0; {
		if visited[parentID] {
			return fmt.Errorf("菜单不能挂到自身或形成循环")
		}
		parent, exists := rows[parentID]
		if !exists || parent.Type != model.MENU_TYPE_MENU {
			return fmt.Errorf("上级菜单不存在或类型无效")
		}
		visited[parentID] = true
		parentID = parent.Pid
	}
	return nil
}

func setMenuPermission(menu *model.Menu) error {
	switch menu.Type {
	case model.MENU_TYPE_MENU:
		menu.ApiPermission = ""
	case model.MENU_TYPE_NODE:
		if !strings.HasPrefix(menu.ApiPath, "/") || strings.ContainsAny(menu.ApiPath, "?#") ||
			strings.ContainsFunc(menu.ApiPath, unicode.IsSpace) {
			return fmt.Errorf("请输入本地接口路由路径")
		}
		endpoint := types.BusinessEndpoint{Path: menu.ApiPath, Method: menu.ApiMethod}
		code := endpoint.Permission()
		if len([]rune(code)) > 191 {
			return fmt.Errorf("权限码超过 191 字符")
		}
		menu.ApiPermission = model.PermissionCode(code)
	default:
		return fmt.Errorf("菜单类型无效")
	}
	return nil
}

// Update 校验最新菜单树并保护已授权节点的路由身份。
func (c *Menu) Update(menu *model.Menu) error {
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		rows, err := lockMenus(tx)
		if err != nil {
			return err
		}
		current, exists := rows[menu.Id]
		if !exists {
			return fmt.Errorf("菜单不存在")
		}
		if menu.Type != current.Type {
			return fmt.Errorf("不能修改菜单类型")
		}
		if err := validateMenuParent(rows, menu); err != nil {
			return err
		}
		if err := setMenuPermission(menu); err != nil {
			return err
		}
		if menu.Type == model.MENU_TYPE_NODE && (menu.ApiPermission != current.ApiPermission ||
			menu.ApiPath != current.ApiPath || menu.ApiMethod != current.ApiMethod) {
			return fmt.Errorf("不能修改权限节点的路由或权限码")
		}
		return tx.Model(&current).Select("*").Omit("id", "created_ts", "no_delete").Updates(menu).Error
	})
}

func (c *Menu) SetStatus(id, status int) error {
	if status != model.MENU_STATUS_ON && status != model.MENU_STATUS_OFF {
		return fmt.Errorf("菜单状态无效")
	}
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		var menu model.Menu
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&menu, id).Error; err != nil {
			return err
		}
		return tx.Model(&menu).Update("status", status).Error
	})
}

func (c *Menu) Delete(id int) error {
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		rows, err := lockMenus(tx)
		if err != nil {
			return err
		}
		if _, exists := rows[id]; !exists {
			return fmt.Errorf("菜单不存在")
		}
		children := make(map[int][]int)
		for _, row := range rows {
			children[row.Pid] = append(children[row.Pid], row.Id)
		}
		ids := []int{id}
		visited := map[int]bool{id: true}
		for index := 0; index < len(ids); index++ {
			if rows[ids[index]].NoDelete != 0 {
				return fmt.Errorf("菜单或子节点不可删除")
			}
			for _, child := range children[ids[index]] {
				if !visited[child] {
					visited[child] = true
					ids = append(ids, child)
				}
			}
		}

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
