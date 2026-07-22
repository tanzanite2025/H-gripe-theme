package app

import (
	"fmt"
	"os"

	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"gorm.io/gorm"
)

type Dependencies struct {
	Repositories Repositories
	Services     Services
	Storage      storage.StorageService
}

type Repositories struct {
	User               *repository.UserRepository
	Post               *repository.PostRepository
	Product            *repository.ProductRepository
	Cart               *repository.CartRepository
	Setting            *repository.SettingRepository
	FAQ                *repository.FAQRepository
	Order              *repository.OrderRepository
	Payment            *repository.PaymentRepository
	Shipping           *repository.ShippingRepository
	Coupon             *repository.CouponRepository
	Loyalty            *repository.LoyaltyRepository
	Review             *repository.ReviewRepository
	Ticket             *repository.TicketRepository
	Gallery            *repository.GalleryRepository
	Media              *repository.MediaRepository
	Registration       *repository.RegistrationRepository
	Audit              *repository.AuditRepository
	Showcase           *repository.ShowcaseRepository
	Wishlist           *repository.WishlistRepository
	Feedback           *repository.FeedbackRepository
	SuggestionFeedback *repository.SuggestionFeedbackRepository
	Spoke              *repository.SpokeRepository
	Chat               *repository.ChatRepository
	Subscription       *repository.SubscriptionRepository
}

type Services struct {
	Auth               *service.AuthService
	Post               *service.PostService
	Product            *service.ProductService
	Cart               *service.CartService
	Setting            *service.SettingService
	AdminSettings      *service.AdminSettingsService
	AdminPublicChat    *service.AdminPublicChatAgentService
	FAQ                *service.FAQService
	Gallery            *service.GalleryService
	Media              *service.MediaService
	Registration       *service.RegistrationService
	Checkout           *service.CheckoutService
	Order              *service.OrderService
	Payment            *service.PaymentService
	Marketing          *service.MarketingService
	Review             *service.ReviewService
	Ticket             *service.TicketService
	Subscription       *service.SubscriptionService
	Sitemap            *service.SitemapService
	Showcase           *service.ShowcaseService
	Wishlist           *service.WishlistService
	Feedback           *service.FeedbackService
	SuggestionFeedback *service.SuggestionFeedbackService
	User               *service.UserService
	Dashboard          *service.DashboardService
	Audit              *service.AuditService
	Shipping           *service.ShippingService
	Spoke              *service.SpokeService
	Chat               *service.ChatService
}

func NewDependencies(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) (*Dependencies, error) {
	repos := Repositories{
		User:               repository.NewUserRepository(db),
		Post:               repository.NewPostRepository(db),
		Product:            repository.NewProductRepository(db),
		Cart:               repository.NewCartRepository(db),
		Setting:            repository.NewSettingRepository(db),
		FAQ:                repository.NewFAQRepository(db),
		Order:              repository.NewOrderRepository(db),
		Payment:            repository.NewPaymentRepository(db),
		Shipping:           repository.NewShippingRepository(db),
		Coupon:             repository.NewCouponRepository(db),
		Loyalty:            repository.NewLoyaltyRepository(db),
		Review:             repository.NewReviewRepository(db),
		Ticket:             repository.NewTicketRepository(db),
		Gallery:            repository.NewGalleryRepository(db),
		Media:              repository.NewMediaRepository(db),
		Registration:       repository.NewRegistrationRepository(db),
		Audit:              repository.NewAuditRepository(db),
		Showcase:           repository.NewShowcaseRepository(db),
		Wishlist:           repository.NewWishlistRepository(db),
		Feedback:           repository.NewFeedbackRepository(db),
		SuggestionFeedback: repository.NewSuggestionFeedbackRepository(db),
		Spoke:              repository.NewSpokeRepository(db),
		Chat:               repository.NewChatRepository(db),
		Subscription:       repository.NewSubscriptionRepository(db),
	}

	storageConfig := storage.LoadConfigFromEnv()
	if _, configured := os.LookupEnv("STORAGE_BASE_URL"); !configured {
		storageConfig.BaseURL = cfg.Server.BaseURL
	}
	storageSvc, err := storage.NewStorageService(storageConfig)
	if err != nil {
		return nil, fmt.Errorf("initialize storage: %w", err)
	}
	txManager := repository.NewTxManager(db, repos.Order, repos.Product, repos.Coupon, repos.Loyalty, repos.Payment, repos.Shipping)

	shippingService := service.NewShippingService(repos.Shipping, repos.Product)

	services := Services{
		Auth:         service.NewAuthService(repos.User, cfg.JWT, cfg.OAuth),
		Post:         service.NewPostService(repos.Post, redisCache, cfg.Cache.PostTTL),
		Product:      service.NewProductService(repos.Product, redisCache, cfg.Cache.ProductTTL),
		Cart:         service.NewCartService(repos.Cart, repos.Product),
		Setting:      service.NewSettingService(repos.Setting, redisCache, cfg.Cache.SettingsTTL),
		FAQ:          service.NewFAQService(repos.FAQ),
		Gallery:      service.NewGalleryService(repos.Gallery),
		Media:        service.NewMediaService(repos.Media, storageSvc),
		Registration: service.NewRegistrationService(repos.Registration, repos.Product, repos.Order),
		Checkout:     service.NewCheckoutService(repos.Product, repos.Coupon, repos.Payment, repos.Loyalty, shippingService),
		Marketing:    service.NewMarketingService(txManager, repos.Coupon, repos.Loyalty),
		Review:       service.NewReviewService(repos.Review),
		Ticket:       service.NewTicketService(repos.Ticket, repos.User),
		Subscription: service.NewSubscriptionService(repos.Subscription),
		Sitemap:      service.NewSitemapService(repos.Post, cfg.Server.BaseURL),
		Showcase:     service.NewShowcaseService(repos.Showcase, storageSvc),
		Wishlist:     service.NewWishlistService(repos.Wishlist, repos.Product),
		Feedback:     service.NewFeedbackService(repos.Feedback),
		SuggestionFeedback: service.NewSuggestionFeedbackService(
			repos.SuggestionFeedback,
		),
		User:      service.NewUserService(repos.User),
		Dashboard: service.NewDashboardService(repos.Order, repos.User, repos.Ticket, repos.Subscription),
		Audit:     service.NewAuditService(repos.Audit),
		Shipping:  shippingService,
		Spoke:     service.NewSpokeService(repos.Spoke),
		Chat:      service.NewChatService(repos.Chat),
	}
	services.AdminSettings = service.NewAdminSettingsService(services.Setting)
	services.AdminPublicChat = service.NewAdminPublicChatAgentService(repos.User)
	services.Order = service.NewOrderService(
		txManager,
		repos.Order,
		services.Checkout,
	)
	services.Payment = service.NewPaymentService(txManager, repos.Payment)

	return &Dependencies{
		Repositories: repos,
		Services:     services,
		Storage:      storageSvc,
	}, nil
}
