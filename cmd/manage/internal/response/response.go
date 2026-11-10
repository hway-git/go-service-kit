package response

import (
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`              // 业务错误码
	Message string      `json:"message,omitempty"` // 提示信息
	Data    interface{} `json:"data,omitempty"`    // 响应数据
}

// Success 返回成功响应  --  response.Success(ctx, userInfo, "注册成功")
func Success(ctx *gin.Context, data interface{}, message ...string) {
	if len(message) > 0 {
		JSON(ctx, http.StatusOK, CODE_SUCCESS, message[0], data)
	} else {
		JSON(ctx, http.StatusOK, CODE_SUCCESS, "", data)
	}
}

// Error 返回业务错误响应（默认 HTTP 400） --   response.Error(ctx, response.CodeParamError, "参数不合法", nil)
func Error(ctx *gin.Context, code int, message string, data interface{}) {
	JSON(ctx, 206, code, message, data)
}

// ApiError
func ApiError(ctx *gin.Context, message string, data interface{}) {
	JSON(ctx, 901, 901, message, data)
}

// HTTPError 返回 HTTP 层错误（比如 401、500 等）
func HTTPError(ctx *gin.Context, httpCode int, code int, message string, data interface{}) {
	JSON(ctx, httpCode, code, message, data)
}

// Login Expired Error
func LoginExpiredError(ctx *gin.Context, code int, msg ...string) {
	if len(msg) > 0 {
		JSON(ctx, 206, code, msg[0], nil)
	} else {
		JSON(ctx, 206, code, "Login Expired", nil)
	}
}

// ServerError
func ServerError(ctx *gin.Context, msg ...string) {
	if len(msg) > 0 {
		JSON(ctx, 206, SERVER_ERROR, msg[0], nil)
	} else {
		JSON(ctx, 206, SERVER_ERROR, "Server Error", nil)
	}
}

// JSON 统一响应构造器
func JSON(ctx *gin.Context, httpCode int, code int, message string, data interface{}) {
	resp := Response{
		Code:    code,
		Message: help.T(ctx, message),
	}
	if data != nil {
		resp.Data = data
	}
	ctx.JSON(httpCode, resp)
}
