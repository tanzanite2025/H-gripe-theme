package database

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	appLogger "tanzanite/internal/pkg/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormlogger "gorm.io/gorm/logger"
)

func TestParseGormLogLevelDefaultsToSilent(t *testing.T) {
	tests := map[string]gormlogger.LogLevel{
		"":        gormlogger.Silent,
		"silent":  gormlogger.Silent,
		"warn":    gormlogger.Warn,
		"error":   gormlogger.Error,
		"info":    gormlogger.Info,
		"unknown": gormlogger.Silent,
	}

	for input, expected := range tests {
		if got := parseGormLogLevel(input); got != expected {
			t.Fatalf("parseGormLogLevel(%q) = %v, want %v", input, got, expected)
		}
	}
}

func TestSafeGormLoggerDoesNotExpandSQLParams(t *testing.T) {
	var output bytes.Buffer
	previousLogger := appLogger.Log
	appLogger.Log = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&output),
		zapcore.DebugLevel,
	))
	defer func() {
		appLogger.Log = previousLogger
	}()

	logger := newGormLogger("info").(safeGormLogger)
	sql, params := logger.ParamsFilter(context.Background(), "SELECT * FROM users WHERE email = ?", "alice@example.com")
	if sql != "SELECT * FROM users WHERE email = ?" {
		t.Fatalf("unexpected filtered SQL: %s", sql)
	}
	if params != nil {
		t.Fatalf("expected params to be removed, got %#v", params)
	}

	called := false
	logger.Trace(context.Background(), time.Now(), func() (string, int64) {
		called = true
		return "SELECT * FROM users WHERE email = 'alice@example.com'", 1
	}, nil)

	if called {
		t.Fatal("Trace called SQL formatter and may expand sensitive params")
	}
	if strings.Contains(output.String(), "alice@example.com") {
		t.Fatalf("log output contains sensitive SQL parameter: %s", output.String())
	}
}
