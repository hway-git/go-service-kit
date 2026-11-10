package language

import (
	"fmt"

	"example.com/go-service-kit/utils/tool"

	"github.com/gin-gonic/gin"
)

var supported = []string{"zh", "en", "pt", "vi"}

// Middleware 在启动时加载内嵌字典，随后为每次请求协商语言。
func Middleware() (gin.HandlerFunc, error) {
	dictionaries := make(map[string]map[string]string)
	manager := NewManager()
	for _, lang := range supported[1:] {
		dictionary, err := manager.Load(lang)
		if err != nil {
			return nil, fmt.Errorf("加载语言字典 %s 失败: %w", lang, err)
		}
		dictionaries[lang] = dictionary
	}

	return func(c *gin.Context) {
		lang := tool.MatchLanguage(c.GetHeader("Accept-Language"), "zh", supported...)
		c.Set("lang", lang)
		c.Request = c.Request.WithContext(tool.WithLanguage(c.Request.Context(), lang))
		c.Set("languageMap", dictionaries[lang])
		c.Header("Content-Language", lang)
		c.Writer.Header().Add("Vary", "Accept-Language")
		c.Next()
	}, nil
}

// T 使用当前请求的字典翻译，缺少译文时返回源文案。
func T(c *gin.Context, message string) string {
	value, _ := c.Get("languageMap")
	if dictionary, ok := value.(map[string]string); ok {
		if translated, ok := dictionary[message]; ok {
			return translated
		}
	}
	return message
}
