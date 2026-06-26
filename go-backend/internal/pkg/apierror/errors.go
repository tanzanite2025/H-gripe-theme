package apierror

import "errors"

// Domain errors - 定义所有业务错误

// User errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrWeakPassword       = errors.New("password is too weak")
	ErrPasswordMismatch   = errors.New("password confirmation does not match")
)

// Product errors
var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidProductID  = errors.New("invalid product ID")
	ErrProductDisabled   = errors.New("product is disabled")
	ErrInvalidPrice      = errors.New("invalid price")
	ErrInvalidQuantity   = errors.New("invalid quantity")
)

// Order errors
var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderCancelled     = errors.New("order has been cancelled")
	ErrOrderCompleted     = errors.New("order has been completed")
	ErrInvalidOrderStatus = errors.New("invalid order status")
	ErrCannotCancelOrder  = errors.New("order cannot be cancelled")
	ErrEmptyCart          = errors.New("cart is empty")
)

// Payment errors
var (
	ErrPaymentFailed        = errors.New("payment failed")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrPaymentExpired       = errors.New("payment has expired")
	ErrRefundFailed         = errors.New("refund failed")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
)

// Cart errors
var (
	ErrCartNotFound        = errors.New("cart not found")
	ErrCartItemNotFound    = errors.New("cart item not found")
	ErrInvalidCartQuantity = errors.New("invalid cart quantity")
	ErrCartEmpty           = errors.New("cart is empty")
)

// Wishlist errors
var (
	ErrWishlistNotFound        = errors.New("wishlist not found")
	ErrWishlistItemNotFound    = errors.New("wishlist item not found")
	ErrWishlistItemExists      = errors.New("item already in wishlist")
	ErrWishlistProductNotFound = errors.New("product not found for wishlist")
)

// Coupon errors
var (
	ErrCouponNotFound          = errors.New("coupon not found")
	ErrCouponExpired           = errors.New("coupon has expired")
	ErrCouponInactive          = errors.New("coupon is inactive")
	ErrCouponUsageLimitReached = errors.New("coupon usage limit reached")
	ErrInvalidCouponCode       = errors.New("invalid coupon code")
	ErrCouponMinAmountNotMet   = errors.New("minimum amount requirement not met")
)

// Shipping errors
var (
	ErrShippingNotFound    = errors.New("shipping method not found")
	ErrInvalidAddress      = errors.New("invalid shipping address")
	ErrShippingUnavailable = errors.New("shipping unavailable for this location")
)

// Post/Content errors
var (
	ErrPostNotFound = errors.New("post not found")
	ErrInvalidSlug  = errors.New("invalid slug")
	ErrSlugExists   = errors.New("slug already exists")
)

// Setting errors
var (
	ErrSettingNotFound     = errors.New("setting not found")
	ErrInvalidSettingValue = errors.New("invalid setting value")
)

// Ticket errors
var (
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrTicketClosed       = errors.New("ticket is closed")
	ErrUnauthorizedTicket = errors.New("unauthorized to access this ticket")
)

// Feedback errors
var (
	ErrFeedbackNotFound      = errors.New("feedback not found")
	ErrFeedbackMissingThread = errors.New("thread key is required")
)

// FAQ errors
var (
	ErrFAQNotFound     = errors.New("FAQ not found")
	ErrInvalidCategory = errors.New("invalid FAQ category")
)

// Gallery errors
var (
	ErrGalleryNotFound      = errors.New("gallery not found")
	ErrGalleryImageNotFound = errors.New("gallery image not found")
)

// Subscription errors
var (
	ErrSubscriptionNotFound    = errors.New("subscription not found")
	ErrEmailAlreadySubscribed  = errors.New("email already subscribed")
	ErrInvalidUnsubscribeToken = errors.New("invalid unsubscribe token")
)

// Registration errors
var (
	ErrRegistrationNotFound = errors.New("registration not found")
	ErrInvalidSerialNumber  = errors.New("invalid serial number")
	ErrSerialNumberUsed     = errors.New("serial number already used")
	ErrWarrantyExpired      = errors.New("warranty has expired")
)

// Auth/Permission errors
var (
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token expired")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)

// Upload/File errors
var (
	ErrInvalidFile         = errors.New("invalid file")
	ErrFileTooLarge        = errors.New("file too large")
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrUploadFailed        = errors.New("file upload failed")
)

// Validation errors
var (
	ErrValidationFailed     = errors.New("validation failed")
	ErrInvalidInput         = errors.New("invalid input")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidFormat        = errors.New("invalid format")
)

// Database errors
var (
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseQuery      = errors.New("database query error")
	ErrRecordNotFound     = errors.New("record not found")
	ErrDuplicateEntry     = errors.New("duplicate entry")
)

// Cache errors
var (
	ErrCacheNotFound   = errors.New("cache entry not found")
	ErrCacheConnection = errors.New("cache connection error")
)

// Generic errors
var (
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrTimeout            = errors.New("request timeout")
	ErrNotImplemented     = errors.New("not implemented")
)
