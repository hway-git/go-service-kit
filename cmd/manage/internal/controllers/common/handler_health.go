package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping 返回服务存活状态。
// @Summary 检查管理服务存活状态
// @Tags 公共接口
// @Produce json
// @Success 200 {string} string "pong"
// @Failure 907 {object} response.LegacyError "IP访问限制"
// @Router /ping [post]
func (c *CommonController) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "pong")
}
