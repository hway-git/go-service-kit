package validatorx

import (
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// 获取参数错误信息
func ErrorMessage(ctx *gin.Context, err error, messages map[string]string) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		fieldErr := errs[0] // 取消循环，取第一个就行
		key := fmt.Sprintf("%s.%s", fieldErr.StructField(), fieldErr.Tag())
		if msg, exists := messages[key]; exists {
			return msg
		}
		return help.Tf(ctx, "参数 %s 错误 (%s)", fieldErr.Field(), fieldErr.Tag())
	}
	return "参数错误"
}

// func ErrorMessage(ctx *gin.Context, err error, messages map[string]string) string {
// 	if errs, ok := err.(validator.ValidationErrors); ok {
// 		for _, fieldErr := range errs {
// 			key := fmt.Sprintf("%s.%s", fieldErr.StructField(), fieldErr.Tag())
// 			if msg, exists := messages[key]; exists {
// 				return msg
// 			}
// 			return tool.T(ctx, "参数 %s 验证失败 (%s)", fieldErr.Field(), fieldErr.Tag())
// 		}
// 		return "参数验证失败"
// 	}
// 	return "参数验证失败"
// }
