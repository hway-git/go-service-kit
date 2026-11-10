package logger

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"example.com/go-service-kit/utils/logdata"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
)

type contextKey struct{}

// ThirdParty 为三方交互选用独立分类，保留请求上下文和原有真实日志级别。
func ThirdParty(value *zap.Logger) *zap.Logger {
	if value == nil {
		value = getLogger()
	}
	return value.Named("third_party")
}

// categoryCore 先允许 INFO 进入检查，再按日志分类应用业务阈值，避免把成功回调伪装成 ERROR。
type categoryCore struct {
	core  zapcore.Core
	level zapcore.Level
}

func (c *categoryCore) Enabled(level zapcore.Level) bool {
	return level >= min(c.level, zapcore.InfoLevel)
}
func (c *categoryCore) With(fields []zapcore.Field) zapcore.Core {
	return &categoryCore{core: c.core.With(fields), level: c.level}
}
func (c *categoryCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	thirdParty := entry.LoggerName == "third_party" || strings.HasSuffix(entry.LoggerName, ".third_party")
	if entry.Level >= c.level || thirdParty && entry.Level >= zapcore.InfoLevel {
		return checked.AddCore(entry, c)
	}
	return checked
}
func (c *categoryCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return c.core.Write(entry, fields)
}
func (c *categoryCore) Sync() error { return c.core.Sync() }

// WithContext 将请求日志传给下游调用，保留请求标识等关联字段。
func WithContext(ctx context.Context, value *zap.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, contextKey{}, value)
}

// FromContext 优先使用请求日志，后台任务回退到进程日志。
func FromContext(ctx context.Context) *zap.Logger {
	if ctx != nil {
		if value, ok := ctx.Value(contextKey{}).(*zap.Logger); ok && value != nil {
			return value
		}
	}
	return getLogger()
}

// Config 日志配置
type Config struct {
	LogDir      string // 日志目录，如 ./logs
	ServiceName string // 日志基础名，如 account
	MaxAge      int    // 日志保留天数
	Level       string // 日志级别：debug/info/warn/error
	Console     bool   // 是否同时输出到控制台
}

// Sync 刷新缓存
// 在 main中 调用 defer logger.Sync()
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Init 初始化 zap 日志
func Init(cfg *Config) {
	if cfg == nil {
		log.Fatal("logger config cannot be nil")
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "./logs"
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "app"
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 7
	}
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		log.Fatalf("create log dir failed: %v", err)
	}

	// 直接写入按小时分割的文件，避免依赖符号链接权限。
	logPattern := filepath.Join(
		cfg.LogDir,
		cfg.ServiceName, // 服务名
		"%Y%m",          // 202601
		"%d",            // 13
		"%H.log",        // 14.log
	)
	writer, err := rotatelogs.New(
		logPattern,
		rotatelogs.WithRotationTime(time.Hour), // 每小时轮转
		rotatelogs.WithMaxAge(time.Duration(cfg.MaxAge)*24*time.Hour), // 保留 N 天
	)
	if err != nil {
		log.Fatalf("create rotatelogs failed: %v", err)
	}
	// 轮转器延迟打开文件，启动时主动校验写权限，避免服务运行后静默丢失文件日志。
	if _, err := writer.Write(nil); err != nil {
		_ = writer.Close()
		log.Fatalf("打开日志文件失败，请检查运行用户和日志目录权限: %v", err)
	}

	// 设置日志级别
	var level zapcore.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zap.DebugLevel
	case "warn", "warning":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "info", "":
		level = zap.InfoLevel
	default:
		log.Printf("unknown log level: %s, use info", cfg.Level)
		level = zap.InfoLevel
	}

	// 编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stack",
		EncodeTime:    zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
		LineEnding:    zapcore.DefaultLineEnding,
	}

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(writer), min(level, zapcore.InfoLevel))

	var cores []zapcore.Core
	cores = append(cores, fileCore)

	// 控制台输出
	if cfg.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), min(level, zapcore.InfoLevel))
		cores = append(cores, consoleCore)
	}

	core := &categoryCore{core: zapcore.NewTee(cores...), level: level}
	Logger = zap.New(
		core,
		zap.AddCaller(),
		// zap.Fields(
		// 	zap.String("service", cfg.Filename),
		// ),
		zap.AddCallerSkip(1),
	)
	zap.ReplaceGlobals(Logger)

	Logger.Info("Logger initialized",
		zap.String("dir", cfg.LogDir),
		zap.String("service", cfg.ServiceName),
		zap.Int("max_age_days", cfg.MaxAge),
	)
}

func getLogger() *zap.Logger {
	if Logger == nil {
		Logger, _ = zap.NewProduction()
	}
	return Logger
}

func Info(msg string, fields ...zapcore.Field) {
	getLogger().Info(msg, fields...)
}

// Error 记录错误日志
// 调用方式：logger.Error("error message", zap.String("field_name", "field_value"))
// 如果要打印堆栈信息，可以用 zap.AddStack(zap.ErrorLevel)：
//   - logger.Error("数据库写入失败", zap.Error(err), zap.Stack("stack"))
func Error(msg string, fields ...zapcore.Field) {
	getLogger().Error(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	getLogger().Warn(msg, fields...)
}

func Debug(msg string, fields ...zapcore.Field) {
	getLogger().Debug(msg, fields...)
}

// Data 拼装错误和脱敏参数，不修改调用方传入的数据。
func Data(err error, params ...any) []zap.Field {
	data := make([]zap.Field, 0, 2)
	if len(params) > 0 {
		var safe any = logdata.Hidden
		if raw, err := json.Marshal(params); err == nil {
			safe = logdata.JSON(string(raw))
		}
		data = append(data, zap.Any("req", safe))
	}
	if err != nil {
		data = append(data, zap.Error(err))
	}
	return data
}
