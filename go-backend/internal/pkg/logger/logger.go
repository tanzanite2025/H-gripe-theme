package logger

import (
	"tanzanite/internal/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init 初始化日志
func Init(cfg config.LogConfig) error {
	var zapConfig zap.Config

	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// 设置输出
	if cfg.Output != "" && cfg.Output != "stdout" {
		zapConfig.OutputPaths = []string{cfg.Output}
	}

	Log, err = zapConfig.Build()
	if err != nil {
		return err
	}

	return nil
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Fatal 记录致命错误并退出
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}
