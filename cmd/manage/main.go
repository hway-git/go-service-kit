package main

import (
	"errors"
	"flag"
	"log"

	"example.com/go-service-kit/cmd/manage/internal/bootstrap"
	"example.com/go-service-kit/cmd/manage/internal/cache"
	"example.com/go-service-kit/cmd/manage/internal/config"
	"example.com/go-service-kit/cmd/manage/internal/utils/jwt"
	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/internal/httpserver"
	"example.com/go-service-kit/internal/startup"

	// 为管理员时区选项内嵌时区数据库。
	_ "time/tzdata"
)

func main() {
	initDB := flag.Bool("init-db", false, "初始化后台表、管理员与权限后退出")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("不支持位置参数")
	}
	if err := run(*initDB); err != nil {
		log.Fatal(err)
	}
}

func run(initDB bool) error {
	runtime, err := startup.Load("manage")
	if err != nil {
		return err
	}
	defer runtime.Close()

	if initDB {
		if err := runtime.Connect(true, false, false); err != nil {
			return err
		}
		return bootstrap.Initialize(mysql.DB, runtime.Admin.BootstrapUser, runtime.Admin.BootstrapPassword)
	}

	if err := jwt.Configure(runtime.Admin.JWTSecret); err != nil {
		return err
	}
	cache.Prefix = runtime.Admin.RedisKeyPrefix
	if cache.Prefix == "" {
		return errors.New("请配置独立的 REDIS_KEY_PREFIX")
	}
	if err := runtime.Connect(true, true, true); err != nil {
		return err
	}
	stopRefresh, err := config.NewConfig()
	if err != nil {
		return err
	}
	defer stopRefresh()
	router, err := runtime.Router()
	if err != nil {
		return err
	}
	ConsoleRouters(router, runtime.Mail, runtime.Admin.JWTSecret)

	return httpserver.Run(runtime.Addr, router)
}
