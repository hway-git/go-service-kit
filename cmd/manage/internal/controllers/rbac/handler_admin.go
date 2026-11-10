package rbac

import (
	"example.com/go-service-kit/cmd/manage/internal/cache"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/credential"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/cmd/manage/internal/utils/jwt"
	"example.com/go-service-kit/cmd/manage/internal/utils/validatorx"
	"example.com/go-service-kit/shared/types/page"
	"example.com/go-service-kit/utils/tool"
	"time"

	"github.com/gin-gonic/gin"
)

var Admin = new(AdminController)

type AdminController struct{}

// 管理员列表
// @Summary 查询管理员
// @Tags 管理员管理
// @Produce json
// @Security BearerAuth
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页条数" minimum(1)
// @Param name query string false "管理员昵称或用户名"
// @Success 200 {object} response.Response{data=types.AdminIndexResponse} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /admin/list [get]
func (c *AdminController) List(ctx *gin.Context) {
	// 时区
	adminTz := help.GetTimezone(ctx)
	// 接收参数
	var req types.AdminListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, AdminListValidator)
		response.ApiError(ctx, msg, nil)
		return
	}
	// 取数据
	list, total, err := (&repository.Admin{}).GetAdminList(req)
	if err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}
	//返回列表
	var adminIndexRows []types.AdminIndexRow
	for _, v := range list {
		row := types.AdminIndexRow{
			Id:       v.Id,
			Username: v.Username,
			Nickname: v.Nickname,
			Phone:    v.Phone,
			Email:    v.Email,
			Avatar:   v.Avatar,
			Status:   v.Status,
			Remark:   v.Remark,
			RoleId:   v.RoleId,
			IsSuper:  v.IsSuper,
			RoleName: v.RoleName,
			// 转换时区时间
			CreatedAt: tool.GetTimestampToTimeZone(v.CreatedTs, adminTz),
			NoDelete:  v.NoDelete,
		}
		adminIndexRows = append(adminIndexRows, row)
	}

	data := types.AdminIndexResponse{
		Rows: adminIndexRows,
		PageResponse: page.PageResponse{
			Page:      req.Page,
			PageSize:  req.PageSize,
			PageTotal: len(adminIndexRows),
			Total:     total,
		},
	}
	response.Success(ctx, data)
}

// 新增管理员
// @Summary 新增管理员
// @Tags 管理员管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.AdminAddRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /admin [post]
func (c *AdminController) Add(ctx *gin.Context) {
	actor, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	// 接收参数
	var req types.AdminAddRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, AdminAddValidator)
		response.ApiError(ctx, msg, nil)
		return
	}
	// 实例化DB
	adminDB := repository.Admin{}
	//存在相同名称的管理员
	var existAdmin model.Admin
	if err := adminDB.GetByUsername(req.Username, &existAdmin); err == nil && existAdmin.Id > 0 {
		response.ApiError(ctx, "存在相同名称的管理员", nil)
		return
	}
	// 组装数据
	password, salt, err := credential.New(req.Password)
	if err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}
	row := model.Admin{
		Username: req.Username,
		Nickname: req.Nickname,
		Password: password,
		Salt:     salt,
		Phone:    req.Phone,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Status:   req.Status,
		Remark:   req.Remark,
		RoleId:   req.RoleId,
		IsSuper:  req.IsSuper,
	}
	// 入库
	if err := adminDB.CreateWithRole(&row, actor.IsSuper == model.ADMIN_IS_SUPER_Y); err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}
	response.Success(ctx, nil)
}

// 编辑管理员
// @Summary 修改管理员
// @Tags 管理员管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.AdminEditRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /admin [put]
func (c *AdminController) Edit(ctx *gin.Context) {
	actor, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	// 接收参数
	var params types.AdminEditRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, AdminEditValidator)
		response.ApiError(ctx, msg, nil)
		return
	}
	// 实例化DB
	adminDB := repository.Admin{}

	admin := model.Admin{
		Id: int(params.Id), Nickname: params.Nickname, Phone: params.Phone,
		Email: params.Email, Avatar: params.Avatar, Status: params.Status,
		Remark: params.Remark, RoleId: params.RoleId, IsSuper: params.IsSuper,
	}
	// 入库
	if err := adminDB.UpdateWithRole(&admin, params.Password, actor.IsSuper == model.ADMIN_IS_SUPER_Y); err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}

	response.Success(ctx, nil)
}

