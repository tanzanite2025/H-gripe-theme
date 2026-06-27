package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	appLogger "tanzanite/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type safeGormLogger struct {
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newGormLogger(level string) gormlogger.Interface {
	return safeGormLogger{
		level:         parseGormLogLevel(level),
		slowThreshold: time.Second,
	}
}

func parseGormLogLevel(level string) gormlogger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "silent", "off", "none":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn", "warning":
		return gormlogger.Warn
	case "info", "debug":
		return gormlogger.Info
	default:
		return gormlogger.Silent
	}
}

func (l safeGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l.level = level
	return l
}

func (l safeGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		appLogger.Info("gorm info", zap.String("message", msg))
	}
}

func (l safeGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		appLogger.Warn("gorm warning", zap.String("message", msg))
	}
}

func (l safeGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		appLogger.Error("gorm error", zap.String("message", msg))
	}
}

func (l safeGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.level >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		appLogger.Error("database query failed",
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.Duration("elapsed", elapsed),
		)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		appLogger.Warn("slow database query",
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.slowThreshold),
		)
	case l.level >= gormlogger.Info:
		appLogger.Info("database query executed", zap.Duration("elapsed", elapsed))
	}
}

func (l safeGormLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	return sql, nil
}
