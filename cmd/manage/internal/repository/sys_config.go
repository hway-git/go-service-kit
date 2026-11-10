package repository

import (
	"fmt"
	"strconv"

	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/shared/models"
)

type SysConfig struct {
	BaseRepository[models.SysConfig]
}

// LoadSysConfData 读取系统开关，失败时由调用方保留原缓存。
func (c *SysConfig) LoadSysConfData() (models.SysConfData, error) {
	var data models.SysConfData
	var list []models.SysConfig
	if err := c.GetAll(&list); err != nil {
		return data, err
	}
	for _, row := range list {
		var target *int
		switch row.Code {
		case models.SysConfCode_Admin_IP_White_Switch:
			target = &data.Admin_IP_White_Switch
		case models.SysConfCode_Admin_Pwd_Code_Switch:
			target = &data.Admin_Pwd_Code_Switch
		case models.SysConfCode_Admin_Single_Login_Switch:
			target = &data.Admin_Single_Login_Switch
		default:
			continue
		}
		value, err := strconv.Atoi(row.Value)
		if err != nil || (value != 0 && value != 1) {
			return data, fmt.Errorf("系统开关 %s 只支持 0 或 1", row.Code)
		}
		*target = value
	}
	return data, nil
}

// SetSysConfigValue 校验并更新已有后台开关。
func (c *SysConfig) SetSysConfigValue(code, value string) error {
	switch code {
	case models.SysConfCode_Admin_IP_White_Switch, models.SysConfCode_Admin_Pwd_Code_Switch, models.SysConfCode_Admin_Single_Login_Switch:
	default:
		return fmt.Errorf("不支持的后台开关")
	}
	if value != "0" && value != "1" {
		return fmt.Errorf("系统开关只支持 0 或 1")
	}
	var row models.SysConfig
	if err := mysql.DB.Where("code = ?", code).Take(&row).Error; err != nil {
		return err
	}
	return mysql.DB.Model(&row).Update("value", value).Error
}
