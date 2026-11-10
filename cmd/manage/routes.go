package main

import (
	"example.com/go-service-kit/cmd/manage/internal/controllers/auth"
	"example.com/go-service-kit/cmd/manage/internal/controllers/common"
	"example.com/go-service-kit/cmd/manage/internal/controllers/ipwhitelist"
	"example.com/go-service-kit/cmd/manage/internal/controllers/rbac"
	"example.com/go-service-kit/cmd/manage/internal/middlewares"
	"example.com/go-service-kit/shared/services/email"

	"github.com/gin-gonic/gin"
)

func ConsoleRouters(router *gin.Engine, mail *email.Client, secret string) {
	authController := auth.NewController(mail, secret)
	manage := router.Group("", middlewares.IPWhiteMiddleware())
	manage.POST("/ping", common.Common.Ping)
	manage.POST("/auth/login", authController.Login)
	manage.POST("/auth/login/verify-email", authController.VerifyEmail)
	manage.POST("/common/tz-set", common.System.TimezoneSet)

	secured := manage.Group("", middlewares.AuthMiddleware(), middlewares.LoggingMiddleware())
	secured.GET("/common/init", common.Common.Init)
	secured.POST("/auth/logout", authController.Logout)
	secured.POST("/auth/reset-pwd", authController.ResetPwd)
	secured.POST("/auth/verify", authController.Verify)

	// 管理员及授权关系的写操作只允许超级管理员执行。
	super := middlewares.SuperOnly
	admin := secured.Group("/admin")
	admin.GET("/list", rbac.Admin.List)
	admin.GET("/all", rbac.Admin.All)
	admin.GET("/log", rbac.Admin.AdminLogList)
	admin.POST("", super, rbac.Admin.Add)
	admin.PUT("", super, rbac.Admin.Edit)
	admin.DELETE("", super, rbac.Admin.Delete)
	admin.POST("/kick", super, rbac.Admin.KickLogin)

	menu := secured.Group("/menu")
	menu.GET("", rbac.Menu.List)
	menu.GET("/refresh", rbac.Menu.Refresh)
	menu.POST("", super, rbac.Menu.Add)
	menu.PUT("", super, rbac.Menu.Edit)
	menu.DELETE("", super, rbac.Menu.Delete)
	menu.PUT("/set/status", super, rbac.Menu.SetStatus)

	role := secured.Group("/role")
	role.GET("", rbac.Role.List)
	role.GET("/all", rbac.Role.All)
	role.GET("/detail", rbac.Role.Detail)
	role.GET("/permission", rbac.Role.GetPermission)
	role.POST("", super, rbac.Role.Add)
	role.PUT("", super, rbac.Role.Edit)
	role.DELETE("", super, rbac.Role.Delete)
	role.POST("/permission", super, rbac.Role.SetPermission)

	group := role.Group("/group")
	group.GET("", rbac.RoleGroup.List)
	group.GET("/all", rbac.RoleGroup.All)
	group.POST("", super, rbac.RoleGroup.Add)
	group.PUT("", super, rbac.RoleGroup.Edit)
	group.DELETE("", super, rbac.RoleGroup.Delete)

	ips := secured.Group("/ip/whitelist")
	ips.GET("/list", ipwhitelist.IpWhitelist.List)
	ips.GET("/status", ipwhitelist.IpWhitelist.GetStatus)
	ips.GET("/sync", super, ipwhitelist.IpWhitelist.Sync)
	ips.POST("/add", super, ipwhitelist.IpWhitelist.Add)
	ips.DELETE("/del", super, ipwhitelist.IpWhitelist.Delete)
	ips.PUT("/status", super, ipwhitelist.IpWhitelist.SetStatus)

	secured.GET("/system/config/list", common.Common.SettingsList)
	secured.PUT("/system/config/edit", super, common.Common.SettingsEdit)
}
