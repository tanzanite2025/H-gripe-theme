package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	appLogger "commerce-platform/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const gormSlowSQLThreshold = 200 * time.Millisecond

type safeGormLogger struct {
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newGormLogger(level string) gormlogger.Interface {
	return safeGormLogger{
		level:         parseGormLogLevel(level),
		slowThreshold: gormSlowSQLThreshold,
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
		sqlText, rows := fc()
		appLogger.Error("database query failed",
			zap.String("source", utils.FileWithLineNum()),
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sqlText),
		)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		sqlText, rows := fc()
		appLogger.Warn("[SLOW SQL] database query exceeded threshold",
			zap.String("source", utils.FileWithLineNum()),
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.slowThreshold),
			zap.Int64("rows", rows),
			zap.String("sql", sqlText),
		)
	case l.level >= gormlogger.Info:
		sqlText, rows := fc()
		appLogger.Info("database query executed",
			zap.String("source", utils.FileWithLineNum()),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sqlText),
		)
	}
}

func (l safeGormLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	filtered := make([]interface{}, len(params))
	for index, param := range params {
		filtered[index] = sanitizeGormLogParam(param)
	}
	return sql, filtered
}

func sanitizeGormLogParam(param interface{}) interface{} {
	if param == nil {
		return nil
	}
	switch value := param.(type) {
	case sql.NamedArg:
		value.Value = sanitizeGormLogParam(value.Value)
		return value
	case time.Time:
		return "[REDACTED_TIME]"
	case *time.Time:
		if value == nil {
			return nil
		}
		return "[REDACTED_TIME]"
	case string:
		return redactedGormString(value)
	case []byte:
		return fmt.Sprintf("[REDACTED_BYTES len=%d]", len(value))
	}

	reflected := reflect.ValueOf(param)
	if reflected.Kind() == reflect.Ptr {
		if reflected.IsNil() {
			return nil
		}
		return sanitizeGormLogParam(reflected.Elem().Interface())
	}
	switch reflected.Kind() {
	case reflect.Bool:
		return param
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "[REDACTED_NUMBER]"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "[REDACTED_NUMBER]"
	case reflect.Float32, reflect.Float64:
		return "[REDACTED_NUMBER]"
	default:
		return fmt.Sprintf("[REDACTED_%s]", strings.ToUpper(reflected.Kind().String()))
	}
}

func redactedGormString(value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("[REDACTED_STRING len=%d]", len(value))
}
