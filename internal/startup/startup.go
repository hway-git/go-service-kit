package startup

import (
	"errors"

	"example.com/go-service-kit/db"
	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/db/redis"
	"example.com/go-service-kit/internal/httpserver"
	"example.com/go-service-kit/shared/config"
	"example.com/go-service-kit/shared/services/email"
	"example.com/go-service-kit/shared/services/r2"
	"example.com/go-service-kit/shared/services/rsa"
	"example.com/go-service-kit/shared/types"
	"example.com/go-service-kit/utils/logger"

	"github.com/gin-gonic/gin"
)

type Runtime struct {
	settings *config.Startup
	Addr     string
	Admin    config.AdminConfig
	Mail     *email.Client
	Storage  r2.Storage
	RSA      *rsa.Keys
}

// Load 加载启动配置并初始化日志及启用的基础服务。
func Load(service string) (*Runtime, error) {
	cfg, err := config.LoadStartup(service)
	if err != nil {
		return nil, err
	}
	rt := &Runtime{settings: cfg, Addr: cfg.HTTP.APIAddr}
	if service == "manage" {
		rt.Addr = cfg.HTTP.ManageAddr
		rt.Admin = cfg.Admin
	}
	if cfg.Email.Configured() {
		rt.Mail, err = email.New(cfg.Email)
		if err != nil {
			return nil, err
		}
	}
	if cfg.R2.Enabled {
		rt.Storage, err = r2.New(cfg.R2, nil)
		if err != nil {
			return nil, err
		}
	}
	if cfg.RSA.Enabled {
		rt.RSA, err = rsa.New(cfg.RSA)
		if err != nil {
			return nil, err
		}
	}
	if cfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	logger.Init(&logger.Config{ServiceName: service, LogDir: cfg.Log.Dir, Level: cfg.Log.Level, Console: true})
	return rt, nil
}

// Connect 按入口所需启用连接。显式初始化仅要求 MySQL。
func (r *Runtime) Connect(requireMySQL, useRedis, requireRedis bool) error {
	mysqlConfigured := r.settings.MySQL.Dsn != ""
	redisConfigured := r.settings.Redis.Mode == types.RedisCluster || r.settings.Redis.Address != ""
	if requireMySQL && !mysqlConfigured {
		return errors.New("请配置 MYSQL_DSN")
	}
	if requireRedis && !redisConfigured {
		return errors.New("请配置 Redis 单点地址或集群节点")
	}
	if mysqlConfigured {
		db.InitMySQLFromSetting(&r.settings.Setting, r.settings.Log.Level == "debug")
	}
	if useRedis && redisConfigured {
		db.InitRedisFromSetting(&r.settings.Setting)
	}
	return nil
}

func (r *Runtime) Close() {
	if redis.RDB != nil {
		redis.Close()
	}
	mysql.Close()
	logger.Sync()
}

func (r *Runtime) Router() (*gin.Engine, error) {
	return httpserver.NewRouter(r.settings.HTTP.TrustedProxies, r.settings.HTTP.CORSOrigins)
}
