package rbac

import (
	"errors"
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/shared/types/page"
	"example.com/go-service-kit/utils/tool"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var Role = new(RoleController)

type RoleController struct{}

// 角色列表
// @Summary 查询角色
// @Tags 角色管理
// @Produce json
// @Security BearerAuth
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页条数" minimum(1) maximum(100)
// @Param name query string false "角色名称关键词"
// @Success 200 {object} response.LegacyData{data=types.RoleListResponse} "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role [get]
func (*RoleController) List(ctx *gin.Context) {
	// 参数
	var params types.RoleListRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	// 数据仓储
	roleDB := repository.Role{}

	// 列表
	roleList, total, err := roleDB.RoleList(params)
	if err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, err.Error())})
		return
	}
	// 时区
	adminTz := help.GetTimezone(ctx)
	//返回列表
	roleListRows := make([]types.RoleListRow, 0, len(roleList))
	for _, v := range roleList {
		row := types.RoleListRow{
			Id:        v.Id,
			GroupId:   v.GroupId,
			Name:      v.Name,
			Status:    v.Status,
			Remark:    v.Remark,
			GroupName: v.GroupName,
			CreatedAt: tool.GetTimestampToTimeZone(v.CreatedTs, adminTz),
			NoDelete:  v.NoDelete,
		}
		roleListRows = append(roleListRows, row)
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{
		"data": types.RoleListResponse{
			Rows: roleListRows,
			PageResponse: page.PageResponse{
				Page:      params.Page,
				PageSize:  params.PageSize,
				PageTotal: len(roleListRows),
				Total:     total,
			},
		},
	})
}

// 新增角色
// @Summary 新增角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.RoleAddRequest true "请求参数"
// @Success 200 {object} response.LegacyMessage "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role [post]
func (*RoleController) Add(ctx *gin.Context) {
	// 参数
	var params types.RoleAddRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	// 数据仓储
	roleDB := repository.Role{}

	data := model.Role{
		GroupId:   params.GroupId,
		Name:      params.Name,
		Status:    params.Status,
		Remark:    params.Remark,
		CreatedTs: time.Now().Unix(),
	}
	if err := roleDB.Create(&data); err != nil {
		roleRequestError(ctx, err)
		return
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{"msg": help.T(ctx, "操作成功")})
}

// 编辑角色
// @Summary 修改角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.RoleEditRequest true "请求参数"
// @Success 200 {object} response.LegacyMessage "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role [put]
func (*RoleController) Edit(ctx *gin.Context) {
	// 参数
	var params types.RoleEditRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	// 数据仓储
	roleDB := repository.Role{}

	role := model.Role{Id: params.Id, GroupId: params.GroupId, Name: params.Name, Status: params.Status, Remark: params.Remark}
	if err := roleDB.Update(&role); err != nil {
		roleRequestError(ctx, err)
		return
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{"msg": help.T(ctx, "操作成功")})
}

// 删除角色
// @Summary 删除角色
// @Tags 角色管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "记录ID" minimum(1)
// @Success 200 {object} response.LegacyMessage "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role [delete]
func (*RoleController) Delete(ctx *gin.Context) {
	// 接收参数
	var params types.RoleDeleteRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}

	// 数据仓储
	roleDB := repository.Role{}

	if err := roleDB.Delete([]int{params.Id}); err != nil {
		roleRequestError(ctx, err)
		return
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{"msg": help.T(ctx, "操作成功")})
}

// 所有角色
// @Summary 获取角色选项
// @Tags 角色管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.LegacyData{data=types.RoleAllResponse} "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/all [get]
func (*RoleController) All(ctx *gin.Context) {
	rows, err := (&repository.Role{}).RoleAll()
	if err != nil {
		roleRequestError(ctx, err)
		return
	}
	list := make([]types.RoleAllSingle, 0, len(rows))
	for _, row := range rows {
		list = append(list, types.RoleAllSingle{Id: row.Id, Name: row.Name})
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{"data": types.RoleAllResponse{List: list}})
}

// 获取角色详情
// @Summary 获取角色详情
// @Tags 角色管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "角色ID" minimum(1)
// @Success 200 {object} response.LegacyData{data=types.RoleDetailResponse} "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/detail [get]
func (*RoleController) Detail(ctx *gin.Context) {
	var params types.RoleDeleteRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	row, err := (&repository.Role{}).RoleDetail(params.Id)
	if err != nil {
		roleRequestError(ctx, err)
		return
	}
	timezone := help.GetTimezone(ctx)
	ctx.JSON(response.HTTP_API_OK, gin.H{"data": types.RoleDetailResponse{
		Id: row.Id, GroupId: row.GroupId, GroupName: row.GroupName, Name: row.Name,
		Status: row.Status, Remark: row.Remark, NoDelete: row.NoDelete,
		CreatedAt: tool.GetTimestampToTimeZone(row.CreatedTs, timezone),
		UpdatedAt: tool.GetTimestampToTimeZone(row.UpdatedTs, timezone),
	}})
}

