package httpserver

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"example.com/go-service-kit/internal/language"
	"example.com/go-service-kit/utils/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const maxBodyBytes = 1 << 20

func NewRouter(trustedProxies, allowedOrigins []string) (*gin.Engine, error) {
	router := gin.New()
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		return nil, errors.New("TRUSTED_PROXIES 包含无效地址或网段")
	}
	localize, err := language.Middleware()
	if err != nil {
		return nil, err
	}

	router.Use(accessLog, gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, _ any) {
		logger.FromContext(c.Request.Context()).Error("HTTP 请求发生异常", zap.Stack("stack"))
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	if len(allowedOrigins) > 0 {
		for _, origin := range allowedOrigins {
			if strings.Contains(origin, "*") {
				return nil, errors.New("CORS_ORIGINS 须填写具体来源")
			}
		}
		settings := cors.Config{
			AllowOrigins:  allowedOrigins,
			AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", "Timezone", "Accept-Language"},
			ExposeHeaders: []string{"X-Request-ID", "Authorization", "Content-Language"},
			MaxAge:        12 * time.Hour,
		}
		if err := settings.Validate(); err != nil {
			return nil, err
		}
		router.Use(cors.New(settings))
	}
	router.Use(localize, limitBody)
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return router, nil
}

func accessLog(c *gin.Context) {
	started := time.Now()
	requestID := uuid.NewString()
	c.Header("X-Request-ID", requestID)
	requestLog := logger.FromContext(c.Request.Context()).With(
		zap.String("request_id", requestID),
		zap.String("route", c.FullPath()),
	)
	c.Request = c.Request.WithContext(logger.WithContext(c.Request.Context(), requestLog))
	c.Next()

	status := c.Writer.Status()
	fields := []zap.Field{
		zap.String("method", c.Request.Method),
		zap.String("ip", c.ClientIP()),
		zap.Int("status", status),
		zap.Duration("duration", time.Since(started)),
	}
	switch {
	case status >= 500 && status < 600:
		requestLog.Error("HTTP 访问", fields...)
	case status >= 400:
		requestLog.Warn("HTTP 访问", fields...)
	default:
		requestLog.Info("HTTP 访问", fields...)
	}
}

func limitBody(c *gin.Context) {
	if c.Request.ContentLength > maxBodyBytes {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": language.T(c, "请求正文超限")})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	c.Next()
}
