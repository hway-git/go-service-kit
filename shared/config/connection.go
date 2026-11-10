package config

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"

	"example.com/go-service-kit/shared/types"
)

// environmentName 将连接名称转换为固定环境变量片段。
func environmentName(name string) string {
	return strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(name))
}

// ApplyConnectionEnvironment 注入连接凭据并校验 Redis 拓扑和单点地址覆盖。
func ApplyConnectionEnvironment(setting *types.Setting) error {
	setting.MySQL.Dsn = os.Getenv("MYSQL_DSN")
	for i := range setting.MySQL.Replica {
		setting.MySQL.Replica[i].Dsn = os.Getenv("MYSQL_REPLICA_" + strconv.Itoa(i) + "_DSN")
	}
	for name, connection := range setting.MySQL.Other {
		connection.Dsn = os.Getenv("MYSQL_OTHER_" + environmentName(name) + "_DSN")
		setting.MySQL.Other[name] = connection
	}
	setting.Redis.Password = os.Getenv("REDIS_PASSWORD")
	for name, connection := range setting.Redis.Other {
		connection.Password = os.Getenv("REDIS_OTHER_" + environmentName(name) + "_PASSWORD")
		setting.Redis.Other[name] = connection
	}
	return resolveRedis(&setting.Redis.RedisConfig, strings.TrimSpace(os.Getenv("REDIS_ADDRESS")))
}

func resolveRedis(cfg *types.RedisConfig, envAddress string) error {
	switch cfg.Mode {
	case types.RedisSingle:
		if len(cfg.Addrs) > 0 {
			return errors.New("配置 redis.addrs 时须声明 redis.mode 为 cluster")
		}
		if envAddress != "" {
			cfg.Address = envAddress
		}
		if cfg.Address != "" && !validRedisAddress(cfg.Address) {
			return errors.New("Redis 单点地址须为单个 host:port")
		}
	case types.RedisCluster:
		if envAddress != "" || cfg.Address != "" {
			return errors.New("Redis 集群模式不能设置 REDIS_ADDRESS 或 redis.address")
		}
		if len(cfg.Addrs) == 0 {
			return errors.New("Redis 集群模式须配置 redis.addrs")
		}
		if cfg.DB != 0 {
			return errors.New("Redis 集群只支持数据库 0")
		}
		for _, addr := range cfg.Addrs {
			if !validRedisAddress(addr) {
				return errors.New("redis.addrs 中的节点须为 host:port")
			}
		}
	default:
		return errors.New("redis.mode 须为 single 或 cluster")
	}
	return nil
}

func validRedisAddress(addr string) bool {
	host, port, err := net.SplitHostPort(addr)
	number, portErr := strconv.ParseUint(port, 10, 16)
	return err == nil && portErr == nil && number > 0 && host != "" && !strings.ContainsAny(host, ", /\t\r\n")
}
