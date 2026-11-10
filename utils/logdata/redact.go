// logdata 包统一处理请求与响应日志脱敏，不修改实际传输内容。
package logdata

import (
	"encoding/json"
	"io"
	"net/url"
	"strings"
)

const Hidden = "[已隐藏]"

func normalized(key string) string {
	return strings.ToLower(strings.NewReplacer("_", "", "-", "", ".", "").Replace(key))
}

func sensitive(key string) bool {
	key = normalized(key)
	if strings.Contains(key, "password") || strings.Contains(key, "secret") || strings.HasSuffix(key, "token") {
		return true
	}
	switch key {
	case "merchantkey", "sign", "fingerprint", "credential":
		return true
	case "pwd", "authpwd", "salt", "phone", "authorization", "param", "apikey", "accesskeyid", "xamzsignature", "xamzcredential", "xamzsecuritytoken", "verificationcode", "verifycode", "emailcode", "otp", "otpcode", "captcha", "email":
		return true
	}
	return false
}

func codeValue(value any) bool {
	switch value.(type) {
	case string, []any, []string:
		return true
	}
	return false
}

func redact(value any) any {
	switch data := value.(type) {
	case map[string]any:
		for key, child := range data {
			if sensitive(key) || (normalized(key) == "code" && codeValue(child)) {
				data[key] = Hidden
			} else {
				data[key] = redact(child)
			}
		}
	case []any:
		for index, child := range data {
			data[index] = redact(child)
		}
	case string:
		if signedURL(data) {
			return "[临时访问地址已隐藏]"
		}
	}
	return value
}

// JSON 严格解析完整JSON并递归脱敏，保留数字原值，无法解析时不输出原文。
func JSON(body string) any {
	if body == "" {
		return nil
	}
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	var data any
	if decoder.Decode(&data) != nil || decoder.Decode(new(any)) != io.EOF {
		return "[正文已隐藏]"
	}
	if _, ok := data.(string); ok {
		return "[文本正文已隐藏]"
	}
	return redact(data)
}

// Text 供后台操作日志保存脱敏后的JSON文本。
func Text(body string) string {
	if body == "" {
		return ""
	}
	data := JSON(body)
	if marker, ok := data.(string); ok {
		return marker
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return "[正文已隐藏]"
	}
	return string(raw)
}

// Query 将查询或普通表单参数转为结构化日志，解析失败时隐藏全部内容。
func Query(raw string) any {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return "[参数无效，已隐藏]"
	}
	data := make(map[string]any, len(values))
	for key, items := range values {
		entries := make([]any, len(items))
		for index, item := range items {
			entries[index] = item
		}
		data[key] = entries
	}
	return redact(data)
}

func signedURL(address string) bool {
	u, err := url.Parse(address)
	if err != nil {
		lower := strings.ToLower(address)
		return strings.Contains(lower, "token=") || strings.Contains(lower, "x-amz-signature=")
	}
	if u.User != nil {
		return true
	}
	for key := range u.Query() {
		if sensitive(key) {
			return true
		}
	}
	return false
}

// URL 复制URL后隐藏查询凭据，不修改业务路由收到的地址。
func URL(original *url.URL) string {
	copyURL := *original
	query := copyURL.Query()
	for key, values := range query {
		if sensitive(key) || normalized(key) == "code" {
			query.Set(key, Hidden)
		} else {
			for index, value := range values {
				if signedURL(value) {
					values[index] = "[临时访问地址已隐藏]"
				}
			}
		}
	}
	copyURL.RawQuery = query.Encode()
	return copyURL.String()
}
