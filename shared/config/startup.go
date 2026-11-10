package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"example.com/go-service-kit/shared/services/email"
	"example.com/go-service-kit/shared/services/r2"
	"example.com/go-service-kit/shared/services/rsa"
	"example.com/go-service-kit/shared/types"
)

type Startup struct {
	types.Setting
	HTTP  HTTPConfig   `json:"http"`
	Log   LogConfig    `json:"log"`
	Email email.Config `json:"email"`
	R2    r2.Config    `json:"r2"`
	RSA   rsa.Config   `json:"rsa"`
	Admin AdminConfig  `json:"admin"`
}

type HTTPConfig struct {
	APIAddr        string   `json:"api_addr"`
	ManageAddr     string   `json:"manage_addr"`
	CORSOrigins    []string `json:"cors_origins"`
	TrustedProxies []string `json:"trusted_proxies"`
}

type LogConfig struct {
	Dir   string `json:"dir"`
	Level string `json:"level"`
}

type AdminConfig struct {
	JWTSecret         string `json:"jwt_secret,omitempty"`
	RedisKeyPrefix    string `json:"redis_key_prefix"`
	BootstrapUser     string `json:"bootstrap_user"`
	BootstrapPassword string `json:"bootstrap_password,omitempty"`
}

// LoadStartup 合并 Go 默认值、共享配置、应用配置和环境变量。
func LoadStartup(service string) (*Startup, error) {
	if err := LoadDotEnv(""); err != nil {
		return nil, err
	}
	cfg := &Startup{
		HTTP: HTTPConfig{APIAddr: "127.0.0.1:8080", ManageAddr: "127.0.0.1:8081"},
		Log:  LogConfig{Dir: "./logs", Level: "info"},
	}
	cfg.MySQL.MySQLConfig = types.MySQLConfig{Idle: 5, Open: 20, Lifetime: 300}
	cfg.Redis.RedisConfig = types.RedisConfig{Mode: types.RedisSingle, PoolSize: 10}

	if path := os.Getenv("CONFIG_FILE"); path != "" {
		if err := cfg.loadJSON(path); err != nil {
			return nil, fmt.Errorf("加载 %s: %w", path, err)
		}
	} else {
		for _, name := range []string{"common.json", service + ".json"} {
			path := filepath.Join("config", name)
			if err := cfg.loadJSON(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("加载 %s: %w", path, err)
			}
		}
	}
	if err := ApplyConnectionEnvironment(&cfg.Setting); err != nil {
		return nil, err
	}
	for _, entry := range []struct {
		env   string
		field *string
	}{
		{"API_ADDR", &cfg.HTTP.APIAddr},
		{"MANAGE_ADDR", &cfg.HTTP.ManageAddr},
		{"LOG_DIR", &cfg.Log.Dir},
		{"LOG_LEVEL", &cfg.Log.Level},
	} {
		if value := strings.TrimSpace(os.Getenv(entry.env)); value != "" {
			*entry.field = value
		}
	}
	if value, ok := os.LookupEnv("CORS_ORIGINS"); ok {
		cfg.HTTP.CORSOrigins = splitList(value)
	}
	if value, ok := os.LookupEnv("TRUSTED_PROXIES"); ok {
		cfg.HTTP.TrustedProxies = splitList(value)
	}
	cfg.Log.Level = strings.ToLower(strings.TrimSpace(cfg.Log.Level))
	switch cfg.Log.Level {
	case "debug", "info", "warn", "warning", "error":
	default:
		return nil, errors.New("日志级别须为 debug、info、warn 或 error")
	}
	if cfg.HTTP.APIAddr == "" || cfg.HTTP.ManageAddr == "" {
		return nil, errors.New("HTTP 监听地址不能为空")
	}

	overrideString("JWT_SECRET", &cfg.Admin.JWTSecret)
	overrideString("REDIS_KEY_PREFIX", &cfg.Admin.RedisKeyPrefix)
	cfg.Admin.RedisKeyPrefix = strings.TrimSpace(cfg.Admin.RedisKeyPrefix)
	overrideString("BOOTSTRAP_ADMIN_USER", &cfg.Admin.BootstrapUser)
	overrideString("BOOTSTRAP_ADMIN_PASSWORD", &cfg.Admin.BootstrapPassword)
	if err := cfg.Email.ApplyEnvironment(); err != nil {
		return nil, err
	}
	cfg.R2.ApplyEnvironment()
	overrideString("RSA_PRIVATE_KEY_FILE", &cfg.RSA.PrivateKeyFile)
	return cfg, nil
}

func overrideString(name string, target *string) {
	if value, exists := os.LookupEnv(name); exists {
		*target = value
	}
}

func (c *Startup) loadJSON(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	raw = bytes.TrimSpace(bytes.TrimPrefix(raw, []byte("\uFEFF")))
	if len(raw) == 0 || raw[0] != '{' {
		return errors.New("启动配置文件须为 JSON 对象")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(c) != nil || decoder.Decode(new(any)) != io.EOF {
		return errors.New("启动配置文件格式无效或含有未支持的字段")
	}
	return nil
}

func splitList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}
