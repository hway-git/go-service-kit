package middlewares

import (
	"example.com/go-service-kit/cmd/manage/internal/config"
	"example.com/go-service-kit/cmd/manage/internal/response"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"

	"github.com/gin-gonic/gin"
)

func IPWhiteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 判断IP是否在白名单中
		if !config.Config.AllowsIP(c.ClientIP()) {
			c.JSON(response.HTTP_NOT_ALLOWED_IP, gin.H{"error": help.T(c, "Not allowed IP")})
			c.Abort()
			return
		}

		c.Next()
	}
}
