package database

import (
	"context"
	"database/sql"
	"fmt"
	"tanzanite/internal/domain/audit"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/domain/feedback"
	"tanzanite/internal/domain/gallery"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/media"
	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/domain/payment"
	"tanzanite/internal/domain/post"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/domain/registration"
	"tanzanite/internal/domain/review"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/domain/showcase"
	"tanzanite/internal/domain/spoke"
	"tanzanite/internal/domain/subscription"
	"tanzanite/internal/domain/suggestionfeedback"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/domain/verification"
	"tanzanite/internal/domain/visitor"
	"tanzanite/internal/domain/wishlist"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // register file source driver
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, serverMode string) error {
	if serverMode == "release" {
		logger.Warn("GORM AutoMigrate is running in release mode. It is recommended to use SQL migrations for production. Proceeding with caution.")
	}

	err := db.AutoMigrate(
		&user.User{},
		&user.AgentProfile{},
		&post.Post{},
		&post.Category{},
		&post.PostCategory{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductAttribute{},
		&product.AttributeValue{},
		&product.ProductType{},
		&product.SpecDefinition{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&product.Cart{},
		&product.CartItem{},
		&orderdomain.Order{},
		&orderdomain.OrderItem{},
		&payment.PaymentMethod{},
		&payment.TaxRate{},
		&payment.Transaction{},
		&payment.Refund{},
		&shipping.ShippingTemplate{},
		&shipping.ShippingRule{},
		&shipping.Carrier{},
		&shipping.CarrierService{},
		&shipping.TrackingProviderConfig{},
		&shipping.TrackingCarrierMapping{},
		&shipping.TrackingShipment{},
		&shipping.TrackingEvent{},
		&shipping.ShippingZone{},
		&shipping.PackagingRule{},
		&shipping.PackagingRuleApply{},
		&coupon.Coupon{},
		&coupon.CouponUsage{},
		&coupon.GiftCard{},
		&coupon.GiftCardTransaction{},
		&coupon.GiftCardRedemption{},
		&loyalty.LoyaltyTransaction{},
		&loyalty.ProgramConfig{},
		&loyalty.ProgramRedeemOption{},
		&loyalty.CheckIn{},
		&loyalty.Referral{},
		&loyalty.MemberLevel{},
		&loyalty.UserLoyalty{},
		&faq.FAQPage{},
		&faq.FAQCategory{},
		&faq.FAQ{},
		&gallery.Gallery{},
		&gallery.GalleryImage{},
		&registration.ProductRegistration{},
		&registration.WarrantyClaim{},
		&review.Review{},
		&review.ReviewHelpful{},
		&setting.Setting{},
		&ticket.Ticket{},
		&ticket.TicketMessage{},
		&ticket.AutoReplyRule{},
		&visitor.Profile{},
		&subscription.Subscription{},
		&verification.EmailChallenge{},
		&showcase.Showcase{},
		&showcase.Comment{},
		&media.Media{},
		&media.MediaAsset{},
		&audit.AuditLog{},
		&wishlist.Item{},
		&feedback.Feedback{},
		&suggestionfeedback.SuggestionFeedback{},
		&spoke.History{},
	)
	if err != nil {
		return err
	}
	if err := SeedDefaultSettings(db); err != nil {
		return err
	}
	return SeedDefaultLoyaltyProgramConfig(db)
}

// SeedDefaultSettings 种子数据初始化
func SeedDefaultSettings(db *gorm.DB) error {
	defaultSettings := []setting.Setting{}

	for _, s := range defaultSettings {
		var count int64
		if err := db.Model(&setting.Setting{}).Where("key = ? AND locale = ?", s.Key, s.Locale).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&s).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func SeedDefaultLoyaltyProgramConfig(db *gorm.DB) error {
	var count int64
	if err := db.Model(&loyalty.ProgramConfig{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	config := loyalty.DefaultProgramConfig()
	config.RedeemOptions = []loyalty.ProgramRedeemOption{
		{ValueCents: 1000, SortOrder: 0},
		{ValueCents: 5000, SortOrder: 1},
		{ValueCents: 10000, SortOrder: 2},
		{ValueCents: 20000, SortOrder: 3},
		{ValueCents: 50000, SortOrder: 4},
	}
	return db.Create(&config).Error
}

func PrepareSchema(ctx context.Context, db *gorm.DB, cfg *config.DatabaseConfig, serverMode string) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}
	if cfg == nil {
		return fmt.Errorf("database config is nil")
	}
	if cfg.Driver != "postgres" {
		return fmt.Errorf("schema preparation requires postgres, got %q", cfg.Driver)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql database: %w", err)
	}

	var applicationTableCount int
	if err := sqlDB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		  AND table_name <> 'schema_migrations'
	`).Scan(&applicationTableCount); err != nil {
		return fmt.Errorf("inspect public schema: %w", err)
	}

	if applicationTableCount == 0 {
		logger.Info("empty database detected; creating current schema baseline")
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return AutoMigrate(tx, serverMode)
		}); err != nil {
			return fmt.Errorf("create schema baseline: %w", err)
		}
	} else {
		logger.Info("existing database detected; skipping GORM schema baseline",
			zap.Int("application_table_count", applicationTableCount),
		)
	}

	if err := RunSQLMigrations(sqlDB, cfg); err != nil {
		return fmt.Errorf("run SQL migrations: %w", err)
	}
	return nil
}

// RunSQLMigrations executes SQL migrations using golang-migrate
func RunSQLMigrations(sqlDB *sql.DB, cfg *config.DatabaseConfig) error {
	if cfg.Driver != "postgres" {
		logger.Info("SQL Migrations only implemented for postgres currently, skipping", zap.String("driver", cfg.Driver))
		return nil
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not instantiate migrate: %w", err)
	}

	if err := repairKnownDirtySQLMigration(m); err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	logger.Info("SQL migrations completed successfully")
	return nil
}

func repairKnownDirtySQLMigration(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect SQL migration version: %w", err)
	}
	if !dirty {
		return nil
	}
	if version != 44 {
		return fmt.Errorf("SQL migration version %d is dirty; manual repair is required", version)
	}

	logger.Warn("Repairing dirty SQL migration 44 before replaying idempotent loyalty migration")
	if err := m.Force(43); err != nil {
		return fmt.Errorf("force dirty SQL migration 44 back to version 43: %w", err)
	}
	return nil
}
