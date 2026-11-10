package common

import (
	"example.com/go-service-kit/cmd/manage/internal/controllers/rbac"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/utils/tool"

	"github.com/gin-gonic/gin"
)

// 管理员的菜单
func getAdminMenu(ctx *gin.Context, auth model.Admin) []types.AdminMenu {

	var normalMenus []model.Menu                      // 有效菜单
	var allMenus = (&repository.Menu{}).GetAllMenus() // 全部菜单
	for _, v := range allMenus {
		if v.Status == model.MENU_STATUS_ON {
			normalMenus = append(normalMenus, v)
		}
	}

	// 替换语言
	var lang string = help.GetLanguage(ctx) // 语言
	for k, v := range normalMenus {
		switch lang {
		case "en":
			normalMenus[k].Title = v.TitleEn
		case "vi":
			normalMenus[k].Title = help.T(ctx, v.Title)
		}
	}
	// 不是管理员，需要根据权限过滤
	if auth.IsSuper != model.ADMIN_IS_SUPER_Y {
		roleDB := repository.Role{}
		ids, err := roleDB.GetMenuIds(auth.RoleId)
		if err != nil {
			return []types.AdminMenu{}
		}
		RoleMenuIds := tool.UniqueArray(ids) // 管理员拥有的菜单ID
		TmpMenu := []model.Menu{}
		for _, v := range normalMenus {
			if ok := tool.InArray(v.Id, RoleMenuIds); ok {
				TmpMenu = append(TmpMenu, v)
			}
		}
		normalMenus = TmpMenu
	}

	// "path": "/home/index",
	// "name": "HomeIndex",
	// "component": "/home/index",
	// "meta": {
	// 	"icon": "HomeFilled",
	// 	"title": "首页",
	// 	"isLink": "",
	// 	"isHide": false,
	// 	"isFull": false,
	// 	"isAffix": true,
	// 	"isKeepAlive": true
	// }

	// 慷慨一点，首页直接给，不做配置了
	normalMenus = append(normalMenus, model.Menu{
		Id:                 -999,
		Type:               model.MENU_TYPE_MENU,
		Sort:               -999,
		Status:             model.MENU_STATUS_ON,
		WebRouterPath:      "/home/index",
		WebRouterName:      "HomeIndex",
		WebRouterComponent: "/home/index",
		Icon:               "HomeFilled",
		Title:              help.T(ctx, "首页"),
		TitleEn:            "Home",
		IsLink:             "",
		IsHide:             0,
		IsFull:             0,
		IsAffix:            1,
		IsKeepAlive:        1,
	})

	MenuList := rbac.GetMenuList(normalMenus)
	MenuTree := rbac.SortMenuTree(rbac.MenuToTree(MenuList, 0)) //  树状结构 && 排序

	return rbac.FilterAdminMenu(MenuTree)
}

// 管理员节点
func getAdminNode(ctx *gin.Context, auth model.Admin) []string {

	var normalNodes []model.Menu                      // 有效菜单
	var allNodes = (&repository.Menu{}).GetAllNodes() // 全部节点
	for _, v := range allNodes {
		if v.Status != model.MENU_STATUS_ON || v.ApiPermission == "" {
			continue
		}
		normalNodes = append(normalNodes, v)
	}
	// 不是超管，需要根据权限过滤
	if auth.IsSuper != model.ADMIN_IS_SUPER_Y {
		roleDB := repository.Role{}
		ids, err := roleDB.GetPermission(auth.RoleId)
		if err != nil {
			return []string{}
		}
		RolePermissionMenuIds := tool.UniqueArray(ids) // 管理员拥有的权限节点ID
		var tmpNodes []model.Menu
		for _, v := range normalNodes {
			if ok := tool.InArray(v.Id, RolePermissionMenuIds); ok {
				tmpNodes = append(tmpNodes, v)
			}
		}
		normalNodes = tmpNodes
	}
	// 取出节点字符串
	var nodeStrings = []string{}
	for _, v := range normalNodes {
		nodeStrings = append(nodeStrings, string(v.ApiPermission))
	}
	return nodeStrings
}
