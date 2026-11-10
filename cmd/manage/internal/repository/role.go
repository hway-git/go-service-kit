package repository

import (
	"errors"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/db/mysql"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	driver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Role 负责角色资料及授权的原子维护。
type Role struct {
	BaseRepository[model.Role]
	DB *gorm.DB // 可注入数据库，未提供时使用应用数据库
}

func (c *Role) db() *gorm.DB {
	if c.DB != nil {
		return c.DB
	}
	return mysql.DB
}

// GetById 查询目标角色，查不到记录时返回错误。
func (c *Role) GetById(id int64, row *model.Role) error {
	if id <= 0 {
		return fmt.Errorf("角色ID必须为正整数")
	}
	return c.db().Where("id = ?", id).Take(row).Error
}

func validateRole(row *model.Role) error {
	row.Name = strings.TrimSpace(row.Name)
	if !utf8.ValidString(row.Name) || utf8.RuneCountInString(row.Name) < 1 || utf8.RuneCountInString(row.Name) > 64 {
		return fmt.Errorf("角色名称必须为1至64个字符")
	}
	if !utf8.ValidString(row.Remark) || utf8.RuneCountInString(row.Remark) > 255 {
		return fmt.Errorf("备注不能超过255个字符")
	}
	if row.GroupId <= 0 {
		return fmt.Errorf("分组ID必须为正整数")
	}
	if row.Status != model.ROLE_STATUS_NORMAL && row.Status != model.ROLE_STATUS_OFF {
		return fmt.Errorf("角色状态无效")
	}
	return nil
}

func roleWriteError(err error) error {
	var duplicate *driver.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		return fmt.Errorf("存在相同名称的角色")
	}
	return err
}

func requireRoleGroup(tx *gorm.DB, id int) error {
	var group model.RoleGroup
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", id).Take(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色分组不存在")
		}
		return err
	}
	return nil
}

func lockRole(tx *gorm.DB, id int) (*model.Role, error) {
	if id <= 0 {
		return nil, fmt.Errorf("角色ID必须为正整数")
	}
	var row model.Role
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("角色不存在")
	}
	return &row, err
}

// Create 校验分组并创建角色，名称唯一性由数据库保证。
func (c *Role) Create(row *model.Role) error {
	if err := validateRole(row); err != nil {
		return err
	}
	if row.Id != 0 || row.NoDelete != 0 {
		return fmt.Errorf("不能指定角色ID或删除保护")
	}
	err := c.db().Transaction(func(tx *gorm.DB) error {
		if err := requireRoleGroup(tx, row.GroupId); err != nil {
			return err
		}
		return tx.Create(row).Error
	})
	if err != nil {
		row.Id = 0
	}
	return roleWriteError(err)
}

// Update 仅修改允许维护的字段，不使用可能新增记录的保存操作。
func (c *Role) Update(row *model.Role) error {
	if row.Id <= 0 {
		return fmt.Errorf("角色ID必须为正整数")
	}
	if err := validateRole(row); err != nil {
		return err
	}
	err := c.db().Transaction(func(tx *gorm.DB) error {
		if _, err := lockRole(tx, row.Id); err != nil {
			return err
		}
		if err := requireRoleGroup(tx, row.GroupId); err != nil {
			return err
		}
		return tx.Model(&model.Role{}).Where("id = ?", row.Id).Updates(map[string]any{
			"group_id": row.GroupId, "name": row.Name, "status": row.Status, "remark": row.Remark,
		}).Error
	})
	return roleWriteError(err)
}

func normalizedRoleIDs(ids []int) ([]int, error) {
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, fmt.Errorf("菜单或权限ID必须为正整数")
		}
		out = append(out, id)
	}
	slices.Sort(out)
	return slices.Compact(out), nil
}

// Delete 在同一事务中验证引用、删除角色并清理全部授权。
func (c *Role) Delete(ids []int) error {
	ids, err := normalizedRoleIDs(ids)
	if err != nil || len(ids) == 0 {
		return fmt.Errorf("角色ID必须为正整数")
	}
	return c.db().Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			row, err := lockRole(tx, id)
			if err != nil {
				return err
			}
			if row.NoDelete != 0 {
				return fmt.Errorf("该角色不可删除")
			}
		}
		var count int64
		if err := tx.Model(&model.Admin{}).Where("role_id IN ? AND status <> ?", ids, model.ADMIN_STATUS_DElETE).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("角色仍被管理员使用，请先调整管理员角色")
		}
		if err := tx.Where("role_id IN ?", ids).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id IN ?", ids).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Where("id IN ?", ids).Delete(&model.Role{}).Error
	})
}

