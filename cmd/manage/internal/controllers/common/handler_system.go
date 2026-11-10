package common

import (
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/utils/tool"

	"github.com/gin-gonic/gin"
)

var System = new(SystemController)

type SystemController struct{}

// 系统时区
// @Summary 获取指定时区的当前时间
// @Description 读取Timezone请求头，只返回当前时间，不保存用户时区设置。
// @Tags 公共接口
// @Produce json
// @Param Timezone header string false "时区标识，未指定时使用系统默认值"
// @Success 200 {object} response.Response{data=types.TimezoneSetResponse} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 907 {object} response.LegacyError "访问限制"
// @Router /common/tz-set [post]
func (c *SystemController) TimezoneSet(ctx *gin.Context) {
	tz := help.GetTimezone(ctx)
	response.Success(ctx, types.TimezoneSetResponse{
		Time: tool.GetCurrTime(tz), // 当前时间
	})
}

// 系统时区
// func (c *SystemController) Timezone(ctx *gin.Context) {
// 	var tz string = help.GetTimezone(ctx)
// 	var defaultTz string = enum.TIMEZONE_UTC_MAP[tz]
// 	// 可选时区
// 	initTimezone := []map[string]string{
// 		{"key": "W3", "val": "UTC -3"},
// 		{"key": "E8", "val": "UTC +8"},
// 	}
// 	response.Success(ctx, types.TimezoneResponse{
// 		CurrTime:        tool.GetCurrTime(tz), // 当前时间
// 		DefaultTimezone: defaultTz,            // 默认时区
// 		InitTimezone:    initTimezone,         // 时区选项
// 	})
// }
