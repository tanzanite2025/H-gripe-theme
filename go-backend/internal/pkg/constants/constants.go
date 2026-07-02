package constants

import "time"

// File upload constants
const (
	MaxUploadSize   = 10 * 1024 * 1024  // 10MB
	MaxImageSize    = 5 * 1024 * 1024   // 5MB
	MaxVideoSize    = 100 * 1024 * 1024 // 100MB
	MaxDocumentSize = 50 * 1024 * 1024  // 50MB
)

// Pagination constants
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)

// Cache TTL constants (in seconds)
const (
	CacheTTLShort  = 5 * time.Minute
	CacheTTLMedium = 30 * time.Minute
	CacheTTLLong   = 1 * time.Hour
	CacheTTLDay    = 24 * time.Hour
)

// Rate limiting constants
const (
	RateLimitGlobal  = 1000 // requests per second
	RateLimitPerIP   = 100  // requests per second per IP
	RateLimitPerUser = 50   // requests per second per user
	RateLimitAuth    = 5    // auth attempts per minute
	RateLimitPayment = 10   // payment requests per minute
)

// User constants
const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

// Order constants
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusPaid      = "paid"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
	OrderStatusRefunded  = "refunded"
)

// Payment constants
const (
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
	PaymentStatusCancelled = "cancelled"
)

// User roles
const (
	RoleUser    = "user"
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleSupport = "support"
)

// Ticket status
const (
	TicketStatusOpen       = "open"
	TicketStatusInProgress = "in_progress"
	TicketStatusResolved   = "resolved"
	TicketStatusClosed     = "closed"
)

// Ticket priority
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Product status
const (
	ProductStatusDraft     = "draft"
	ProductStatusPublished = "published"
	ProductStatusArchived  = "archived"
)

// Context keys (for Gin context)
const (
	ContextKeyUserID    = "userID"
	ContextKeyUser      = "user"
	ContextKeyLocale    = "locale"
	ContextKeyRequestID = "requestID"
)

// Header keys
const (
	HeaderAcceptLanguage = "Accept-Language"
	HeaderContentType    = "Content-Type"
	HeaderUserAgent      = "User-Agent"
	HeaderXRequestID     = "X-Request-ID"
	HeaderXForwardedFor  = "X-Forwarded-For"
)

// Supported locales
var SupportedLocales = []string{
	"en", "zh", "zh-CN", "zh-TW",
	"fr", "de", "es", "it", "pt",
	"ja", "ko", "ru", "ar",
	"th", "vi", "id", "ms",
	"tr", "pl", "nl", "sv",
	"da", "fi", "no", "cs",
}

// Supported currencies
var SupportedCurrencies = []string{
	"USD", "EUR", "GBP", "JPY",
	"CNY", "KRW", "AUD", "CAD",
	"CHF", "SEK", "NOK", "DKK",
}

// Image formats
var SupportedImageFormats = []string{
	"jpg", "jpeg", "png", "gif",
	"webp", "svg", "ico",
}

// Video formats
var SupportedVideoFormats = []string{
	"mp4", "avi", "mov", "wmv",
	"flv", "webm", "mkv",
}

// Document formats
var SupportedDocumentFormats = []string{
	"pdf", "doc", "docx",
	"xls", "xlsx",
	"ppt", "pptx",
	"txt", "csv",
}

// Time formats
const (
	TimeFormatISO     = "2006-01-02T15:04:05Z07:00"
	TimeFormatDate    = "2006-01-02"
	TimeFormatDisplay = "2006-01-02 15:04:05"
)

// Error messages (can be used for i18n keys)
const (
	MsgInternalError    = "errors.internal_server_error"
	MsgUnauthorized     = "errors.unauthorized"
	MsgForbidden        = "errors.forbidden"
	MsgNotFound         = "errors.not_found"
	MsgValidationFailed = "errors.validation_failed"
	MsgTooManyRequests  = "errors.too_many_requests"
)

// Success messages
const (
	MsgSuccess = "success.operation_successful"
	MsgCreated = "success.created"
	MsgUpdated = "success.updated"
	MsgDeleted = "success.deleted"
)

// API versioning
const (
	APIVersion  = "v1"
	APIBasePath = "/api"
	APIFullPath = APIBasePath + "/" + APIVersion
)

// Email templates
const (
	EmailTemplateWelcome           = "welcome"
	EmailTemplatePasswordReset     = "password_reset"
	EmailTemplateOrderConfirmation = "order_confirmation"
	EmailTemplateShippingUpdate    = "shipping_update"
	EmailTemplateRefundProcessed   = "refund_processed"
)

// Regex patterns
const (
	RegexEmail    = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	RegexPhone    = `^\+?[1-9]\d{1,14}$`
	RegexURL      = `^https?://[^\s/$.?#].[^\s]*$`
	RegexSlug     = `^[a-z0-9]+(?:-[a-z0-9]+)*$`
	RegexUsername = `^[a-zA-Z0-9_-]{3,50}$`
)

// Feature flags
const (
	FeatureChatEnabled       = "chat_enabled"
	FeatureWishlistEnabled   = "wishlist_enabled"
	FeatureReviewsEnabled    = "reviews_enabled"
	FeatureLoyaltyEnabled    = "loyalty_enabled"
	FeatureNewsletterEnabled = "newsletter_enabled"
)
