package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// LoadDotEnv 读取本地环境文件，只补充不存在的变量，不输出文件内容或解析原文。
func LoadDotEnv(path string) error {
	if path == "" {
		path = ".env"
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return errors.New("无法读取.env配置文件")
	}
	// 先完整解析再写入环境，兼容Windows编辑器生成的UTF-8 BOM。
	values, err := godotenv.Unmarshal(strings.TrimPrefix(string(raw), "\uFEFF"))
	if err != nil {
		return errors.New(".env配置文件格式无效，请检查变量名和引号")
	}
	for name, value := range values {
		if _, exists := os.LookupEnv(name); exists {
			continue
		}
		if os.Setenv(name, value) != nil {
			return errors.New(".env环境变量加载失败，请检查变量格式")
		}
	}
	return nil
}
