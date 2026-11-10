// r2 包封装Cloudflare R2的S3兼容操作，供应用服务复用。
package r2

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const DefaultMaxUploadBytes int64 = 20 << 20

type Config struct {
	Enabled         bool   `json:"enabled"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	PublicBaseURL   string `json:"public_base_url"`
	MaxUploadBytes  int64  `json:"max_upload_bytes"`
	AccessKeyID     string `json:"-"`
	SecretAccessKey string `json:"-"`
}

func (c *Config) defaults() {
	if c.MaxUploadBytes == 0 {
		c.MaxUploadBytes = DefaultMaxUploadBytes
	}
}

// ApplyEnvironment 只从环境读取 R2 凭据。
func (c *Config) ApplyEnvironment() {
	c.AccessKeyID = os.Getenv("R2_ACCESS_KEY_ID")
	c.SecretAccessKey = os.Getenv("R2_SECRET_ACCESS_KEY")
}

var bucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

func (c *Config) Validate() error {
	u, err := url.Parse(c.Endpoint)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("R2 endpoint须为无凭据、路径或参数的HTTPS地址")
	}
	if strings.TrimSpace(c.AccessKeyID) == "" || strings.TrimSpace(c.SecretAccessKey) == "" {
		return fmt.Errorf("缺少R2_ACCESS_KEY_ID或R2_SECRET_ACCESS_KEY")
	}
	if !bucketPattern.MatchString(c.Bucket) || strings.Contains(c.Bucket, "..") {
		return fmt.Errorf("R2 bucket名称无效")
	}
	if c.MaxUploadBytes <= 0 || c.MaxUploadBytes > 512<<20 {
		return fmt.Errorf("上传上限须在1字节至512MiB之间")
	}
	if c.PublicBaseURL != "" {
		public, err := url.Parse(c.PublicBaseURL)
		if err != nil || public.Scheme != "https" || public.Hostname() == "" || public.User != nil || public.RawQuery != "" || public.Fragment != "" {
			return fmt.Errorf("R2公开域名须为无凭据或参数的HTTPS地址")
		}
	}
	return nil
}
