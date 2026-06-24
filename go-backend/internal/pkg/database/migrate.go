package database

import (
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
	"tanzanite/internal/domain/wishlist"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&user.User{},
		&user.AgentProfile{},
		&post.Post{},
		&post.Category{},
		&post.PostCategory{},
		&product.Product{},
		&product.ProductImage{},
		&product.ProductAttribute{},
		&product.AttributeValue{},
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
		&shipping.PackagingRule{},
		&shipping.PackagingRuleApply{},
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
		&ticket.AutoReplyRule{},
		&subscription.Subscription{},
		&showcase.Showcase{},
		&showcase.Comment{},
		&media.Media{},
		&audit.AuditLog{},
		&wishlist.Item{},
		&feedback.Feedback{},
		&suggestionfeedback.SuggestionFeedback{},
		&spoke.History{},
	)
	if err != nil {
		return err
	}
	return SeedDefaultSettings(db)
}

// SeedDefaultSettings 种子数据初始化
func SeedDefaultSettings(db *gorm.DB) error {
	defaultSettings := []setting.Setting{
		{Key: "tz_redeem_enabled", Value: "1", Type: "boolean", Group: "redeem", Locale: "en", IsPublic: true, Description: "Whether point redemption is enabled"},
		{Key: "tz_redeem_enabled", Value: "1", Type: "boolean", Group: "redeem", Locale: "zh", IsPublic: true, Description: "是否启用积分兑换"},
		{Key: "tz_redeem_exchange_rate", Value: "100", Type: "number", Group: "redeem", Locale: "en", IsPublic: true, Description: "Redemption exchange rate (e.g. 100 points = 1 unit)"},
		{Key: "tz_redeem_exchange_rate", Value: "100", Type: "number", Group: "redeem", Locale: "zh", IsPublic: true, Description: "积分兑换比例（如100积分=1元）"},
		{Key: "tz_redeem_min_points", Value: "1000", Type: "number", Group: "redeem", Locale: "en", IsPublic: true, Description: "Minimum points required to redeem"},
		{Key: "tz_redeem_min_points", Value: "1000", Type: "number", Group: "redeem", Locale: "zh", IsPublic: true, Description: "兑换所需最小积分"},
		{Key: "tz_redeem_max_value_per_day", Value: "500", Type: "number", Group: "redeem", Locale: "en", IsPublic: true, Description: "Maximum value redeemable per day"},
		{Key: "tz_redeem_max_value_per_day", Value: "500", Type: "number", Group: "redeem", Locale: "zh", IsPublic: true, Description: "每日最大可兑换价值"},
		{Key: "tz_redeem_card_expiry_days", Value: "365", Type: "number", Group: "redeem", Locale: "en", IsPublic: true, Description: "Redeemed gift card expiry days"},
		{Key: "tz_redeem_card_expiry_days", Value: "365", Type: "number", Group: "redeem", Locale: "zh", IsPublic: true, Description: "兑换出的礼品卡有效期天数"},
		{Key: "tz_redeem_preset_values", Value: "10,50,100,200,500", Type: "string", Group: "redeem", Locale: "en", IsPublic: true, Description: "Preset gift card values for redemption"},
		{Key: "tz_redeem_preset_values", Value: "10,50,100,200,500", Type: "string", Group: "redeem", Locale: "zh", IsPublic: true, Description: "预设的可兑换礼品卡面额"},
	}

	for _, s := range defaultSettings {
		var count int64
		if err := db.Model(&setting.Setting{}).Where("`key` = ? AND locale = ?", s.Key, s.Locale).Count(&count).Error; err != nil {
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

