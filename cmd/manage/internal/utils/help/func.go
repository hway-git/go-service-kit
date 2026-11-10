package help

import (
	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/internal/language"
	"example.com/go-service-kit/utils/enum"
	"example.com/go-service-kit/utils/tool"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ##########   后台   ################

// 获取接口头部时区
func GetTimezone(ctx *gin.Context) string {
	adminTz := ctx.GetHeader("Timezone")
	if tool.InArray(adminTz, enum.TIMEZONE_LIST) {
		return adminTz
	}
	return enum.TIMEZONE_DEFAULT
}

// 获取当前在线用户信息
func GetLoginAdmin(ctx *gin.Context) (model.Admin, bool) {
	value, exists := ctx.Get("admin")
	if !exists {
		return model.Admin{}, false
	}
	authAdmin, ok := value.(model.Admin)
	return authAdmin, ok
}

// ##########   公共   ################

// 翻译多语言
func GetLanguage(c *gin.Context) string {
	var lang string = c.MustGet("lang").(string)
	return lang
}

// 翻译多语言
func T(c *gin.Context, message string) string {
	return language.T(c, message)
}

func Tf(c *gin.Context, format string, args ...interface{}) string {
	translated := make([]any, len(args))
	for index, value := range args {
		if text, ok := value.(string); ok {
			value = T(c, text)
		}
		translated[index] = value
	}
	return fmt.Sprintf(T(c, format), translated...)
}

func GetIpGeolocation(ip string, lang ...string) string {
	if strings.HasPrefix(ip, "192.168.") || ip == "127.0.0.1" || ip == "::1" {
		return ip
	}
	return "未知"
}

// 把开始结束时间格式 按时区返回时间戳
func GetStartAndEndTs(TimeStart, TimeEnd, tz string) (int64, int64, error) {

	// urldecode 解码， 如果不需要解码，会返回原样
	// decoded, err := url.QueryUnescape("2024-01-01 00:00:00")
	// decoded = "2024-01-01 00:00:00" (相同)
	// err = nil (无错误)

	decoded, err := url.QueryUnescape(TimeStart)
	if err != nil {
		return 0, 0, err
	}
	TimeStart = decoded

	decoded, err = url.QueryUnescape(TimeEnd)
	if err != nil {
		return 0, 0, err
	}
	TimeEnd = decoded

	startTs, err1 := tool.GetTimeZoneToTimestamp(TimeStart, tz)
	if err1 != nil {
		return 0, 0, err1
	}

	endTs, err2 := tool.GetTimeZoneToTimestamp(TimeEnd, tz)
	if err2 != nil {
		return 0, 0, err2
	}
	return startTs, endTs, nil
}

// 获取两个时间戳之间的月份
func MonthRangeStartAndEndTs(startTs, endTs int64) int {
	start := time.Unix(int64(startTs), 0)
	end := time.Unix(int64(endTs), 0)
	y1, m1 := start.Year(), int(start.Month())
	y2, m2 := end.Year(), int(end.Month())
	return (y2-y1)*12 + (m2 - m1) + 1
}
