package config

import (
	"os"
	"testing"
)

func TestLoadStartupLayering(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("CONFIG_FILE", "")
	t.Setenv("LOG_DIR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MANAGE_ADDR", "127.0.0.1:9081")
	t.Setenv("EMAIL_PASSWORD", "from-env")
	if err := os.Mkdir("config", 0755); err != nil {
		t.Fatal(err)
	}
	for path, data := range map[string]string{
		"config/common.json": `{"log":{"dir":"./shared-logs","level":"warn"},"http":{"manage_addr":"127.0.0.1:8082"},"email":{"from":"test@example.com","server_host":"smtp.example.com","server_port":465,"username":"test","password":""},"mysql":{"dsn":""},"redis":{"password":""},"r2":{"access_key_id":"","secret_access_key":""},"rsa":{"private_key_file":""}}`,
		"config/manage.json": `{"log":{"level":"debug"},"http":{"manage_addr":"127.0.0.1:8083"}}`,
	} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := LoadStartup("manage")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Log.Dir != "./shared-logs" || cfg.Log.Level != "debug" || cfg.HTTP.ManageAddr != "127.0.0.1:9081" {
		t.Fatalf("unexpected layered config: log=%+v manage_addr=%s", cfg.Log, cfg.HTTP.ManageAddr)
	}
	if cfg.Email.EmailPassword != "from-env" {
		t.Fatal("EMAIL_PASSWORD must override the JSON placeholder")
	}
	cfg, err = LoadStartup("api")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Log.Level != "warn" {
		t.Fatalf("API must use the shared value when its file is absent, got %s", cfg.Log.Level)
	}
}
