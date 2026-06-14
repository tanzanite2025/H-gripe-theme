package database

import (
	"tanzanite/internal/domain/audit"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/faq"
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
	"tanzanite/internal/domain/subscription"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/domain/user"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},
		&post.Post{},
		&post.Category{},
		&post.PostCategory{},
		&product.Product{},
		&product.ProductImage{},
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
		&shipping.TrackingEvent{},
		&shipping.ShippingZone{},
		&coupon.Coupon{},
		&coupon.CouponUsage{},
		&coupon.GiftCard{},
		&coupon.GiftCardTransaction{},
		&loyalty.LoyaltyTransaction{},
		&loyalty.CheckIn{},
		&loyalty.Referral{},
		&loyalty.MemberLevel{},
		&loyalty.UserLoyalty{},
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
		&subscription.Subscription{},
		&showcase.Showcase{},
		&showcase.Comment{},
		&media.Media{},
		&audit.AuditLog{},
	)
}
