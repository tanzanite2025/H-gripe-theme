package database

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	appLogger "commerce-platform/internal/pkg/logger"

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
	if len(params) != 1 || params[0] == "alice@example.com" {
		t.Fatalf("expected the parameter to be retained in a redacted form, got %#v", params)
	}

	called := false
	logger.Trace(context.Background(), time.Now().Add(-logger.slowThreshold-time.Millisecond), func() (string, int64) {
		called = true
		return "SELECT * FROM users WHERE email = '[REDACTED_STRING len=17]'", 1
	}, nil)

	if !called {
		t.Fatal("Trace did not collect the SQL for the slow-query log")
	}
	if !strings.Contains(output.String(), "[SLOW SQL]") {
		t.Fatalf("slow-query marker is missing from log output: %s", output.String())
	}
	if !strings.Contains(output.String(), "rows") {
		t.Fatalf("row count is missing from log output: %s", output.String())
	}
	if strings.Contains(output.String(), "alice@example.com") {
		t.Fatalf("log output contains sensitive SQL parameter: %s", output.String())
	}
}

func TestNewGormLoggerUsesExplicitSlowQueryThreshold(t *testing.T) {
	logger := newGormLogger("warn").(safeGormLogger)
	if logger.slowThreshold != 200*time.Millisecond {
		t.Fatalf("slow query threshold = %s, want 200ms", logger.slowThreshold)
	}
}
