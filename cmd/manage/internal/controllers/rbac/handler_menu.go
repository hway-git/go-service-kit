package rbac

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/validatorx"
	"example.com/go-service-kit/utils/tool"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var Menu = new(MenuController)

type MenuController struct{}

// 菜单列表
// @Summary 获取后台菜单树
// @Tags 菜单管理
// @Produce json
// @Security BearerAuth
// @Param is_only_menu query int false "0包含权限节点，其他值只返回菜单" default(0)
// @Param title query string false "菜单标题关键词"
// @Success 200 {object} response.Response{data=[]types.MenuList} "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /menu [get]
func (*MenuController) List(ctx *gin.Context) {
	// 参数
	isOnlyMenu := ctx.DefaultQuery("is_only_menu", "0")
	title := ctx.DefaultQuery("title", "")

	// 全部菜单(包括节点和菜单)
	allTableMenu := (&repository.Menu{}).GetAll()
	// 是否纯菜单显示
	if isOnlyMenu != "0" {
		TmpMenu := []model.Menu{}
		for _, v := range allTableMenu {
			if v.Type == model.MENU_TYPE_MENU { //菜单
				TmpMenu = append(TmpMenu, v)
			}
		}
		allTableMenu = TmpMenu
	}
	// 筛选标题
	if title != "" {
		TmpMenu := []model.Menu{}
		for _, v := range allTableMenu {
			if strings.Contains(v.Title, title) {
				TmpMenu = append(TmpMenu, v)
			}
		}
		allTableMenu = TmpMenu
	}
	// 填充字段
	var menuList = GetMenuList(allTableMenu)
	// 树状结构
	if len(menuList) > 0 {
		if title != "" {
			menuList = FilterTitleMenu(menuList)
		} else {
			menuList = MenuToTree(menuList, 0) // 树状结构
		}
	}
	// 排序
	menuList = SortMenuTree(menuList)

	response.Success(ctx, menuList)
}

// 新增菜单
// @Summary 新增菜单或权限节点
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.MenuAddRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /menu [post]
func (*MenuController) Add(ctx *gin.Context) {
	// 接收参数
	var params types.MenuAddRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		msg := validatorx.ErrorMessage(ctx, err, MenuAddValidator)
		response.ApiError(ctx, msg, nil)
		return
	}

	// 实例化DB
	menuDB := repository.Menu{}

	//如果没有，则用中文的
	if params.TitleEn == "" {
		params.TitleEn = params.Title
	}

	//类型
	switch params.Type {
	// 菜单
	case model.MENU_TYPE_MENU:
		{
			// 组件和重定向以及外链不能一起空
			if params.WebRouterComponent == "" && params.WebRouterRedirect == "" && params.IsLink == "" {
				response.ApiError(ctx, "参数错误", nil)
				return
			}
			if params.Icon == "" {
				params.Icon = "Cloudy"
			}
			params.WebRouterName = strings.ReplaceAll(strings.TrimLeft(strings.TrimSpace(params.WebRouterPath), "/"), "/", "-")
			// 清空节点字段
			params.ApiMethod = ""
			params.ApiPath = ""
		}
	// 节点
	case model.MENU_TYPE_NODE:
		{
			if params.ApiPath == "" || params.ApiMethod == "" {
				response.ApiError(ctx, "参数错误", nil)
				return
			}
			if ok := tool.InArray(params.ApiMethod, []string{"GET", "POST", "PUT", "DELETE"}); !ok {
				response.ApiError(ctx, "参数错误", nil)
				return
			}
			// 清空前端字段
			params.Icon = ""
			params.WebRouterComponent = ""
			params.WebRouterName = ""
			params.WebRouterRedirect = ""
			params.IsLink = ""
			params.IsHide = 0
			params.IsFull = 0
			params.IsAffix = 0
			params.IsKeepAlive = 0
		}
	}

	// 新增数据
	menu := model.Menu{
		Title:              params.Title,
		TitleEn:            params.TitleEn,
		Pid:                params.Pid,
		Icon:               params.Icon,
		ApiMethod:          params.ApiMethod,
		ApiPath:            params.ApiPath,
		WebRouterPath:      params.WebRouterPath,
		WebRouterName:      params.WebRouterName,
		WebRouterRedirect:  params.WebRouterRedirect,
		WebRouterComponent: params.WebRouterComponent,
		Type:               params.Type,
		Sort:               params.Sort,
		Status:             params.Status,
		Remark:             params.Remark,
		IsLink:             params.IsLink,
		IsHide:             params.IsHide,
		IsFull:             params.IsFull,
		IsAffix:            params.IsAffix,
		IsKeepAlive:        params.IsKeepAlive,
	}

	if _, err := menuDB.Add(menu); err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}

	response.Success(ctx, nil)
}

