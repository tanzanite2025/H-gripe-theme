package app

import (
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
	FAQ                *service.FAQService
	Gallery            *service.GalleryService
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
}

func NewDependencies(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) *Dependencies {
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

	storageSvc, _ := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: "./uploads", BaseURL: cfg.Server.BaseURL})

	services := Services{
		Auth:         service.NewAuthService(repos.User, cfg.JWT, cfg.OAuth),
		Post:         service.NewPostService(repos.Post, redisCache, cfg.Cache.PostTTL),
		Product:      service.NewProductService(repos.Product, redisCache, cfg.Cache.ProductTTL),
		Cart:         service.NewCartService(repos.Cart, repos.Product),
		Setting:      service.NewSettingService(repos.Setting, redisCache, cfg.Cache.SettingsTTL),
		FAQ:          service.NewFAQService(repos.FAQ),
		Gallery:      service.NewGalleryService(repos.Gallery),
		Registration: service.NewRegistrationService(repos.Registration, repos.Product, repos.Order),
		Checkout:     service.NewCheckoutService(repos.Product, repos.Coupon, repos.Payment, repos.Loyalty),
		Marketing:    service.NewMarketingService(repos.Coupon, repos.Loyalty),
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
		Shipping:  service.NewShippingService(repos.Shipping),
	}
	services.AdminSettings = service.NewAdminSettingsService(services.Setting, repos.User)
	services.Order = service.NewOrderService(db, repos.Order, repos.Product, repos.Coupon, services.Checkout, repos.Shipping, repos.Audit, repos.Loyalty)
	services.Payment = service.NewPaymentService(repos.Payment, services.Order)

	return &Dependencies{
		Repositories: repos,
		Services:     services,
		Storage:      storageSvc,
	}
}
