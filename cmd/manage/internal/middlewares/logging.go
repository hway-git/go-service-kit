package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"

	"example.com/go-service-kit/cmd/manage/internal/model"
	"example.com/go-service-kit/cmd/manage/internal/repository"
	"example.com/go-service-kit/cmd/manage/internal/utils/help"
	"example.com/go-service-kit/utils/logdata"
	"example.com/go-service-kit/utils/logger"

	"github.com/gin-gonic/gin"
)

const auditLimit = 32 << 10

// LoggingMiddleware 将后台写操作的脱敏摘要保存到审计表。
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			c.Next()
			return
		}

		start := time.Now()
		request := auditRequest(c)
		writer := &auditWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()

		admin, _ := help.GetLoginAdmin(c)
		audit := model.AdminLog{
			RequestIp:     c.ClientIP(),
			RequestUrl:    c.FullPath(),
			RequestMethod: c.Request.Method,
			RequestData:   request,
			ResponseData:  writer.summary(),
			Status:        c.Writer.Status(),
			Duration:      time.Duration(time.Since(start).Milliseconds()),
			Operator:      admin.Username,
		}
		if err := (&repository.AdminLog{}).Create(&audit); err != nil {
			logger.FromContext(c.Request.Context()).Error("后台操作审计写入失败")
		}
	}
}

func auditRequest(c *gin.Context) string {
	if c.Request.Method == http.MethodDelete {
		query := c.Request.URL.RawQuery
		if len(query) == 0 {
			return ""
		}
		if len(query) > auditLimit {
			return logdata.Hidden
		}
		data, err := json.Marshal(logdata.Query(query))
		if err != nil {
			return logdata.Hidden
		}
		return string(data)
	}
	contentType, _, _ := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if contentType != "application/json" || c.Request.Body == nil {
		return ""
	}
	original := c.Request.Body
	data, err := io.ReadAll(io.LimitReader(original, auditLimit+1))
	c.Request.Body = struct {
		io.Reader
		io.Closer
	}{Reader: io.MultiReader(bytes.NewReader(data), original), Closer: original}
	if err != nil || len(data) > auditLimit {
		return logdata.Hidden
	}
	return logdata.Text(string(data))
}

type auditWriter struct {
	gin.ResponseWriter
	body    bytes.Buffer
	omitted bool
}

func (w *auditWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.capture(data[:n])
	return n, err
}

func (w *auditWriter) WriteString(data string) (int, error) {
	n, err := w.ResponseWriter.WriteString(data)
	if n > auditLimit-w.body.Len() {
		w.omitted = true
	}
	if !w.omitted {
		w.body.WriteString(data[:n])
	}
	return n, err
}

func (w *auditWriter) capture(data []byte) {
	if len(data) > auditLimit-w.body.Len() {
		w.omitted = true
	}
	if !w.omitted {
		w.body.Write(data)
	}
}

func (w *auditWriter) summary() string {
	if w.omitted {
		return logdata.Hidden
	}
	return logdata.Text(w.body.String())
}
