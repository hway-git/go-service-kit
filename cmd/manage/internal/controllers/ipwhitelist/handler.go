package ipwhitelist

import (
	"example.com/go-service-kit/cmd/manage/internal/config"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/cmd/manage/internal/utils/validatorx"
	"net/netip"
	"strconv"
	"strings"

	"example.com/go-service-kit/shared/models"
	"example.com/go-service-kit/shared/types/page"

	"github.com/gin-gonic/gin"
)

var IpWhitelist = new(IpWhitelistController)

type IpWhitelistController struct{}

// @Summary 查询IP白名单
// @Tags IP白名单
// @Produce json
// @Security BearerAuth
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页条数" minimum(1)
// @Param ip query string false "IP地址关键词"
// @Success 200 {object} response.Response{data=types.IpWhitelistListResponse} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /ip/whitelist/list [get]
func (c *IpWhitelistController) List(ctx *gin.Context) {
	var req types.IpWhitelistListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, IpWhitelistListValidator)
		response.ApiError(ctx, msg, nil)
		return
	}

	// 列表
	repo := repository.IpWhitelist{}
	list, total, err := repo.List(req)
	if err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}

	var rows = []map[string]any{}
	for _, v := range list {
		row := map[string]interface{}{
			"id":       v.ID,
			"ip":       v.IP,
			"remark":   v.Remark,
			"operator": v.Operator,
		}
		rows = append(rows, row)
	}

	response.Success(ctx, page.ListResponse{
		List: rows,
		PageResponse: page.PageResponse{
			Page:      req.Page,
			PageSize:  req.PageSize,
			PageTotal: len(rows),
			Total:     total,
		},
	})
}

// @Summary 新增IP白名单
// @Tags IP白名单
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.IpWhitelistAddRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /ip/whitelist/add [post]
func (c *IpWhitelistController) Add(ctx *gin.Context) {

	auth, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	if err := (&repository.Admin{}).GetById(int64(auth.Id), &auth); err != nil {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}

	// 接受参数
	var req types.IpWhitelistAddRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, IpWhitelistAddValidator)
		response.ApiError(ctx, msg, nil)
		return
	}

	address, err := netip.ParseAddr(req.IP)
	if err != nil || address.Zone() != "" {
		response.ApiError(ctx, "IP 地址格式无效", nil)
		return
	}
	ip := address.Unmap().String()

	repo := repository.IpWhitelist{}
	var row model.IpWhitelist
	if err := repo.GetBySingleField(ip, "ip", &row); err != nil || row.ID > 0 {
		response.ApiError(ctx, "IP已经存在", nil)
		return
	}

	row = model.IpWhitelist{
		IP:       ip,
		Remark:   req.Remark,
		Operator: auth.Username,
	}
	if err := repo.Create(&row); err != nil {
		response.ApiError(ctx, "操作失败", nil)
		return
	}
	response.Success(ctx, nil)
}

// @Summary 删除IP白名单
// @Tags IP白名单
// @Produce json
// @Security BearerAuth
// @Param ids query string true "记录ID，多个使用逗号分隔"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /ip/whitelist/del [delete]
func (c *IpWhitelistController) Delete(ctx *gin.Context) {
	auth, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	if err := (&repository.Admin{}).GetById(int64(auth.Id), &auth); err != nil {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}

	// 接受参数
	var req types.IpWhitelistDeleteRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, IpWhitelistDeleteValidator)
		response.ApiError(ctx, msg, nil)
		return
	}

	ids := []int{}
	idArr := strings.Split(req.Ids, ",") // 把 req.Ids 按逗号切割
	for _, id := range idArr {
		if val, err := strconv.Atoi(id); err == nil {
			ids = append(ids, val)
		}
	}

	if len(ids) == 0 {
		response.Success(ctx, nil)
		return
	}
	repo := repository.IpWhitelist{}
	if err := repo.BatchDeleteByIds(ids); err != nil {
		response.ApiError(ctx, "操作失败", nil)
		return
	}
	response.Success(ctx, nil)
}

// 同步IP白名单
// @Summary 同步IP白名单缓存
// @Tags IP白名单
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /ip/whitelist/sync [get]
func (c *IpWhitelistController) Sync(ctx *gin.Context) {
	if err := config.Config.Reload(); err != nil {
		response.ApiError(ctx, "操作失败", nil)
		return
	}
	response.Success(ctx, nil)
}

// IP白名单开启状态
// @Summary 获取IP白名单开关
// @Tags IP白名单
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=types.IpWhitelistStatusResponse} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /ip/whitelist/status [get]
func (c *IpWhitelistController) GetStatus(ctx *gin.Context) {
	data := map[string]any{
		"status": config.Config.System().Admin_IP_White_Switch,
	}
	response.Success(ctx, data)
}

// 设置IP白名单状态
// @Summary 设置IP白名单开关
// @Description JSON正文包含status，0关闭，1开启。
// @Tags IP白名单
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.IpWhitelistStatusRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /ip/whitelist/status [put]
func (c *IpWhitelistController) SetStatus(ctx *gin.Context) {

	auth, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	if err := (&repository.Admin{}).GetById(int64(auth.Id), &auth); err != nil {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}

	// 接受参数
	var req types.IpWhitelistStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, IpWhitelistStatusValidator)
		response.ApiError(ctx, msg, nil)
		return
	}

	// 入库
	repo := repository.SysConfig{}
	if err := repo.SetSysConfigValue(models.SysConfCode_Admin_IP_White_Switch, strconv.Itoa(*req.Status)); err != nil {
		response.ApiError(ctx, "操作失败", nil)
		return
	}

	if err := config.Config.Reload(); err != nil {
		response.ApiError(ctx, "操作失败", nil)
		return
	}
	response.Success(ctx, nil)
}
