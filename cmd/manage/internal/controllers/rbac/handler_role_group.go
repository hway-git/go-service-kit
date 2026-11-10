package rbac

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/types"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/shared/types/page"

	"github.com/gin-gonic/gin"
)

var RoleGroup = new(RoleGroupController)

type RoleGroupController struct{}

// 所有分组
// @Summary 获取角色分组选项
// @Tags 角色分组
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.LegacyData{data=types.RoleAllGroupResponse} "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/group/all [get]
func (c *RoleGroupController) All(ctx *gin.Context) {
	// 实例化DB
	roleGroupDB := repository.RoleGroup{}
	// 数据库取数据
	groups := roleGroupDB.RoleAllGroupAll()
	// 返回数据
	var list []map[string]any
	for _, v := range groups {
		list = append(list, map[string]any{"key": v.Id, "val": v.Name})
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{
		// "msg": "",
		"data": types.RoleAllGroupResponse{
			List: list,
		},
	})
}

// 角色组列表
// @Summary 查询角色分组
// @Tags 角色分组
// @Produce json
// @Security BearerAuth
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页条数" minimum(1)
// @Param name query string false "分组名称关键词"
// @Success 200 {object} response.LegacyData{data=types.RoleGroupListResponse} "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/group [get]
func (c *RoleGroupController) List(ctx *gin.Context) {
	// 参数
	var params types.RoleGroupListRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	//DB
	roleGroupDB := repository.RoleGroup{}

	// 列表
	roleList, total, err := roleGroupDB.RoleGroupList(params)
	if err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, err.Error())})
		return
	}
	//返回列表
	var rows []types.RoleGroupRow
	for _, v := range roleList {
		row := types.RoleGroupRow{
			Id:        v.Id,
			Name:      v.Name,
			Status:    v.Status,
			Remark:    v.Remark,
			RoleCount: v.RoleCount,
			NoDelete:  v.NoDelete,
		}
		rows = append(rows, row)
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{
		"data": types.RoleGroupListResponse{
			Rows: rows,
			PageResponse: page.PageResponse{
				Page:      params.Page,
				PageSize:  params.PageSize,
				PageTotal: len(rows),
				Total:     total,
			},
		},
	})
}

// 新增角色
// @Summary 新增角色分组
// @Tags 角色分组
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.RoleGroupAddRequest true "请求参数"
// @Success 200 {object} response.LegacyMessage "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/group [post]
func (c *RoleGroupController) Add(ctx *gin.Context) {
	// 参数
	var params types.RoleGroupAddRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	//DB
	roleGroupDB := repository.RoleGroup{}

	// 是否存在同名角色
	if existRole := roleGroupDB.GetByName(params.Name); existRole.Id > 0 {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "存在相同名称的角色组")})
		return
	}
	data := model.RoleGroup{
		Name:   params.Name,
		Status: params.Status,
		Remark: params.Remark,
	}
	if err := roleGroupDB.Create(&data); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "操作失败")})
		return
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{"msg": help.T(ctx, "操作成功")})
}

// 编辑角色
// @Summary 修改角色分组
// @Tags 角色分组
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body types.RoleGroupEditRequest true "请求参数"
// @Success 200 {object} response.LegacyMessage "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/group [put]
func (c *RoleGroupController) Edit(ctx *gin.Context) {
	// 参数
	var params types.RoleGroupEditRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	//DB
	roleGroupDB := repository.RoleGroup{}

	var role model.RoleGroup
	if params.Id <= 0 {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	if err := roleGroupDB.GetById(int64(params.Id), &role); err != nil { // 角色组不存在
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误"), "msg": help.T(ctx, "ID不存在")})
		return
	}
	// 是否存在同名角色
	if existRoleGroup := roleGroupDB.GetByNameNoSelf(params.Name, params.Id); existRoleGroup.Id > 0 {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "存在相同名称的角色")})
		return
	}

	// 修改数据
	role.Name = params.Name
	role.Status = params.Status
	role.Remark = params.Remark
	if err := roleGroupDB.Update(&role); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, help.T(ctx, err.Error()))
		return
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{"msg": help.T(ctx, "操作成功")})
}

// 删除角色
// @Summary 删除角色分组
// @Tags 角色分组
// @Produce json
// @Security BearerAuth
// @Param id query int true "记录ID" minimum(1)
// @Success 200 {object} response.LegacyMessage "操作成功"
// @Failure 901 {object} response.LegacyError "参数或业务错误"
// @Failure 902,903,904,906,907 {object} response.LegacyError "访问限制"
// @Router /role/group [delete]
func (c *RoleGroupController) Delete(ctx *gin.Context) {
	// 接收参数
	var params types.RoleGroupDeleteRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}

	//DB
	roleGroupDB := repository.RoleGroup{}

	var roleGroup model.RoleGroup
	if err := roleGroupDB.GetById(int64(params.Id), &roleGroup); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "参数错误")})
		return
	}
	if roleGroup.NoDelete == 1 {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, "该角色组不可删除")})
		return
	}

	ids := []int{}
	ids = append(ids, params.Id)

	if err := roleGroupDB.Delete(ids); err != nil {
		ctx.JSON(response.HTTP_API_ERROR, gin.H{"error": help.T(ctx, err.Error())})
		return
	}
	ctx.JSON(response.HTTP_API_OK, gin.H{"msg": help.T(ctx, "操作成功")})
}