// KickLogin 撤销目标管理员的访问令牌。
// @Summary 撤销管理员访问令牌
// @Tags 管理员管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.AdminKickLoginRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /admin/kick [post]
func (c *AdminController) KickLogin(ctx *gin.Context) {
	actor, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	var req types.AdminKickLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, AdminKickLoginValidator)
		response.ApiError(ctx, msg, nil)
		return
	}
	var target model.Admin
	if err := (&repository.Admin{}).GetById(req.AdminId, &target); err != nil || target.Id == 0 {
		response.ApiError(ctx, "管理员不存在", nil)
		return
	}

	if target.IsSuper == model.ADMIN_IS_SUPER_Y && actor.IsSuper != model.ADMIN_IS_SUPER_Y {
		response.ApiError(ctx, "仅超级管理员可管理超级管理员", nil)
		return
	}
	if err := cache.SetAdminTokenRevokedAt(ctx.Request.Context(), req.AdminId, time.Now().Unix(), jwt.GetAccessExpireDuration()+jwt.RefreshGrace); err != nil {
		response.ApiError(ctx, "操作失败", nil)
		return
	}
	response.Success(ctx, nil)
}

// 删除管理员
// @Summary 删除管理员
// @Tags 管理员管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "记录ID" minimum(1)
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /admin [delete]
func (c *AdminController) Delete(ctx *gin.Context) {
	actor, ok := help.GetLoginAdmin(ctx)
	if !ok {
		response.ApiError(ctx, "获取登录信息失败", nil)
		return
	}
	// 接收参数
	var req types.AdminDeleteRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, AdminDeleteValidator)
		response.ApiError(ctx, msg, nil)
		return
	}
	if err := (&repository.Admin{}).DeleteAuthorized(req.Id, actor.IsSuper == model.ADMIN_IS_SUPER_Y); err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}
	response.Success(ctx, nil)
}

// 所有管理员
// @Summary 获取启用的管理员选项
// @Tags 管理员管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=types.AdminAllResponse} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /admin/all [get]
func (c *AdminController) All(ctx *gin.Context) {
	var admins []model.Admin
	if err := (&repository.Admin{}).GetByCondition(&admins, "`status` = ?", model.ADMIN_STATUS_ON); err != nil {
		response.ApiError(ctx, err.Error(), nil)
	}
	var reponse = types.AdminAllResponse{}
	if len(admins) > 0 {
		var list []map[string]any
		for _, v := range admins {
			list = append(list, map[string]any{"key": v.Id, "val": v.Username})
		}
		reponse.List = list
	}
	response.Success(ctx, reponse)
}

// 管理员日志列表
// @Summary 查询管理员操作日志
// @Tags 管理员管理
// @Produce json
// @Security BearerAuth
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页条数" minimum(1)
// @Param username query string false "操作管理员用户名"
// @Param command query int false "操作类型"
// @Param request_url query string false "请求地址"
// @Param request_method query string false "请求方法"
// @Param request_ip query string false "请求IP"
// @Param status query string false "响应状态"
// @Success 200 {object} response.Response{data=types.AdminLogListResponse} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /admin/log [get]
func (c *AdminController) AdminLogList(ctx *gin.Context) {
	var params types.AdminLogListRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, AdminLogListValidator)
		response.ApiError(ctx, msg, nil)
		return
	}
	list, total, err := (&repository.AdminLog{}).GetList(params)
	if err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}
	adminTz := help.GetTimezone(ctx) // 时区
	rows := []types.AdminLogListData{}
	for _, v := range list {
		v.CreatedAt = tool.GetTimestampToTimeZone(v.CreatedTs, adminTz) // 转换时间
		rows = append(rows, v)
	}
	response.Success(ctx, types.AdminLogListResponse{
		Rows: rows,
		PageResponse: page.PageResponse{
			Page:      params.Page,
			PageSize:  params.PageSize,
			PageTotal: len(rows),
			Total:     total,
		},
	})
}

// =================================