// RoleAll 返回启用角色，读取失败时由调用方处理。
func (c *Role) RoleAll() ([]model.Role, error) {
	rows := make([]model.Role, 0)
	err := c.db().Where("status = ?", model.ROLE_STATUS_NORMAL).Order("id ASC").Find(&rows).Error
	return rows, err
}

func (c *Role) roleQuery() *gorm.DB {
	return c.db().Table("rbac_role AS r").Select("r.*, rg.name AS group_name").
		Joins("LEFT JOIN rbac_role_group AS rg ON r.group_id = rg.id")
}

// RoleDetail 返回角色与所属分组的详情。
func (c *Role) RoleDetail(id int) (*types.RoleListData, error) {
	if id <= 0 {
		return nil, fmt.Errorf("角色ID必须为正整数")
	}
	var row types.RoleListData
	err := c.roleQuery().Where("r.id = ?", id).Take(&row).Error
	return &row, err
}

// RoleList 查询分页角色并传递数据库错误。
func (c *Role) RoleList(params types.RoleListRequest) ([]types.RoleListData, int64, error) {
	if params.Page <= 0 || params.PageSize <= 0 || params.PageSize > 100 || params.Page > int(^uint(0)>>1)/params.PageSize {
		return nil, 0, fmt.Errorf("分页参数无效，每页最多100条")
	}
	query := c.roleQuery()
	if params.Name != "" {
		query = query.Where("r.name LIKE ?", "%"+params.Name+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]types.RoleListData, 0)
	err := query.Order("r.id DESC").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Scan(&rows).Error
	return rows, total, err
}

// GetMenuIds 读取当前页面授权，不保存进程内缓存。
func (c *Role) GetMenuIds(roleID int) ([]int, error) {
	if roleID <= 0 {
		return nil, fmt.Errorf("角色ID必须为正整数")
	}
	ids := make([]int, 0)
	err := c.db().Model(&model.RoleMenu{}).Where("role_id = ?", roleID).Order("menu_id ASC").Pluck("menu_id", &ids).Error
	return ids, err
}

// GetPermission 读取当前接口授权，不保存进程内缓存。
func (c *Role) GetPermission(roleID int) ([]int, error) {
	if roleID <= 0 {
		return nil, fmt.Errorf("角色ID必须为正整数")
	}
	ids := make([]int, 0)
	err := c.db().Model(&model.RolePermission{}).Where("role_id = ?", roleID).Order("menu_id ASC").Pluck("menu_id", &ids).Error
	return ids, err
}

func validateRoleAssignments(tx *gorm.DB, menuIDs, permissionIDs []int) error {
	ids, _ := normalizedRoleIDs(append(slices.Clone(menuIDs), permissionIDs...))
	if len(ids) == 0 {
		return nil
	}
	var rows []model.Menu
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id IN ?", ids).Order("id ASC").Find(&rows).Error; err != nil {
		return err
	}
	nodes := make(map[int]model.Menu, len(rows))
	for _, row := range rows {
		nodes[row.Id] = row
	}
	for _, selection := range []struct {
		ids  []int
		kind int
	}{{menuIDs, model.MENU_TYPE_MENU}, {permissionIDs, model.MENU_TYPE_NODE}} {
		for _, id := range selection.ids {
			row, exists := nodes[id]
			if !exists || row.Status != model.MENU_STATUS_ON || row.Type != selection.kind {
				return fmt.Errorf("菜单或权限节点不存在、已禁用或类型不符")
			}
		}
	}
	return nil
}

// SetPermission 先验证全部节点，再在事务中完整替换授权。
func (c *Role) SetPermission(roleID int, menuIDs, permissionIDs []int) error {
	if roleID <= 0 {
		return fmt.Errorf("角色ID必须为正整数")
	}
	var err error
	menuIDs, err = normalizedRoleIDs(menuIDs)
	if err != nil {
		return err
	}
	permissionIDs, err = normalizedRoleIDs(permissionIDs)
	if err != nil {
		return err
	}
	return c.db().Transaction(func(tx *gorm.DB) error {
		if _, err := lockRole(tx, roleID); err != nil {
			return err
		}
		if err := validateRoleAssignments(tx, menuIDs, permissionIDs); err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		if len(menuIDs) > 0 {
			rows := make([]model.RoleMenu, 0, len(menuIDs))
			for _, id := range menuIDs {
				rows = append(rows, model.RoleMenu{RoleId: roleID, MenuId: id})
			}
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
		if len(permissionIDs) > 0 {
			rows := make([]model.RolePermission, 0, len(permissionIDs))
			for _, id := range permissionIDs {
				rows = append(rows, model.RolePermission{RoleId: roleID, MenuId: id})
			}
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
