package repository

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/credential"
	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/utils/tool"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Admin struct {
	BaseRepository[model.Admin]
	DB *gorm.DB // 可注入数据库，未提供时使用应用数据库
}

func (c *Admin) db() *gorm.DB {
	if c.DB != nil {
		return c.DB
	}
	return mysql.DB
}

// GetById 查询管理员，不存在时返回错误。
func (c *Admin) GetById(id int64, row *model.Admin) error {
	if id <= 0 {
		return fmt.Errorf("管理员ID必须为正整数")
	}
	return c.db().Where("id = ?", id).Take(row).Error
}

// CreateWithRole 锁定有效角色后创建管理员，避免与角色删除并发。
func (c *Admin) CreateWithRole(row *model.Admin, actorIsSuper bool) error {
	if row.IsSuper != model.ADMIN_IS_SUPER_F && row.IsSuper != model.ADMIN_IS_SUPER_Y {
		return fmt.Errorf("超级管理员标记只支持0或1")
	}
	if row.IsSuper == model.ADMIN_IS_SUPER_Y && !actorIsSuper {
		return fmt.Errorf("仅超级管理员可管理超级管理员")
	}
	return c.db().Transaction(func(tx *gorm.DB) error {
		role, err := lockRole(tx, row.RoleId)
		if err != nil {
			return err
		}
		if role.Status != model.ROLE_STATUS_NORMAL {
			return fmt.Errorf("不能分配已禁用的角色")
		}
		return tx.Create(row).Error
	})
}

// UpdateWithRole 校验目标角色并更新管理员，允许保留原来的禁用角色。
func (c *Admin) UpdateWithRole(row *model.Admin, password string, actorIsSuper bool) error {
	if row.Id <= 0 {
		return fmt.Errorf("管理员ID必须为正整数")
	}
	return c.db().Transaction(func(tx *gorm.DB) error {
		role, err := lockRole(tx, row.RoleId)
		if err != nil {
			return err
		}
		var current model.Admin
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", row.Id).Take(&current).Error; err != nil {
			return err
		}
		if row.IsSuper != model.ADMIN_IS_SUPER_F && row.IsSuper != model.ADMIN_IS_SUPER_Y {
			return fmt.Errorf("超级管理员标记只支持0或1")
		}
		if !actorIsSuper && (current.IsSuper == model.ADMIN_IS_SUPER_Y || row.IsSuper == model.ADMIN_IS_SUPER_Y) {
			return fmt.Errorf("仅超级管理员可管理超级管理员")
		}
		if role.Status != model.ROLE_STATUS_NORMAL && current.RoleId != row.RoleId {
			return fmt.Errorf("不能分配已禁用的角色")
		}
		fields := map[string]any{
			"nickname": row.Nickname, "phone": row.Phone, "email": row.Email, "avatar": row.Avatar,
			"status": row.Status, "remark": row.Remark, "role_id": row.RoleId, "is_super": row.IsSuper,
		}
		if password != "" {
			digest, salt, err := credential.New(password)
			if err != nil {
				return err
			}
			fields["password"], fields["salt"] = digest, salt
		}
		return tx.Model(&model.Admin{}).Where("id = ?", row.Id).Updates(fields).Error
	})
}

// ChangePassword 在锁内校验当前密码，只更新密码字段。
func (c *Admin) ChangePassword(id int, previous, next string) error {
	return c.db().Transaction(func(tx *gorm.DB) error {
		var current model.Admin
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, id).Error; err != nil {
			return err
		}
		if !credential.Matches(previous, current.Salt, current.Password) {
			return fmt.Errorf("原密码输入错误")
		}
		digest, salt, err := credential.New(next)
		if err != nil {
			return err
		}
		return tx.Model(&current).Updates(map[string]any{"password": digest, "salt": salt}).Error
	})
}

// DeleteAuthorized 锁定目标后再次检查受保护的管理员身份。
func (c *Admin) DeleteAuthorized(id int64, actorIsSuper bool) error {
	return c.db().Transaction(func(tx *gorm.DB) error {
		var current model.Admin
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, id).Error; err != nil {
			return err
		}
		if current.NoDelete == 1 {
			return fmt.Errorf("该管理员不可删除")
		}
		if current.IsSuper == model.ADMIN_IS_SUPER_Y && !actorIsSuper {
			return fmt.Errorf("仅超级管理员可管理超级管理员")
		}
		return tx.Delete(&current).Error
	})
}

// 获取管理员 by username
func (c *Admin) GetByUsername(username string, m *model.Admin) error {
	if username != "" {
		if tx := mysql.DB.Where("`username` = ?", username).Find(m); tx.Error != nil {
			return tx.Error
		}
	}
	return nil
}

// 获取管理员列表
func (c *Admin) GetAdminList(params types.AdminListRequest) (List []types.AdminIndexData, Total int64, Err error) {
	var adminT model.Admin
	var roleT model.Role
	var myDb = mysql.DB
	myDb = myDb.Table(adminT.TableName() + " as a").
		Select("a.*, r.`name` as role_name").
		Joins("left join " + roleT.TableName() + " as r on a.`role_id`=r.`id`")
	if params.Name != "" {
		myDb = myDb.Where("a.`username` like ?", "%"+params.Name+"%")
	}

	// total
	if err := myDb.Count(&Total).Error; err != nil {
		Err = err
		return
	}
	// list
	offset, size := tool.GetPageInfo(params.Page, params.PageSize)
	Err = myDb.Offset(offset).Limit(size).Scan(&List).Error
	return
}
