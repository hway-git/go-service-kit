// email 包提供共享SMTP邮件发送和JSON消息处理，不依赖业务模型。
package email

import (
	"errors"
	"net"
	"net/mail"
	"os"
	"strconv"
	"strings"
)

var ErrConfig = errors.New("邮箱配置无效，请检查发件地址、SMTP主机、端口及登录凭据")

// Config 保存共享SMTP参数，解析和校验不连接邮件服务器。
type Config struct {
	EmailFrom       string `json:"from"`
	EmailServerHost string `json:"server_host"`
	EmailServerPort int    `json:"server_port"`
	EmailUsername   string `json:"username"`
	EmailPassword   string `json:"password,omitempty"`
}

// Configured 判断是否填写了任一邮箱配置，空配置不会阻止其他业务启动。
func (c *Config) Configured() bool {
	return c != nil && (c.EmailFrom != "" || c.EmailServerHost != "" || c.EmailServerPort != 0 || c.EmailUsername != "" || c.EmailPassword != "")
}

// ApplyEnvironment 应用显式设置的 SMTP 环境变量。
func (c *Config) ApplyEnvironment() error {
	for _, entry := range []struct {
		name  string
		field *string
	}{
		{"EMAIL_FROM", &c.EmailFrom}, {"EMAIL_SERVER_HOST", &c.EmailServerHost},
		{"EMAIL_USERNAME", &c.EmailUsername}, {"EMAIL_PASSWORD", &c.EmailPassword},
	} {
		if value, exists := os.LookupEnv(entry.name); exists {
			*entry.field = value
		}
	}
	if value, exists := os.LookupEnv("EMAIL_SERVER_PORT"); exists {
		c.EmailServerPort = 0
		if value != "" {
			port, err := strconv.Atoi(value)
			if err != nil || port < 1 || port > 65535 {
				return ErrConfig
			}
			c.EmailServerPort = port
		}
	}
	return nil
}

// Validate 验证发送邮件所需配置，不访问SMTP服务，不在错误中包含配置原文。
func (c *Config) Validate() error {
	if c == nil || !singleLine(c.EmailFrom) || c.EmailServerPort < 1 || c.EmailServerPort > 65535 {
		return ErrConfig
	}
	if _, err := mail.ParseAddress(c.EmailFrom); err != nil {
		return ErrConfig
	}
	host := c.EmailServerHost
	if host == "" || host != strings.TrimSpace(host) || strings.ContainsAny(host, " /\\\r\n\t\x00") || (strings.ContainsAny(host, ":[]") && net.ParseIP(host) == nil) {
		return ErrConfig
	}
	if (c.EmailUsername == "") != (c.EmailPassword == "") || strings.ContainsAny(c.EmailUsername+c.EmailPassword, "\x00\r\n") {
		return ErrConfig
	}
	return nil
}
