package email

import (
	"errors"
	"net/mail"
	"strings"
)

var ErrAddress = errors.New("邮箱地址无效")

// NormalizeAddress 校验账号使用的单一邮箱地址，去除两端空白并统一小写。
func NormalizeAddress(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value || len(value) > 254 || strings.Count(value, "@") != 1 || strings.ContainsAny(value, "\"()<>[],;\\ \r\n\x00") {
		return "", ErrAddress
	}
	for _, char := range value {
		if char > 127 {
			return "", ErrAddress
		}
	}
	return value, nil
}