// 获取角色权限
// @Summary 获取角色菜单与权限
// @Tags 角色管理
// @Produce json
// @Security BearerAuth
// @Param role_id query int true "角色ID" minimum(1)
// @Success 200 {object} response.LegacyData{data=types.RoleGetPermissionResponse} "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/permission [get]
func (c *RoleController) GetPermission(ctx *gin.Context) {
	var params types.RoleGetPermissionRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	roleDB := repository.Role{}
	var role model.Role
	if err := roleDB.GetById(int64(params.RoleId), &role); err != nil {
		roleRequestError(ctx, err)
		return
	}
	menus, err := (&repository.Menu{}).ReadAll()
	if err != nil {
		roleRequestError(ctx, err)
		return
	}
	menuIDs, err := roleDB.GetMenuIds(params.RoleId)
	if err != nil {
		roleRequestError(ctx, err)
		return
	}
	permissionIDs, err := roleDB.GetPermission(params.RoleId)
	if err != nil {
		roleRequestError(ctx, err)
		return
	}
	normal := make([]types.MenuList, 0)
	checkedMenus, checkedPermissions := make([]int, 0), make([]int, 0)
	for _, row := range menus {
		if row.Status != model.MENU_STATUS_ON {
			continue
		}
		if help.GetLanguage(ctx) == "vi" {
			row.Title = help.T(ctx, row.Title)
		}
		normal = append(normal, types.MenuList{Menu: row})
		if row.Type == model.MENU_TYPE_MENU && tool.InArray(row.Id, menuIDs) {
			checkedMenus = append(checkedMenus, row.Id)
		}
		if row.Type == model.MENU_TYPE_NODE && tool.InArray(row.Id, permissionIDs) {
			checkedPermissions = append(checkedPermissions, row.Id)
		}
	}
	tree := SortMenuTree(MenuToTree(normal, 0))
	list := c.handleRolePermissionList(tree, checkedMenus, checkedPermissions)
	ctx.JSON(response.HTTP_API_OK, gin.H{"data": types.RoleGetPermissionResponse{
		List: list, CheckedMenuIds: checkedMenus, CheckedPermissionIds: checkedPermissions,
	}})
}

// 设置角色权限
// @Summary 设置角色菜单与权限
// @Description 校验有效节点及类型后完整替换授权，空字符串清空对应授权。
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.RoleSetPermissionRequest true "请求参数"
// @Success 200 {object} response.LegacyMessage "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/permission [post]
func (*RoleController) SetPermission(ctx *gin.Context) {
	var params types.RoleSetPermissionRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	menuIDs, err := parseRoleIDs(params.MenuIds)
	if err != nil {
		roleRequestError(ctx, err)
		return
	}
	permissionIDs, err := parseRoleIDs(params.PermissionIds)
	if err != nil {
		roleRequestError(ctx, err)
		return
	}
	if err := (&repository.Role{}).SetPermission(params.RoleId, menuIDs, permissionIDs); err != nil {
		roleRequestError(ctx, err)
		return
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{"msg": help.T(ctx, "操作成功")})
}

// handleRolePermissionList 按菜单和权限节点各自的授权设置勾选状态。
func (c *RoleController) handleRolePermissionList(rows []types.MenuList, menuIDs, permissionIDs []int) []types.RoleGetPermissionList {
	list := make([]types.RoleGetPermissionList, 0, len(rows))
	for _, row := range rows {
		item := types.RoleGetPermissionList{Id: row.Id, Title: row.Title, Type: row.Type}
		if (row.Type == model.MENU_TYPE_MENU && tool.InArray(row.Id, menuIDs)) ||
			(row.Type == model.MENU_TYPE_NODE && tool.InArray(row.Id, permissionIDs)) {
			item.Checked = 1
		}
		item.Children = c.handleRolePermissionList(row.Children, menuIDs, permissionIDs)
		list = append(list, item)
	}
	return list
}

// parseRoleIDs 保留逗号分隔参数格式，拒绝非正整数并去重。
func parseRoleIDs(value string) ([]int, error) {
	ids := make([]int, 0)
	if strings.TrimSpace(value) == "" {
		return ids, nil
	}
	if len(value) > 16384 {
		return nil, errors.New("授权ID数量过多")
	}
	seen := make(map[int]bool)
	for _, part := range strings.Split(value, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			return nil, errors.New("菜单或权限ID必须为正整数")
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func roleRequestError(ctx *gin.Context, err error) {
	message := err.Error()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		message = "角色不存在"
	}
	ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, message)})
}
