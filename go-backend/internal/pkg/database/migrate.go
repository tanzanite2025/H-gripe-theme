package database

import (
	"commerce-platform/internal/domain/aftersales"
	attributiondomain "commerce-platform/internal/domain/attribution"
	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/domain/feedback"
	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/loyalty"
	marketdomain "commerce-platform/internal/domain/market"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/merchant"
	"commerce-platform/internal/domain/ops"
	orderdomain "commerce-platform/internal/domain/order"
	outboxdomain "commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/post"
	preflightdomain "commerce-platform/internal/domain/preflight"
	procurementdomain "commerce-platform/internal/domain/procurement"
	"commerce-platform/internal/domain/product"
	recommendationdomain "commerce-platform/internal/domain/recommendation"
	"commerce-platform/internal/domain/review"
	selectionconfigurationdomain "commerce-platform/internal/domain/selectionconfiguration"
	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/domain/showcase"
	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/domain/social"
	"commerce-platform/internal/domain/spoke"
	"commerce-platform/internal/domain/subscription"
	"commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/domain/ticket"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/domain/verification"
	"commerce-platform/internal/domain/visitor"
	visualshowcasedomain "commerce-platform/internal/domain/visualshowcase"
	"commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/domain/wishlist"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"context"
	"database/sql"
	"fmt"

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
		&user.AgentGroup{},
		&user.AgentProfile{},
		&user.AgentGroupMember{},
		&post.Post{},
		&post.Category{},
		&post.PostCategory{},
		&product.ProductInformationTemplate{},
		&product.CustomsClassificationProfile{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductAttribute{},
		&product.AttributeValue{},
		&product.ProductSpecificationTemplate{},
		&product.ProductCategory{},
		&product.ProductCategoryTranslation{},
		&product.SpecDefinition{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&product.ProductVariantOptionValue{},
		&product.Cart{},
		&product.CartItem{},
		&procurementdomain.ProductProcurement{},
		&procurementdomain.ProductProfitCalculation{},
		&fitmentcatalogdomain.FrameFitmentEntry{},
		&fitmentcatalogdomain.HubSpecification{},
		&fitmentcatalogdomain.FrameHubSpecification{},
		&merchant.GoogleMerchantConnection{},
		&merchant.GoogleMerchantOffer{},
		&social.OAuthConnection{},
		&social.OAuthSession{},
		&ops.DomainBinding{},
		&ops.Connector{},
		&ops.VPSBinding{},
		&ops.ProjectBinding{},
		&ops.NetworkRule{},
		&marketdomain.StorefrontMarket{},
		&marketdomain.MarketCountry{},
		&orderdomain.Order{},
		&orderdomain.OrderItem{},
		&orderdomain.OrderIdempotency{},
		&orderdomain.PolicyDisclosure{},
		&aftersales.AfterSalesCase{},
		&aftersales.AfterSalesCaseItem{},
		&aftersales.AfterSalesCaseEvent{},
		&aftersales.AfterSalesCaseEventArchive{},
		&aftersales.AfterSalesCaseAttachment{},
		&aftersales.AfterSalesRefundReview{},
		&attributiondomain.OrderAttribution{},
		&outboxdomain.Event{},
		&payment.PaymentMethod{},
		&payment.TaxRate{},
		&payment.Transaction{},
		&payment.Refund{},
		&payment.RefundLineItem{},
		&payment.StripeWebhookEvent{},
		&payment.StripeDispute{},
		&payment.PaymentReview{},
		&payment.PaymentRiskEvent{},
		&payment.PaymentRiskSnapshot{},
		&payment.PaymentRiskAlertState{},
		&payment.PaymentRiskCheckoutDecision{},
		&payment.PaymentProtectionControl{},
		&payment.PaymentRefundRecommendation{},
		&currency.ExchangeRate{},
		&currency.ExchangeRateSyncLease{},
		&shipping.ShippingTemplate{},
		&shipping.ShippingRule{},
		&shipping.Carrier{},
		&shipping.CarrierService{},
		&shipping.TrackingProviderConfig{},
		&shipping.TrackingCarrierMapping{},
		&shipping.TrackingShipment{},
		&shipping.TrackingEvent{},
		&shipping.ShipmentRecord{},
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
		&warranty.WarrantyClaim{},
		&warranty.WarrantyServiceRecord{},
		&review.Review{},
		&review.ReviewHelpful{},
		&review.ReviewSummary{},
		&selectionconfigurationdomain.SelectionConfigurationKey{},
		&setting.Setting{},
		&ticket.Ticket{},
		&ticket.TicketMessage{},
		&ticket.CustomerServiceInboxState{},
		&ticket.AutoReplyRule{},
		&visitor.Profile{},
		&visitor.RiskDailyFact{},
		&visitor.RiskDecision{},
		&subscription.Subscription{},
		&verification.EmailChallenge{},
		&showcase.Showcase{},
		&showcase.Comment{},
		&media.Media{},
		&media.MediaAsset{},
		&media.MediaAssetDerivative{},
		&media.MediaDerivativePreset{},
		&media.MediaDerivativeRebuildJob{},
		&audit.AuditLog{},
		&wishlist.Item{},
		&feedback.Feedback{},
		&suggestionfeedback.SuggestionFeedback{},
		&spoke.CatalogRimBrand{},
		&spoke.CatalogRimModel{},
		&spoke.CatalogHubBrand{},
		&spoke.CatalogHubModel{},
		&spoke.CatalogBuildPreset{},
		&spoke.History{},
		&recommendationdomain.Event{},
		&seodomain.StorefrontRouteCatalogEntry{},
		&seodomain.StorefrontRouteCheckResult{},
		&urlmanagementdomain.StorefrontRedirectRule{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
		&preflightdomain.ContentLinkRun{},
		&preflightdomain.ContentLinkIssue{},
		&preflightdomain.ContentLinkIssueEvent{},
		&visualshowcasedomain.Item{},
		&sitequalitydomain.SiteQualityTarget{},
		&sitequalitydomain.SiteQualityJob{},
		&sitequalitydomain.SiteQualityProviderSlot{},
		&sitequalitydomain.SiteQualityEvaluation{},
		&sitequalitydomain.SiteQualityRun{},
		&sitequalitydomain.SiteQualityRunArchive{},
		&sitequalitydomain.SiteQualityFinding{},
		&sitequalitydomain.SiteQualityFindingEvent{},
	)
	if err != nil {
		return err
	}
	if err := migrateVisitorRiskFactIndexes(db); err != nil {
		return err
	}
	if err := SeedDefaultSettings(db); err != nil {
		return err
	}
	return nil
}

func migrateVisitorRiskFactIndexes(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	statements := []string{
		`DROP INDEX IF EXISTS uk_visitor_risk_daily_fact`,
		`DROP INDEX IF EXISTS uk_visitor_risk_daily_fact_device`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_visitor_risk_daily_fact ON visitor_risk_daily_facts(day, ip_hash, user_agent_hash) WHERE device_fingerprint_hash = ''`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_visitor_risk_daily_fact_device ON visitor_risk_daily_facts(day, device_fingerprint_hash) WHERE device_fingerprint_hash <> ''`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
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
		logger.Info("empty database detected; applying SQL migration baseline")
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

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	logger.Info("SQL migrations completed successfully")
	return nil
}
