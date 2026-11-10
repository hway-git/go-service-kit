// bootstrap 包负责后台建表和初始化。
package bootstrap

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/credential"
	"example.com/go-service-kit/shared/models"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Initialize 建表并补齐初始配置，重复执行保留已有账户密码和授权。
func Initialize(db *gorm.DB, username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || len(username) > 64 {
		return fmt.Errorf("BOOTSTRAP_ADMIN_USER must contain 1 to 64 bytes")
	}
	if len(password) < 12 || len(password) > 1024 {
		return fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD 须为 12 至 1024 字节")
	}
	if db.Migrator().HasTable(&model.Menu{}) && db.Migrator().HasColumn(&model.Menu{}, "api_permission") {
		var duplicates []string
		if err := db.Model(&model.Menu{}).Where("api_permission <> ?", "").Group("api_permission").
			Having("COUNT(*) > 1").Limit(1).Pluck("api_permission", &duplicates).Error; err != nil {
			return fmt.Errorf("check permission code uniqueness: %w", err)
		}
		if len(duplicates) != 0 {
			return fmt.Errorf("duplicate permission code %q: review existing nodes and grants before retrying", duplicates[0])
		}
		if err := db.Model(&model.Menu{}).Where("api_permission = ?", "").UpdateColumn("api_permission", nil).Error; err != nil {
			return fmt.Errorf("normalize empty permission codes: %w", err)
		}
	}
	if err := db.AutoMigrate(
		&model.Admin{}, &model.AdminLog{}, &model.IpWhitelist{},
		&model.Menu{}, &model.Role{}, &model.RoleGroup{},
		&model.RoleMenu{}, &model.RolePermission{}, &models.SysConfig{},
	); err != nil {
		return fmt.Errorf("migrate management schema: %w", err)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		group := model.RoleGroup{Name: "系统管理", Status: model.ROLE_GROUP_STATUS_NORMAL, NoDelete: 1}
		if err := tx.Where("name = ?", group.Name).FirstOrCreate(&group).Error; err != nil {
			return err
		}
		role := model.Role{Name: "超级管理员", GroupId: group.Id, Status: model.ROLE_STATUS_NORMAL, NoDelete: 1}
		if err := tx.Where("name = ?", role.Name).FirstOrCreate(&role).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.Admin{}).Where("username = ?", username).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			digest, salt, err := credential.New(password)
			if err != nil {
				return err
			}
			admin := model.Admin{
				Username: username, Nickname: "管理员", Salt: salt,
				Password: digest,
				RoleId:   role.Id, IsSuper: model.ADMIN_IS_SUPER_Y,
				Status: model.ADMIN_STATUS_ON, NoDelete: 1,
			}
			if err := tx.Create(&admin).Error; err != nil {
				return err
			}
		}
		for _, code := range []string{
			models.SysConfCode_Admin_IP_White_Switch,
			models.SysConfCode_Admin_Pwd_Code_Switch,
			models.SysConfCode_Admin_Single_Login_Switch,
		} {
			setting := models.SysConfig{Code: code, Title: code, Value: "0", Type: "bool"}
			if err := tx.Where("code = ?", code).FirstOrCreate(&setting).Error; err != nil {
				return err
			}
		}
		return initializeManagementPermissions(tx)
	})
}

// initializeManagementPermissions 补齐后台管理权限节点，保留已有配置与授权。
func initializeManagementPermissions(tx *gorm.DB) error {
	for _, endpoint := range types.ManagementEndpoints {
		node := model.Menu{Title: endpoint.Title, Type: model.MENU_TYPE_NODE, ApiPath: endpoint.Path, ApiMethod: endpoint.Method,
			ApiPermission: model.PermissionCode(endpoint.Permission()), Status: model.MENU_STATUS_ON, NoDelete: 1}
		if err := tx.Where("api_permission = ?", node.ApiPermission).FirstOrCreate(&node).Error; err != nil {
			return err
		}
		if node.Type != model.MENU_TYPE_NODE || node.ApiPath != endpoint.Path || node.ApiMethod != endpoint.Method {
			return fmt.Errorf("permission %q conflicts with registered route %s %s", node.ApiPermission, endpoint.Method, endpoint.Path)
		}
		if node.NoDelete != 1 {
			if err := tx.Model(&node).UpdateColumn("no_delete", 1).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
