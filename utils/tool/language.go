package tool

import (
	"context"
	"strconv"
	"strings"
)

type languageContextKey struct{}

// WithLanguage 将已选择的业务展示语言放入请求上下文，避免服务层依赖Gin。
func WithLanguage(ctx context.Context, language string) context.Context {
	return context.WithValue(ctx, languageContextKey{}, MatchLanguage(language, "zh", "zh", "en", "pt", "vi"))
}

// LanguageFromContext 未设置展示语言时保持中文，后台任务不依赖请求语言。
func LanguageFromContext(ctx context.Context) string {
	language, _ := ctx.Value(languageContextKey{}).(string)
	if language == "" {
		return "zh"
	}
	return language
}

// MatchLanguage 按语言头的地区标签和q权重选择支持的语言，同权重保留先后顺序。
func MatchLanguage(header, fallback string, supported ...string) string {
	selected, best := fallback, -1.0
	for _, option := range strings.Split(header, ",") {
		parts := strings.Split(strings.TrimSpace(option), ";")
		lang := strings.ToLower(strings.SplitN(strings.TrimSpace(parts[0]), "-", 2)[0])
		allowed := false
		for _, candidate := range supported {
			if lang == candidate {
				allowed = true
				break
			}
		}
		if !allowed {
			continue
		}
		quality := 1.0
		for _, parameter := range parts[1:] {
			if key, value, ok := strings.Cut(strings.TrimSpace(parameter), "="); ok && strings.EqualFold(key, "q") {
				parsed, err := strconv.ParseFloat(value, 64)
				if err != nil || parsed < 0 || parsed > 1 {
					quality = 0
				} else {
					quality = parsed
				}
			}
		}
		if quality > 0 && quality > best {
			selected, best = lang, quality
		}
	}
	return selected
}
