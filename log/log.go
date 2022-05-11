package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

var zapLogger *zap.Logger

// GetZapLogger get cached zap logger, cache was created when call logger generation func in `*/lognex/log` package
func GetZapLogger() *zap.Logger {
	return zapLogger
}

// Sugar wraps the Logger to provide a more ergonomic, but slightly slower,
// API. Sugaring a Logger is quite inexpensive, so it's reasonable for a
// single application to use both Loggers and SugaredLoggers, converting
// between them on the boundaries of performance-sensitive code.
func Sugar() *zap.SugaredLogger {
	return zapLogger.Sugar()
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	zapLogger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	zapLogger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	zapLogger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	zapLogger.Error(msg, fields...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(msg string, fields ...zap.Field) {
	zapLogger.DPanic(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...zap.Field) {
	zapLogger.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, fields ...zap.Field) {
	zapLogger.Fatal(msg, fields...)
}

//===========================================================================

type ZapTeeConfig struct {
	LevelEnablerFunc zap.LevelEnablerFunc
	Writer           io.Writer
	UseJsonEncoder   bool
}

func RecommendLogger() *zap.Logger {
	return RecommendLoggerWithLogPath("/data/logs")
}

func RecommendLoggerWithLogPath(logPath string) *zap.Logger {
	teeConfigs := []ZapTeeConfig{
		{
			LevelEnablerFunc: func(level zapcore.Level) bool {
				return level >= zap.InfoLevel
			},
			Writer: &lumberjack.Logger{
				Filename:   logPath + "/info.log",
				MaxSize:    50, //MB
				MaxBackups: 3,
				MaxAge:     2, //days
			},
			UseJsonEncoder: true,
		},
		{
			LevelEnablerFunc: func(level zapcore.Level) bool {
				return level >= zap.WarnLevel
			},
			Writer: &lumberjack.Logger{
				Filename:   logPath + "/error.log",
				MaxSize:    50,
				MaxAge:     7,
				MaxBackups: 5,
			},
			UseJsonEncoder: false,
		},
	}
	return MultiCoreLogger(teeConfigs, false,
		zap.AddCaller(), zap.AddCallerSkip(0), zap.AddStacktrace(zap.WarnLevel))
}

func MultiCoreLogger(teeConfigs []ZapTeeConfig, removeDevLogger bool, opts ...zap.Option) *zap.Logger {
	var cores []zapcore.Core
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	for _, tee := range teeConfigs {
		syncer := zapcore.AddSync(tee.Writer)
		var encoder zapcore.Encoder
		if tee.UseJsonEncoder {
			encoder = zapcore.NewJSONEncoder(encoderCfg)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderCfg)
		}
		core := zapcore.NewCore(encoder, syncer, tee.LevelEnablerFunc)
		cores = append(cores, core)
	}

	if !removeDevLogger {
		devEncoder := zap.NewDevelopmentEncoderConfig()
		devEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder

		devCore := zapcore.NewCore(zapcore.NewConsoleEncoder(devEncoder), zapcore.Lock(os.Stdout), zap.DebugLevel)
		cores = append(cores, devCore)
	}

	logger := zap.New(zapcore.NewTee(cores...), opts...)
	zapLogger = logger
	return zapLogger
}
