package common

import (
	"example.com/go-service-kit/cmd/manage/internal/config"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/utils/enum"
	"example.com/go-service-kit/utils/tool"

	"github.com/gin-gonic/gin"
)

var Common = new(CommonController)

type CommonController struct{}

// 重新刷新语言包，不需要重启整个程序
// @Summary 获取后台初始化信息
// @Description 返回当前管理员菜单、权限代码、时区选项和系统开关。
// @Tags 公共接口
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=types.InitResponse} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /common/init [get]
func (c *CommonController) Init(ctx *gin.Context) {

	auth, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}

	var curr_timezone string = help.GetTimezone(ctx)

	var initTimezone []map[string]string

	// 管理员可选择的时区。
	initTimezone = []map[string]string{
		{"key": enum.TIMEZONE_GMT, "val": help.T(ctx, enum.TIMEZONE_MAP[enum.TIMEZONE_GMT])},
		{"key": enum.TIMEZONE_BRAZIL, "val": help.T(ctx, enum.TIMEZONE_MAP[enum.TIMEZONE_BRAZIL])},
		{"key": enum.TIMEZONE_CHINA, "val": help.T(ctx, enum.TIMEZONE_MAP[enum.TIMEZONE_CHINA])},
		{"key": enum.TIMEZONE_VIETNAM, "val": help.T(ctx, enum.TIMEZONE_MAP[enum.TIMEZONE_VIETNAM])},
		{"key": enum.TIMEZONE_USA, "val": help.T(ctx, enum.TIMEZONE_MAP[enum.TIMEZONE_USA])},
	}

	init := types.InitResponse{

		// 管理员菜单权限
		AdminMenu: getAdminMenu(ctx, auth),
		AdminNode: getAdminNode(ctx, auth),

		// 授权码开关
		PwdSwitch: config.Config.System().Admin_Pwd_Code_Switch,

		// 时区
		CurrTime:        tool.GetCurrTime(curr_timezone), // 当前时间
		CurrentTimezone: curr_timezone,                   // 默认时区
		InitTimezone:    initTimezone,                    // 时区选项

		// IsSuperAdmin: false, // 是否是超管
	}

	// 是超管
	if auth, ok := help.GetLoginAdmin(ctx); ok {
		init.IsSuperAdmin = auth.IsSuper == model.ADMIN_IS_SUPER_Y
	}

	response.Success(ctx, init)
}
