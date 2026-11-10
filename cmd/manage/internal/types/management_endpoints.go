package types

import "strings"

type BusinessEndpoint struct{ Method, Path, Title string }

func (e *BusinessEndpoint) Permission() string {
	return strings.ReplaceAll(strings.Trim(e.Path, "/"), "/", ".") + "." + strings.ToLower(e.Method)
}

// ManagementEndpoints 列出需要明确授权的后台管理接口。
var ManagementEndpoints = []BusinessEndpoint{
	{"GET", "/system/config/list", "系统配置列表"}, {"PUT", "/system/config/edit", "修改系统配置"},
	{"GET", "/admin/list", "管理员列表"}, {"GET", "/admin/all", "管理员选项"},
	{"GET", "/admin/log", "管理员操作日志"}, {"POST", "/admin", "新增管理员"},
	{"PUT", "/admin", "修改管理员"}, {"DELETE", "/admin", "删除管理员"},
	{"POST", "/admin/kick", "踢下线管理员"},
	{"GET", "/menu", "菜单列表"}, {"POST", "/menu", "新增菜单"},
	{"PUT", "/menu", "修改菜单"}, {"DELETE", "/menu", "删除菜单"},
	{"GET", "/menu/refresh", "刷新菜单"}, {"PUT", "/menu/set/status", "设置菜单状态"},
	{"GET", "/ip/whitelist/list", "IP白名单列表"}, {"POST", "/ip/whitelist/add", "新增IP白名单"},
	{"DELETE", "/ip/whitelist/del", "删除IP白名单"}, {"GET", "/ip/whitelist/sync", "刷新IP白名单"},
	{"GET", "/ip/whitelist/status", "IP白名单开关"}, {"PUT", "/ip/whitelist/status", "修改IP白名单开关"},
	{"GET", "/role", "角色列表"}, {"POST", "/role", "新增角色"},
	{"PUT", "/role", "修改角色"}, {"DELETE", "/role", "删除角色"},
	{"GET", "/role/all", "角色选项"}, {"GET", "/role/detail", "角色详情"},
	{"GET", "/role/permission", "查询角色权限"}, {"POST", "/role/permission", "设置角色权限"},
	{"GET", "/role/group", "角色分组列表"}, {"POST", "/role/group", "新增角色分组"},
	{"PUT", "/role/group", "修改角色分组"}, {"DELETE", "/role/group", "删除角色分组"},
	{"GET", "/role/group/all", "角色分组选项"},
}
