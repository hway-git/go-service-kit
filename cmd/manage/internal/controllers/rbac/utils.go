package rbac

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// 填充侧边菜单
func FilterAdminMenu(Tree []types.MenuList) []types.AdminMenu {
	var menus = []types.AdminMenu{}
	for _, v := range Tree {
		Menu := types.AdminMenu{
			Path:      v.WebRouterPath,
			Name:      PathToComponentName(v.WebRouterPath),
			Redirect:  v.WebRouterRedirect,
			Component: v.WebRouterComponent,
			Meta: types.AdminMenuMeta{
				Title:       v.Title,
				Icon:        v.Icon,
				IsLink:      v.IsLink,
				IsHide:      v.IsHide == 1,
				IsAffix:     v.IsAffix == 1,
				IsFull:      v.IsFull == 1,
				IsKeepAlive: v.IsKeepAlive == 1,
			},
			Children: FilterAdminMenu(v.Children),
		}
		menus = append(menus, Menu)
	}
	return menus
}

// PathToComponentName 把路由路径转成组件名（PascalCase）
// "/user/list"        	->  UserList
// "user-login-log"    	->  UserLoginLog
// "system.config.v2"    	->  SystemConfigV2
// "user_profile.edit"	->  UserProfileEdit
// "测试/abc-123-xyz"	->  测试Abc123Xyz
func PathToComponentName(path string) string {

	var splitterRegexp = regexp.MustCompile(`[/\-\_.]+`) // 正则匹配 / - _ .

	// 替换为统一的空格
	path = splitterRegexp.ReplaceAllString(path, " ")

	// 按空格切割
	parts := strings.Fields(path)

	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		// 如果首字符是字母，转大写
		if unicode.IsLetter(runes[0]) {
			runes[0] = unicode.ToUpper(runes[0])
		}
		parts[i] = string(runes)
	}

	return strings.Join(parts, "")
}

// 处理 title 菜单, 是树状，但不确定 Pid
//
//	func FilterTitleMenu(Menus []MenuList) []MenuList {
//		var tree []MenuList
//		// 确定所有Pid
//		var pids []int
//		for _, Menu := range Menus {
//			for _, v := range Menus {
//				if Menu.Pid == v.Id {
//					break
//				}
//			}
//			pids = append(pids, Menu.Pid)
//		}
//		for _, Pid := range helper.UniqueArray(pids) {
//			var tmpMenuTree = MenuToTree(Menus, Pid) // 树状结构
//			tree = append(tree, tmpMenuTree...)
//		}
//		return tree
//	}
func FilterTitleMenu(Menus []types.MenuList) []types.MenuList {
	// 使用集合来存储唯一的 Pid
	pidSet := make(map[int]struct{})
	// 遍历 Menus，确定所有 Pid
	for _, menu := range Menus {
		pidSet[menu.Pid] = struct{}{}
	}
	var tree []types.MenuList
	// 遍历唯一的 Pid 并构建树状结构
	for pid := range pidSet {
		tmpMenuTree := MenuToTree(Menus, pid) // 树状结构
		tree = append(tree, tmpMenuTree...)
	}
	return tree
}

// 菜单 树状结构 （递归）
func MenuToTree(menuList []types.MenuList, pid int) []types.MenuList {
	children := make(map[int][]types.MenuList)
	for _, menu := range menuList {
		children[menu.Pid] = append(children[menu.Pid], menu)
	}
	visited := map[int]bool{pid: true}
	var build func(int) []types.MenuList
	build = func(parent int) []types.MenuList {
		tree := []types.MenuList{}
		for _, menu := range children[parent] {
			if visited[menu.Id] {
				continue
			}
			visited[menu.Id] = true
			menu.Children = build(menu.Id)
			tree = append(tree, menu)
		}
		return tree
	}
	return build(pid)
}

// 对树状结构排序（ 'sort' 'asc'）
func SortMenuTree(TreeData []types.MenuList) []types.MenuList {
	sort.Slice(TreeData, func(i, j int) bool {
		return TreeData[i].Sort < TreeData[j].Sort // 按照 Sort 字段从小到大排序
	})
	// 遍历数组，递归 children
	for _, v := range TreeData {
		if len(v.Children) > 0 {
			v.Children = SortMenuTree(v.Children)
		}
	}
	return TreeData
}

// 菜单 树状结构 （递归）
func GetMenuList(menus []model.Menu) []types.MenuList {
	var menuList []types.MenuList
	for _, v := range menus {
		menu := types.MenuList{
			Menu:         v,
			CreatedAt:    "",
			UpdatedAt:    "",
			TypeRemark:   "",
			StatusRemark: "",
		}
		if _TypeRemark, ok := model.MENU_TYPE_MAP[v.Type]; ok {
			menu.TypeRemark = _TypeRemark
		}
		if _StatusRemark, ok := model.MENU_STATUS_MAP[v.Status]; ok {
			menu.StatusRemark = _StatusRemark
		}
		menuList = append(menuList, menu)
	}
	return menuList
}

// 获取菜单下的所有子菜单ID
func FindChildrenIds(menuList []model.Menu, id int) []int {
	children := make(map[int][]int)
	for _, menu := range menuList {
		children[menu.Pid] = append(children[menu.Pid], menu.Id)
	}
	visited := map[int]bool{id: true}
	var result []int
	var visit func(int)
	visit = func(parent int) {
		for _, child := range children[parent] {
			if visited[child] {
				continue
			}
			visited[child] = true
			result = append(result, child)
			visit(child)
		}
	}
	visit(id)
	return result
}
