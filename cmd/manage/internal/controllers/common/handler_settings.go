package common

import (
	"strconv"

	"example.com/go-service-kit/cmd/manage/internal/config"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/shared/models"
	"example.com/go-service-kit/utils/logger"

	"github.com/gin-gonic/gin"
)

// SettingsList 返回后台配置开关及其当前值。
func (c *CommonController) SettingsList(ctx *gin.Context) {
	var repo repository.SysConfig
	var rows []models.SysConfig
	if err := repo.GetAll(&rows); err != nil {
		logger.FromContext(ctx.Request.Context()).Error("读取后台设置失败")
		response.ServerError(ctx)
		return
	}
	response.Success(ctx, rows)
}

// SettingsEdit 只接受后台已登记的开关，修改后刷新当前进程缓存。
func (c *CommonController) SettingsEdit(ctx *gin.Context) {
	var req struct {
		Code  string `json:"code" binding:"required"`
		Value *int   `json:"value" binding:"required,oneof=0 1"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ApiError(ctx, "参数错误", nil)
		return
	}
	if err := (&repository.SysConfig{}).SetSysConfigValue(req.Code, strconv.Itoa(*req.Value)); err != nil {
		response.ApiError(ctx, "操作失败", nil)
		return
	}
	if err := config.Config.Reload(); err != nil {
		logger.FromContext(ctx.Request.Context()).Error("刷新后台设置失败")
		response.ServerError(ctx)
		return
	}
	response.Success(ctx, nil)
}
