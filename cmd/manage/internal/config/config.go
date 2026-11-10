package config

import (
	"sync"
	"time"

	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/shared/models"
	"example.com/go-service-kit/utils/logger"
)

var Config *BaseConfig

type BaseConfig struct {
	mu        sync.RWMutex
	system    models.SysConfData
	whitelist map[string]bool
}

// NewConfig 首次加载失败时终止启动，定时刷新失败保留旧快照。
func NewConfig() (func(), error) {
	Config = &BaseConfig{}
	if err := Config.Reload(); err != nil {
		return nil, err
	}
	ticker := time.NewTicker(2 * time.Minute)
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if err := Config.Reload(); err != nil {
					logger.Warn("后台配置刷新失败，保留原配置")
				}
			}
		}
	}()
	return func() {
		ticker.Stop()
		close(stop)
		<-done
	}, nil
}

// Reload 在同一把锁内刷新系统开关及完整白名单，避免中途发布空名单。
func (c *BaseConfig) Reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	system, err := (&repository.SysConfig{}).LoadSysConfData()
	if err != nil {
		return err
	}
	var rows []model.IpWhitelist
	if err := (&repository.IpWhitelist{}).GetAll(&rows); err != nil {
		return err
	}
	whitelist := make(map[string]bool, len(rows))
	for _, row := range rows {
		whitelist[row.IP] = true
	}
	c.system, c.whitelist = system, whitelist
	return nil
}

// System 返回系统开关的值副本，调用方不能修改共享数据。
func (c *BaseConfig) System() models.SysConfData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.system
}

// AllowsIP 同时读取开关与白名单，避免两次读取跨越不同版本。
func (c *BaseConfig) AllowsIP(ip string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.system.Admin_IP_White_Switch == 0 || c.whitelist[ip]
}
