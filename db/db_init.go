package db

import (
	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/db/redis"
	"example.com/go-service-kit/shared/types"
	"log"
)

// InitMySQLFromSetting 根据统一配置初始化 MySQL
//
// - MySQL:
//   - 若配置了 mysql.replica，则启用主从（InitWithSlaves）
//   - 否则单库（Init）
//   - mysql.other 会以 name->dsn 建立额外命名连接（mysql.GetDB(name) 获取）
//
// debug 用于控制 MySQL gorm logger。
func InitMySQLFromSetting(s *types.Setting, debug bool) {
	if s == nil {
		return
	}

	// MySQL
	if len(s.MySQL.Replica) > 0 {
		slaves := make([]string, 0, len(s.MySQL.Replica))
		for _, r := range s.MySQL.Replica {
			if r.Dsn != "" {
				slaves = append(slaves, r.Dsn)
			}
		}
		mysql.InitWithSlaves(s.MySQL.Dsn, slaves, s.MySQL.Idle, s.MySQL.Open, s.MySQL.Lifetime, debug)
	} else {
		mysql.Init(s.MySQL.Dsn, s.MySQL.Idle, s.MySQL.Open, s.MySQL.Lifetime, debug)
	}

	// MySQL other
	if len(s.MySQL.Other) > 0 {
		for name, cfg := range s.MySQL.Other {
			if name == "" || cfg.Dsn == "" {
				continue
			}
			idle := cfg.Idle
			open := cfg.Open
			lifetime := cfg.Lifetime
			if idle == 0 {
				idle = s.MySQL.Idle
			}
			if open == 0 {
				open = s.MySQL.Open
			}
			if lifetime == 0 {
				lifetime = s.MySQL.Lifetime
			}
			if err := mysql.AddDB(name, cfg.Dsn, idle, open, lifetime, debug); err != nil {
				log.Printf("AddDB failed (name=%s): %v", name, err)
			}
		}
	}
}

// InitRedisFromSetting 使用已解析的 Redis 模式初始化连接。
func InitRedisFromSetting(s *types.Setting) {
	if s == nil {
		return
	}

	// Redis
	if s.Redis.Mode == types.RedisCluster {
		redis.InitCluster(s.Redis.Addrs, s.Redis.Password, s.Redis.PoolSize, s.Redis.Tls)
	} else {
		redis.Init(s.Redis.Address, s.Redis.Password, s.Redis.PoolSize, s.Redis.DB, s.Redis.Tls)
	}
}
