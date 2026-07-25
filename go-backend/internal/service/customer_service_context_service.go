package service

import (
	"strconv"
	"strings"
	"time"

	"tanzanite/internal/domain/order"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/domain/wishlist"
	"tanzanite/internal/repository"
)

type CustomerServiceContextService struct {
	ticketService         *TicketService
	userRepo              *repository.UserRepository
	cartRepo              *repository.CartRepository
	wishlistRepo          *repository.WishlistRepository
	orderRepo             *repository.OrderRepository
	visitorProfileService *VisitorProfileService
}

func NewCustomerServiceContextService(
	ticketService *TicketService,
	userRepo *repository.UserRepository,
	cartRepo *repository.CartRepository,
	wishlistRepo *repository.WishlistRepository,
	orderRepo *repository.OrderRepository,
	visitorProfileService *VisitorProfileService,
) *CustomerServiceContextService {
	return &CustomerServiceContextService{
		ticketService:         ticketService,
		userRepo:              userRepo,
		cartRepo:              cartRepo,
		wishlistRepo:          wishlistRepo,
		orderRepo:             orderRepo,
		visitorProfileService: visitorProfileService,
	}
}

type CustomerServiceContext struct {
	Conversation CustomerServiceContextConversation `json:"conversation"`
	Customer     CustomerServiceContextCustomer     `json:"customer"`
	Contact      CustomerServiceContextContact      `json:"contact"`
	Cart         CustomerServiceContextCart         `json:"cart"`
	Wishlist     CustomerServiceContextWishlist     `json:"wishlist"`
	Orders       CustomerServiceContextOrders       `json:"orders"`
	Browsing     CustomerServiceContextBrowsing     `json:"browsing"`
	Signals      CustomerServiceContextSignals      `json:"signals"`
}

type CustomerServiceContextConversation struct {
	ID                 uint       `json:"id"`
	ConversationID     string     `json:"conversation_id"`
	TicketNumber       string     `json:"ticket_number"`
	Status             string     `json:"status"`
	AssignedTo         uint       `json:"assigned_to"`
	CustomerUserID     *uint      `json:"customer_user_id,omitempty"`
	VisitorAnonymous   bool       `json:"visitor_anonymous"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastCustomerSeenAt *time.Time `json:"last_customer_seen_at,omitempty"`
}

type CustomerServiceContextCustomer struct {
	Type            string                                `json:"type"`
	Account         *CustomerServiceContextAccount        `json:"account,omitempty"`
	Anonymous       *CustomerServiceContextAnonymous      `json:"anonymous,omitempty"`
	IdentitySources []CustomerServiceContextIdentityClaim `json:"identity_sources"`
}

type CustomerServiceContextAccount struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Role        string    `json:"role"`
	Locale      string    `json:"locale"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type CustomerServiceContextAnonymous struct {
	VisitorSessionBound bool   `json:"visitor_session_bound"`
	VisitorProfileID    uint   `json:"visitor_profile_id,omitempty"`
	VisitorHashPreview  string `json:"visitor_hash_preview,omitempty"`
	Note                string `json:"note"`
}

type CustomerServiceContextIdentityClaim struct {
	Source string `json:"source"`
	Value  string `json:"value"`
	Status string `json:"status"`
}

type CustomerServiceContextContact struct {
	Email        string `json:"email"`
	EmailSource  string `json:"email_source"`
	Locale       string `json:"locale"`
	LocaleSource string `json:"locale_source"`
}

type CustomerServiceContextCart struct {
	Available bool                             `json:"available"`
	Reason    string                           `json:"reason,omitempty"`
	ItemCount int                              `json:"item_count"`
	Total     float64                          `json:"total"`
	Items     []CustomerServiceContextCartItem `json:"items"`
}

type CustomerServiceContextCartItem struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	VariantID   *uint   `json:"variant_id,omitempty"`
	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Image       string  `json:"image"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	LineTotal   float64 `json:"line_total"`
	VariantName string  `json:"variant_name"`
}

type CustomerServiceContextWishlist struct {
	Available bool                                 `json:"available"`
	Reason    string                               `json:"reason,omitempty"`
	Count     int                                  `json:"count"`
	Items     []CustomerServiceContextWishlistItem `json:"items"`
}

type CustomerServiceContextWishlistItem struct {
	ID        uint      `json:"id"`
	ProductID uint      `json:"product_id"`
	Name      string    `json:"name"`
	SKU       string    `json:"sku"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"created_at"`
}

type CustomerServiceContextOrders struct {
	Available bool                              `json:"available"`
	Reason    string                            `json:"reason,omitempty"`
	Total     int64                             `json:"total"`
	Items     []CustomerServiceContextOrderItem `json:"items"`
}

type CustomerServiceContextOrderItem struct {
	ID             uint      `json:"id"`
	OrderNumber    string    `json:"order_number"`
	Status         string    `json:"status"`
	PaymentStatus  string    `json:"payment_status"`
	ShippingStatus string    `json:"shipping_status"`
	TotalAmount    float64   `json:"total_amount"`
	ItemCount      int       `json:"item_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type CustomerServiceContextBrowsing struct {
	Available bool                                 `json:"available"`
	Reason    string                               `json:"reason,omitempty"`
	Count     int                                  `json:"count"`
	Items     []CustomerServiceContextBrowsingItem `json:"items"`
}

