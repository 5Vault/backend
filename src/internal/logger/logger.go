package logger

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/lumberjack.v2"
)

type contextKey string

const loggerKey contextKey = "logger"

var global *zap.Logger

// Init inicializa o logger global. Deve ser chamado uma vez no startup.
// Variáveis de ambiente: LOG_LEVEL, LOG_FORMAT, LOG_DIR, GIN_MODE,
// LOG_MAX_SIZE_MB, LOG_MAX_BACKUPS, LOG_MAX_AGE_DAYS, LOG_COMPRESS.
func Init() {
	level := parseLevel(getEnv("LOG_LEVEL", "info"))
	format := getEnv("LOG_FORMAT", "")
	ginMode := strings.ToLower(getEnv("GIN_MODE", "debug"))
	isProduction := ginMode == "release"

	if format == "" {
		if isProduction {
			format = "json"
		} else {
			format = "text"
		}
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     localTimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core
	cores = append(cores, buildConsoleCore(encoderCfg, format, level))
	cores = append(cores, buildFileCore(encoderCfg, level))

	combinedCore := zapcore.NewTee(cores...)

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}
	if isProduction {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		opts = append(opts, zap.Development())
	}

	global = zap.New(combinedCore, opts...)
	zap.ReplaceGlobals(global)

	global.WithOptions(zap.AddCallerSkip(-1)).Info("logger inicializado",
		zap.String("level", level.String()),
		zap.String("format", format),
		zap.String("mode", ginMode),
	)
}

func buildConsoleCore(cfg zapcore.EncoderConfig, format string, level zapcore.Level) zapcore.Core {
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(cfg)
	} else {
		colorCfg := cfg
		colorCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(colorCfg)
	}
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.NewAtomicLevelAt(level))
}

func buildFileCore(cfg zapcore.EncoderConfig, level zapcore.Level) zapcore.Core {
	logDir := getEnv("LOG_DIR", "./logs")
	maxSizeMB := getEnvInt("LOG_MAX_SIZE_MB", 100)
	maxBackups := getEnvInt("LOG_MAX_BACKUPS", 30)
	maxAgeDays := getEnvInt("LOG_MAX_AGE_DAYS", 30)
	compress := getEnvBool("LOG_COMPRESS", true)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		os.Stderr.WriteString("logger: falha ao criar diretório de logs: " + err.Error() + "\n")
		return zapcore.NewNopCore()
	}

	rotatingFile := &lumberjack.Logger{
		Filename:   logDir + "/app.log",
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
		LocalTime:  true,
	}

	jsonEncoder := zapcore.NewJSONEncoder(cfg)
	return zapcore.NewCore(jsonEncoder, zapcore.AddSync(rotatingFile), zap.NewAtomicLevelAt(level))
}

func localTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000-07:00"))
}

// L retorna o logger global.
func L() *zap.Logger { return global }

// With retorna um logger filho com campos fixos.
func With(fields ...zap.Field) *zap.Logger { return global.With(fields...) }

func Debug(msg string, fields ...zap.Field) { global.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { global.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { global.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { global.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { global.Fatal(msg, fields...) }

// Sync faz flush dos buffers. Chamar no shutdown.
func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}

// WithContext embute o logger no contexto.
func WithContext(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// FromContext extrai o logger do contexto, ou retorna o global.
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return global
	}
	if log, ok := ctx.Value(loggerKey).(*zap.Logger); ok && log != nil {
		return log
	}
	return global
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(s) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