// 编辑菜单
// @Summary 修改菜单或权限节点
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.MenuEditRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /menu [put]
func (*MenuController) Edit(ctx *gin.Context) {
	// 接收参数
	var params types.MenuEditRequest
	if err := ctx.ShouldBind(&params); err != nil {
		response.ApiError(ctx, "参数错误", nil)
		return
	}
	// 实例化DB
	menuDB := repository.Menu{}

	// 获取菜单
	menu := model.Menu{Id: params.Id}
	//如果没有，则用中文的
	if params.TitleEn == "" {
		params.TitleEn = params.Title
	}

	//菜单
	if params.Type == model.MENU_TYPE_MENU {
		// 组件和重定向以及外链不能一起空
		if params.WebRouterComponent == "" && params.WebRouterRedirect == "" && params.IsLink == "" {
			response.ApiError(ctx, "参数错误", nil)
			return
		}
		if params.Icon == "" {
			params.Icon = "Cloudy"
		}
		params.WebRouterName = strings.ReplaceAll(strings.TrimLeft(strings.TrimSpace(params.WebRouterPath), "/"), "/", "-")
		// 清空节点字段
		params.ApiMethod = ""
		params.ApiPath = ""
	}
	// 节点
	if params.Type == model.MENU_TYPE_NODE {
		if params.ApiPath == "" || params.ApiMethod == "" {
			response.ApiError(ctx, "参数错误", nil)
			return
		}
		if ok := tool.InArray(params.ApiMethod, []string{"GET", "POST", "PUT", "DELETE"}); !ok {
			response.ApiError(ctx, "参数错误", nil)
			return
		}
		// 清空前端字段
		params.Icon = ""
		params.WebRouterComponent = ""
		params.WebRouterName = ""
		params.WebRouterRedirect = ""
		params.IsLink = ""
		params.IsHide = 0
		params.IsFull = 0
		params.IsAffix = 0
		params.IsKeepAlive = 0
	}

	// 修改数据
	menu.Title = params.Title
	menu.TitleEn = params.TitleEn
	menu.Pid = params.Pid
	menu.Icon = params.Icon

	menu.ApiMethod = params.ApiMethod
	menu.ApiPath = params.ApiPath

	menu.WebRouterPath = params.WebRouterPath
	menu.WebRouterName = params.WebRouterName
	menu.WebRouterRedirect = params.WebRouterRedirect
	menu.WebRouterComponent = params.WebRouterComponent

	menu.Type = params.Type
	menu.Sort = params.Sort
	menu.Status = params.Status
	menu.Remark = params.Remark

	menu.IsLink = params.IsLink
	menu.IsHide = params.IsHide
	menu.IsFull = params.IsFull
	menu.IsAffix = params.IsAffix
	menu.IsKeepAlive = params.IsKeepAlive

	if err := menuDB.Update(&menu); err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}

	response.Success(ctx, nil)
}

// 删除菜单
// @Summary 删除菜单及子级
// @Tags 菜单管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "菜单ID" minimum(1)
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /menu [delete]
func (*MenuController) Delete(ctx *gin.Context) {
	// 接收参数
	Id := ctx.DefaultQuery("id", "0")
	var params = types.MenuDeleteRequest{}
	if num, err := strconv.Atoi(Id); err == nil {
		params.Id = num
	}
	if params.Id == 0 {
		response.ApiError(ctx, "参数错误", nil)
		return
	}
	// 实例化DB
	menuDB := repository.Menu{}
	if err := menuDB.Delete(params.Id); err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}

	// 返回
	response.Success(ctx, nil)
}

// 菜单更新缓存
// @Summary 刷新菜单缓存
// @Tags 菜单管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /menu/refresh [get]
func (*MenuController) Refresh(ctx *gin.Context) {
	if _, err := (&repository.Menu{}).ReadAll(); err != nil {
		response.ServerError(ctx)
		return
	}
	response.Success(ctx, nil)
}

// 修改菜单状态
// @Summary 修改菜单状态
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.MenuEditRequest true "请求参数"
// @Success 200 {object} response.Response "操作成功"
// @Failure 901 {object} response.Response "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /menu/set/status [put]
func (*MenuController) SetStatus(ctx *gin.Context) {
	// 接收参数
	var params types.MenuEditRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		response.ApiError(ctx, "参数错误", nil)
		return
	}
	// 实例化DB
	menuDB := repository.Menu{}

	if err := menuDB.SetStatus(params.Id, params.Status); err != nil {
		response.ApiError(ctx, err.Error(), nil)
		return
	}

	response.Success(ctx, nil)
}