type CustomerServiceContextBrowsingItem struct {
	ProductID    uint      `json:"product_id"`
	ViewCount    int       `json:"view_count"`
	LastViewedAt time.Time `json:"last_viewed_at"`
}

type CustomerServiceContextSignals struct {
	Region         CustomerServiceContextSignal `json:"region"`
	CartSession    CustomerServiceContextSignal `json:"cart_session"`
	EmailCapture   CustomerServiceContextSignal `json:"email_capture"`
	VisitorProfile CustomerServiceContextSignal `json:"visitor_profile"`
}

type CustomerServiceContextSignal struct {
	Status string `json:"status"`
	Value  string `json:"value,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func (s *CustomerServiceContextService) GetConversationContextForAgent(ticketID, agentUserID uint, canViewAll bool) (*CustomerServiceContext, error) {
	t, err := s.ticketService.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll)
	if err != nil {
		return nil, err
	}

	context := &CustomerServiceContext{
		Conversation: CustomerServiceContextConversation{
			ID:               t.ID,
			ConversationID:   ticketConversationID(t),
			TicketNumber:     t.TicketNumber,
			Status:           t.Status,
			AssignedTo:       t.AssignedTo,
			CustomerUserID:   t.CustomerUserID,
			VisitorAnonymous: t.CustomerUserID == nil,
			CreatedAt:        t.CreatedAt,
			UpdatedAt:        t.UpdatedAt,
		},
		Customer: CustomerServiceContextCustomer{
			Type: "anonymous",
			Anonymous: &CustomerServiceContextAnonymous{
				VisitorSessionBound: strings.TrimSpace(t.VisitorSessionHash) != "",
				VisitorHashPreview:  hashPreview(t.VisitorSessionHash),
				Note:                "匿名访客只绑定 Public Chat visitor cookie；暂未和购物车 session、邮箱、地区建立统一访客档案。",
			},
			IdentitySources: []CustomerServiceContextIdentityClaim{
				{Source: "customer_user_id", Value: "", Status: "missing"},
				{Source: "visitor_session_hash", Value: hashPreview(t.VisitorSessionHash), Status: availabilityStatus(t.VisitorSessionHash)},
			},
		},
		Contact: CustomerServiceContextContact{
			EmailSource:  "not_captured",
			LocaleSource: "not_captured",
		},
		Cart: CustomerServiceContextCart{
			Available: false,
			Reason:    "匿名访客聊天 cookie 尚未与购物车 session 统一绑定。",
			Items:     []CustomerServiceContextCartItem{},
		},
		Wishlist: CustomerServiceContextWishlist{
			Available: false,
			Reason:    "心愿单只对登录账号可用。",
			Items:     []CustomerServiceContextWishlistItem{},
		},
		Orders: CustomerServiceContextOrders{
			Available: false,
			Reason:    "订单只对登录账号可用。",
			Items:     []CustomerServiceContextOrderItem{},
		},
		Browsing: CustomerServiceContextBrowsing{
			Available: false,
			Reason:    "浏览历史只对登录账号可用。",
			Items:     []CustomerServiceContextBrowsingItem{},
		},
		Signals: CustomerServiceContextSignals{
			Region:         CustomerServiceContextSignal{Status: "not_captured", Reason: "尚未建立 visitor profile / GeoIP 采集层。"},
			CartSession:    CustomerServiceContextSignal{Status: "not_linked", Reason: "Public Chat visitor cookie 和 cart session 仍是两套标识。"},
			EmailCapture:   CustomerServiceContextSignal{Status: "not_captured", Reason: "匿名访客邮箱不能猜测，只能来自登录、订单、订阅或主动填写。"},
			VisitorProfile: CustomerServiceContextSignal{Status: "not_created", Reason: "下一阶段应新增统一 visitor profile。"},
		},
	}

	if t.CustomerUserID == nil {
		s.applyAnonymousVisitorProfile(context, t.VisitorSessionHash)
		return context, nil
	}

	customerUserID := *t.CustomerUserID
	account, err := s.userRepo.FindByID(customerUserID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			context.Customer.IdentitySources[0] = CustomerServiceContextIdentityClaim{
				Source: "customer_user_id",
				Value:  strconv.FormatUint(uint64(customerUserID), 10),
				Status: "missing_user",
			}
			return context, nil
		}
		return nil, err
	}

	context.Customer = CustomerServiceContextCustomer{
		Type: "account",
		Account: &CustomerServiceContextAccount{
			ID:          account.ID,
			Email:       account.Email,
			Username:    account.Username,
			DisplayName: serviceDisplayName(account),
			FirstName:   account.FirstName,
			LastName:    account.LastName,
			Role:        account.Role,
			Locale:      account.Locale,
			Status:      account.Status,
			CreatedAt:   account.CreatedAt,
		},
		IdentitySources: []CustomerServiceContextIdentityClaim{
			{Source: "customer_user_id", Value: strconv.FormatUint(uint64(customerUserID), 10), Status: "verified"},
			{Source: "visitor_session_hash", Value: hashPreview(t.VisitorSessionHash), Status: availabilityStatus(t.VisitorSessionHash)},
		},
	}
	context.Contact = CustomerServiceContextContact{
		Email:        account.Email,
		EmailSource:  "account",
		Locale:       account.Locale,
		LocaleSource: "account",
	}
	context.Signals.EmailCapture = CustomerServiceContextSignal{Status: "captured", Value: account.Email, Reason: "来自登录账号。"}

	context.Cart = s.customerCartContext(customerUserID)
	context.Wishlist = s.customerWishlistContext(customerUserID)
	context.Orders = s.customerOrdersContext(customerUserID)
	context.Browsing = s.customerBrowsingContext(customerUserID)
	s.applyVisitorProfileSignals(context, t.VisitorSessionHash)

	return context, nil
}

func (s *CustomerServiceContextService) applyAnonymousVisitorProfile(context *CustomerServiceContext, visitorSessionHash string) {
	if context == nil || s.visitorProfileService == nil || strings.TrimSpace(visitorSessionHash) == "" {
		return
	}

	profile, err := s.visitorProfileService.FindByCustomerServiceVisitorHash(visitorSessionHash)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return
		}
		context.Signals.VisitorProfile = CustomerServiceContextSignal{
			Status: "error",
			Reason: err.Error(),
		}
		return
	}
	if profile == nil {
		return
	}

	if context.Customer.Anonymous != nil {
		context.Customer.Anonymous.VisitorProfileID = profile.ID
		context.Customer.Anonymous.Note = "匿名访客已建立 visitor profile；仅展示已采集事实，不展示推断信息。"
	}

	if strings.TrimSpace(profile.Email) != "" {
		context.Contact.Email = profile.Email
		context.Contact.EmailSource = profile.EmailSource
		context.Signals.EmailCapture = CustomerServiceContextSignal{
			Status: "captured",
			Value:  profile.Email,
			Reason: "来自 visitor profile 的主动采集字段。",
		}
	}
	if strings.TrimSpace(profile.Locale) != "" {
		context.Contact.Locale = profile.Locale
		context.Contact.LocaleSource = profile.LocaleSource
	}

	if strings.TrimSpace(profile.CartSessionID) != "" {
		context.Cart = s.customerCartContextBySessionID(profile.CartSessionID)
		context.Signals.CartSession = CustomerServiceContextSignal{
			Status: "linked",
			Value:  hashPreview(profile.CartSessionID),
			Reason: "Public Chat visitor cookie 已通过 visitor profile 绑定购物车 session。",
		}
	}

	applyProfileLocationSignals(context, profile.ID, profile.CountryCode, profile.Region, profile.City)
}

func (s *CustomerServiceContextService) applyVisitorProfileSignals(context *CustomerServiceContext, visitorSessionHash string) {
	if context == nil || s.visitorProfileService == nil || strings.TrimSpace(visitorSessionHash) == "" {
		return
	}
	profile, err := s.visitorProfileService.FindByCustomerServiceVisitorHash(visitorSessionHash)
	if err != nil || profile == nil {
		return
	}

	if strings.TrimSpace(profile.CartSessionID) != "" {
		context.Signals.CartSession = CustomerServiceContextSignal{
			Status: "linked",
			Value:  hashPreview(profile.CartSessionID),
			Reason: "当前访问档案已绑定购物车 session；登录用户购物车仍以 user_id 为准。",
		}
	}
	applyProfileLocationSignals(context, profile.ID, profile.CountryCode, profile.Region, profile.City)
}

func applyProfileLocationSignals(context *CustomerServiceContext, profileID uint, countryCode, region, city string) {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(countryCode) != "" {
		parts = append(parts, strings.TrimSpace(countryCode))
	}
	if strings.TrimSpace(region) != "" {
		parts = append(parts, strings.TrimSpace(region))
	}
	if strings.TrimSpace(city) != "" {
		parts = append(parts, strings.TrimSpace(city))
	}
	if len(parts) > 0 {
		context.Signals.Region = CustomerServiceContextSignal{
			Status: "captured",
			Value:  strings.Join(parts, " / "),
			Reason: "来自 visitor profile 的粗略地区字段。",
		}
	}
	context.Signals.VisitorProfile = CustomerServiceContextSignal{
		Status: "created",
		Value:  strconv.FormatUint(uint64(profileID), 10),
		Reason: "已建立统一访客档案。",
	}
}

func (s *CustomerServiceContextService) customerCartContext(userID uint) CustomerServiceContextCart {
	result := CustomerServiceContextCart{Available: true, Items: []CustomerServiceContextCartItem{}}
	cart, err := s.cartRepo.FindByUserID(userID)
	if repository.IsRecordNotFound(err) {
		return result
	}
	if err != nil {
		result.Available = false
		result.Reason = err.Error()
		return result
	}
	summary, err := s.cartRepo.GetSummary(cart.ID)
	if err != nil {
		result.Available = false
		result.Reason = err.Error()
		return result
	}
	result.ItemCount = summary.ItemCount
	result.Total = summary.Total
	result.Items = customerCartItems(summary.Items)
	return result
}

func (s *CustomerServiceContextService) customerCartContextBySessionID(sessionID string) CustomerServiceContextCart {
	result := CustomerServiceContextCart{Available: true, Items: []CustomerServiceContextCartItem{}}
	cart, err := s.cartRepo.FindBySessionID(strings.TrimSpace(sessionID))
	if repository.IsRecordNotFound(err) {
		result.Available = false
		result.Reason = "visitor profile 已绑定购物车 session，但当前 session 没有购物车记录。"
		return result
	}
	if err != nil {
		result.Available = false
		result.Reason = err.Error()
		return result
	}
	summary, err := s.cartRepo.GetSummary(cart.ID)
	if err != nil {
		result.Available = false
		result.Reason = err.Error()
		return result
	}
	result.ItemCount = summary.ItemCount
	result.Total = summary.Total
	result.Items = customerCartItems(summary.Items)
	return result
}

func (s *CustomerServiceContextService) customerWishlistContext(userID uint) CustomerServiceContextWishlist {
	result := CustomerServiceContextWishlist{Available: true, Items: []CustomerServiceContextWishlistItem{}}
	items, err := s.wishlistRepo.ListByUserID(userID)
	if err != nil {
		result.Available = false
		result.Reason = err.Error()
		return result
	}
	result.Count = len(items)
	result.Items = customerWishlistItems(items, 8)
	return result
}

func (s *CustomerServiceContextService) customerOrdersContext(userID uint) CustomerServiceContextOrders {
	result := CustomerServiceContextOrders{Available: true, Items: []CustomerServiceContextOrderItem{}}
	orders, total, err := s.orderRepo.FindByUserID(userID, 1, 5)
	if err != nil {
		result.Available = false
		result.Reason = err.Error()
		return result
	}
	result.Total = total
	result.Items = customerOrderItems(orders)
	return result
}

func (s *CustomerServiceContextService) customerBrowsingContext(userID uint) CustomerServiceContextBrowsing {
	result := CustomerServiceContextBrowsing{Available: true, Items: []CustomerServiceContextBrowsingItem{}}
	items, err := s.userRepo.GetBrowsingHistory(userID, 8)
	if err != nil {
		result.Available = false
		result.Reason = err.Error()
		return result
	}
	result.Count = len(items)
	result.Items = customerBrowsingItems(items)
	return result
}

func customerCartItems(items []product.CartItem) []CustomerServiceContextCartItem {
	result := make([]CustomerServiceContextCartItem, 0, len(items))
	for _, item := range items {
		name := "Unknown product"
		sku := ""
		image := ""
		if item.Product != nil {
			name = item.Product.Name
			sku = item.Product.DisplaySKU()
			image = firstProductImage(item.Product)
		}
		variantName := ""
		if item.Variant != nil {
			if strings.TrimSpace(item.Variant.SKU) != "" {
				sku = item.Variant.SKU
			}
			variantName = strings.TrimSpace(item.Variant.OptionValues)
		}
		result = append(result, CustomerServiceContextCartItem{
			ID:          item.ID,
			ProductID:   item.ProductID,
			VariantID:   item.VariantID,
			Name:        name,
			SKU:         sku,
			Image:       image,
			Quantity:    item.Quantity,
			Price:       item.Price,
			LineTotal:   item.Price * float64(item.Quantity),
			VariantName: variantName,
		})
	}
	return result
}

func customerWishlistItems(items []wishlist.Item, limit int) []CustomerServiceContextWishlistItem {
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	result := make([]CustomerServiceContextWishlistItem, 0, len(items))
	for _, item := range items {
		name := "Unknown product"
		sku := ""
		image := ""
		if item.Product != nil {
			name = item.Product.Name
			sku = item.Product.DisplaySKU()
			image = firstProductImage(item.Product)
		}
		result = append(result, CustomerServiceContextWishlistItem{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      name,
			SKU:       sku,
			Image:     image,
			CreatedAt: item.CreatedAt,
		})
	}
	return result
}

func customerOrderItems(items []order.Order) []CustomerServiceContextOrderItem {
	result := make([]CustomerServiceContextOrderItem, 0, len(items))
	for _, item := range items {
		result = append(result, CustomerServiceContextOrderItem{
			ID:             item.ID,
			OrderNumber:    item.OrderNumber,
			Status:         item.Status,
			PaymentStatus:  item.PaymentStatus,
			ShippingStatus: item.ShippingStatus,
			TotalAmount:    item.TotalAmount,
			ItemCount:      len(item.Items),
			CreatedAt:      item.CreatedAt,
		})
	}
	return result
}

func customerBrowsingItems(items []user.BrowsingHistory) []CustomerServiceContextBrowsingItem {
	result := make([]CustomerServiceContextBrowsingItem, 0, len(items))
	for _, item := range items {
		result = append(result, CustomerServiceContextBrowsingItem{
			ProductID:    item.ProductID,
			ViewCount:    item.ViewCount,
			LastViewedAt: item.LastViewedAt,
		})
	}
	return result
}

func firstProductImage(item *product.Product) string {
	if item == nil {
		return ""
	}
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible && strings.TrimSpace(media.URL) != "" {
			return media.URL
		}
	}
	return ""
}

func serviceDisplayName(item *user.User) string {
	if item == nil {
		return ""
	}
	fullName := strings.TrimSpace(strings.TrimSpace(item.FirstName) + " " + strings.TrimSpace(item.LastName))
	if fullName != "" {
		return fullName
	}
	if strings.TrimSpace(item.Username) != "" {
		return strings.TrimSpace(item.Username)
	}
	return strings.TrimSpace(item.Email)
}

func availabilityStatus(value string) string {
	if strings.TrimSpace(value) == "" {
		return "missing"
	}
	return "bound"
}

func hashPreview(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 12 {
		return value
	}
	return value[:6] + "..." + value[len(value)-6:]
}
