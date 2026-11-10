package mysql

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	applogger "example.com/go-service-kit/utils/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 设计目标：
// - 服务启动时必须能连上 MySQL（Ping 成功），否则直接退出进程，避免 mysql.DB == nil 导致的各种 panic
// - 保留现有 public API/全局变量（mysql.DB / Init / InitWithSlaves / Slave 等），尽量不改业务调用点
// - 去掉重连/健康检查/多库 InitMulti 等复杂逻辑（仓库里无引用），代码更短更直观

var (
	DB         *gorm.DB            // 默认数据库连接（向后兼容）
	Slaves     []*gorm.DB          // 从库连接列表（向后兼容）
	slaveCount int                 // 从库数量
	dbs        map[string]*gorm.DB // 命名数据库连接映射（用于 AddDB/GetDB）
	dbMutex    sync.RWMutex
)

const defaultPingTimeout = 5 * time.Second

func gormConfig(debug bool) *gorm.Config {
	level := logger.Silent
	if debug {
		level = logger.Info
	}
	// 调试模式也不展开SQL参数，避免记录密码摘要、盐和推流凭据。
	return &gorm.Config{Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold: 200 * time.Millisecond, LogLevel: level, ParameterizedQueries: true,
	})}
}

func openAndPing(dsn string, idle, open int, lifetime int64, debug bool) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DSN")
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig(debug))
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(idle)
	sqlDB.SetMaxOpenConns(open)
	sqlDB.SetConnMaxLifetime(time.Duration(lifetime) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return db, nil
}

// Init initializes the database connection (single database mode).
// 如果 MySQL 连不上：直接退出进程（log.Fatalf），避免后续业务任何地方访问 mysql.DB 产生 nil panic。
func Init(dsn string, idle, open int, lifetime int64, debug bool) {
	db, err := openAndPing(dsn, idle, open, lifetime, debug)
	if err != nil {
		log.Fatalf("MySQL connect failed: %v", err)
	}
	DB = db
	Slaves = nil
	slaveCount = 0
	applogger.Info("MySQL连接初始化完成，使用单库模式")
}

// InitWithSlaves initializes the database connection with primary-replica support.
// 规则：
// - 主库必须能连上，否则退出进程
// - 从库：尽量连接；连不上则跳过（读会自动回退到主库）
func InitWithSlaves(masterDSN string, slaveDSNs []string, idle, open int, lifetime int64, debug bool) {
	master, err := openAndPing(masterDSN, idle, open, lifetime, debug)
	if err != nil {
		log.Fatalf("MySQL primary connect failed: %v", err)
	}
	DB = master

	Slaves = make([]*gorm.DB, 0, len(slaveDSNs))
	for i, dsn := range slaveDSNs {
		if dsn == "" {
			continue
		}
		slave, err := openAndPing(dsn, idle, open, lifetime, debug)
		if err != nil {
			applogger.Warn("MySQL从库连接失败，已跳过", zap.Int("replica", i+1), zap.Error(err))
			continue
		}
		Slaves = append(Slaves, slave)
	}
	slaveCount = len(Slaves)

	if slaveCount > 0 {
		applogger.Info("MySQL连接初始化完成，使用主从模式", zap.Int("replica_count", slaveCount))
	} else {
		applogger.Info("MySQL连接初始化完成，无可用从库，使用主库")
	}
}

// InitSimple kept for backward compatibility.
func InitSimple(dsn string, idle, open int, lifetime int64, debug bool) {
	Init(dsn, idle, open, lifetime, debug)
}

// Master returns the primary database connection.
func Master() *gorm.DB { return DB }

// Slave returns a randomly selected replica; if none, returns primary.
func Slave() *gorm.DB {
	if slaveCount == 0 {
		return DB
	}
	return Slaves[rand.Intn(slaveCount)]
}

// GetSlave returns a specific replica by index; fallback to primary if invalid.
func GetSlave(index int) *gorm.DB {
	if slaveCount == 0 || index < 0 || index >= slaveCount {
		return DB
	}
	return Slaves[index]
}

func SlaveCount() int { return slaveCount }

// AddDB adds a named database connection (used by db/db_init.go for mysql.other).
// 这里返回 error（由调用方决定是否 fatal）。
func AddDB(name string, dsn string, idle, open int, lifetime int64, debug bool) error {
	if name == "" || dsn == "" {
		return fmt.Errorf("database name and DSN cannot be empty")
	}
	db, err := openAndPing(dsn, idle, open, lifetime, debug)
	if err != nil {
		return fmt.Errorf("connect database '%s' failed: %w", name, err)
	}

	dbMutex.Lock()
	if dbs == nil {
		dbs = make(map[string]*gorm.DB)
	}
	dbs[name] = db
	dbMutex.Unlock()
	return nil
}

// GetDB returns a named database connection; nil if not exists.
func GetDB(name string) *gorm.DB {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	if dbs == nil {
		return nil
	}
	return dbs[name]
}

func ListDBs() []string {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	if dbs == nil {
		return []string{}
	}
	names := make([]string, 0, len(dbs))
	for name := range dbs {
		names = append(names, name)
	}
	return names
}

func DBExists(name string) bool {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	if dbs == nil {
		return false
	}
	_, ok := dbs[name]
	return ok
}

// IsInitialized checks if primary DB is initialized and pingable.
func IsInitialized() bool {
	if DB == nil {
		return false
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()
	return sqlDB.PingContext(ctx) == nil
}

// Close closes all database connections (primary, replicas, and named dbs).
func Close() {
	// primary
	if DB != nil {
		if sqlDB, err := DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
		DB = nil
	}

	// replicas
	for _, sdb := range Slaves {
		if sdb == nil {
			continue
		}
		if sqlDB, err := sdb.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	Slaves = nil
	slaveCount = 0

	// named dbs
	dbMutex.Lock()
	for name, ndb := range dbs {
		if ndb == nil {
			continue
		}
		if sqlDB, err := ndb.DB(); err == nil {
			_ = sqlDB.Close()
		}
		delete(dbs, name)
	}
	dbMutex.Unlock()
}