// // 管理员的菜单
// func (c *AdminController) AdminMenu(ctx *gin.Context) {
// 	auth, ok := help.GetLoginAdmin(ctx)
// 	if !ok {
// 		response.ApiError(ctx, "获取登录信息失败", nil)
// 		return
// 	}
// 	var normalMenus []model.Menu                      // 有效菜单
// 	var allMenus = (&repository.Menu{}).GetAllMenus() // 全部菜单
// 	for _, v := range allMenus {
// 		if v.Status == model.MENU_STATUS_ON {
// 			normalMenus = append(normalMenus, v)
// 		}
// 	}
// 	// 替换语言
// 	var lang string = help.GetLanguage(ctx) // 语言
// 	for k, v := range normalMenus {
// 		switch lang {
// 		case "en":
// 			normalMenus[k].Title = v.TitleEn
// 		}
// 	}
// 	// 不是管理员，需要根据权限过滤
// 	if auth.IsSuper != model.ADMIN_IS_SUPER_Y {
// 		roleDB := repository.Role{}
// 		RoleMenuIds := tool.UniqueArray(roleDB.GetMenuIds(auth.RoleId)) // 管理员拥有的菜单ID （包含了节点）
// 		TmpMenu := []model.Menu{}
// 		for _, v := range normalMenus {
// 			if ok := tool.InArray(v.Id, RoleMenuIds); ok {
// 				TmpMenu = append(TmpMenu, v)
// 			}
// 		}
// 		normalMenus = TmpMenu
// 	}
// 	var adminMenuList = []types.AdminMenu{}
// 	if len(normalMenus) == 0 {
// 		response.Success(ctx, adminMenuList)
// 		return
// 	}

// 	// "path": "/home/index",
// 	// "name": "HomeIndex",
// 	// "component": "/home/index",
// 	// "meta": {
// 	// 	"icon": "HomeFilled",
// 	// 	"title": "首页",
// 	// 	"isLink": "",
// 	// 	"isHide": false,
// 	// 	"isFull": false,
// 	// 	"isAffix": true,
// 	// 	"isKeepAlive": true
// 	// }
// 	// 慷慨一点，首页直接给，不做配置了
// 	normalMenus = append(normalMenus, model.Menu{
// 		Id:                 -999,
// 		Type:               model.MENU_TYPE_MENU,
// 		Sort:               -999,
// 		Status:             model.MENU_STATUS_ON,
// 		WebRouterPath:      "/home/index",
// 		WebRouterName:      "HomeIndex",
// 		WebRouterComponent: "/home/index",
// 		Icon:               "HomeFilled",
// 		Title:              "首页",
// 		TitleEn:            "Home",
// 		IsLink:             "",
// 		IsHide:             0,
// 		IsFull:             0,
// 		IsAffix:            1,
// 		IsKeepAlive:        1,
// 	})

// 	MenuList := GetMenuList(normalMenus)
// 	MenuTree := SortMenuTree(MenuToTree(MenuList, 0)) //  树状结构 && 排序

// 	response.Success(ctx, FilterAdminMenu(MenuTree))
// }

// // 管理员节点
// func (c *AdminController) AdminNode(ctx *gin.Context) {
// 	auth, ok := help.GetLoginAdmin(ctx)
// 	if !ok {
// 		response.ApiError(ctx, "获取登录信息失败", nil)
// 		return
// 	}
// 	var normalNodes []model.Menu                      // 有效菜单
// 	var allNodes = (&repository.Menu{}).GetAllNodes() // 全部节点
// 	for _, v := range allNodes {
// 		if v.Status == model.MENU_STATUS_OFF {
// 			continue
// 		}
// 		normalNodes = append(normalNodes, v)
// 	}
// 	// 不是超管，需要根据权限过滤
// 	if auth.IsSuper != model.ADMIN_IS_SUPER_Y {
// 		roleDB := repository.Role{}
// 		RolePermissionMenuIds := tool.UniqueArray(roleDB.GetPermission(auth.RoleId)) // 管理员拥有的菜单ID （包含了节点）
// 		var tmpNodes []model.Menu
// 		for _, v := range normalNodes {
// 			if ok := tool.InArray(v.Id, RolePermissionMenuIds); ok {
// 				tmpNodes = append(tmpNodes, v)
// 			}
// 		}
// 		normalNodes = tmpNodes
// 	}
// 	// 取出节点字符串
// 	var nodeStrings = []string{}
// 	for _, v := range normalNodes {
// 		nodeStrings = append(nodeStrings, v.ApiPermission)
// 	}

// 	response.Success(ctx, map[string][]string{
// 		"list": nodeStrings,
// 	})
// }
