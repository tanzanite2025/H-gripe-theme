package database

import (
	"context"
	"fmt"
	"time"

	"tanzanite/internal/pkg/config"
	appLogger "tanzanite/internal/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// Init 初始化数据库连接
func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.GetDSN())
	case "mysql":
		dialector = mysql.Open(cfg.GetDSN())
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: newGormLogger(cfg.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure read/write splitting using dbresolver
	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{dialector},
		Replicas: []gorm.Dialector{dialector}, // Simulate replica using same DSN
		Policy:   dbresolver.RandomPolicy{},
	}).
		SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second).
		SetMaxIdleConns(cfg.MaxIdleConns).
		SetMaxOpenConns(cfg.MaxOpenConns))

	if err != nil {
		return nil, fmt.Errorf("failed to configure dbresolver: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 测试连接
	if err := sqlDB.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	appLogger.Info("database connected successfully")
	return db, nil
}

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is nil")
	}
	return Init(*cfg)
}
