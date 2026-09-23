package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/product"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/domain/wishlist"
	"commerce-platform/internal/repository"
)

type CustomerServiceContextService struct {
	ticketService         *TicketService
	userRepo              *repository.UserRepository
	cartRepo              *repository.CartRepository
	wishlistRepo          *repository.WishlistRepository
	orderRepo             *repository.OrderRepository
	productRepo           *repository.ProductRepository
	shippingRepo          *repository.ShippingRepository
	afterSalesRepo        *repository.AfterSalesCaseRepository
	paymentRepo           *repository.PaymentRepository
	warrantyRepo          *repository.WarrantyRepository
	auditService          *AuditService
	loyaltyRepo           *repository.LoyaltyRepository
	visitorProfileService *VisitorProfileService
	mediaURLResolver      PublicMediaURLResolver
}

// ConfigureFactRepositories wires the optional operational fact sources used
// by the customer-service context. Keeping these sources optional preserves
// the base context when a deployment has not enabled after-sales or payment
// dispute persistence yet.
func (s *CustomerServiceContextService) ConfigureFactRepositories(
	afterSalesRepo *repository.AfterSalesCaseRepository,
	paymentRepo *repository.PaymentRepository,
) {
	if s == nil {
		return
	}
	s.afterSalesRepo = afterSalesRepo
	s.paymentRepo = paymentRepo
}

// ConfigureFulfillmentRepositories wires optional read-only projections used
// by the customer-service order card. They remain optional so older test and
// development databases can still serve the base context.
func (s *CustomerServiceContextService) ConfigureFulfillmentRepositories(
	productRepo *repository.ProductRepository,
	shippingRepo *repository.ShippingRepository,
	warrantyRepo *repository.WarrantyRepository,
) {
	if s == nil {
		return
	}
	s.productRepo = productRepo
	s.shippingRepo = shippingRepo
	s.warrantyRepo = warrantyRepo
}

// ConfigureAuditService enables an append-only read audit for every context
// request. Audit failures never expose a sensitive payload and never erase a
// previously readable context snapshot.
func (s *CustomerServiceContextService) ConfigureAuditService(auditService *AuditService) {
	if s == nil {
		return
	}
	s.auditService = auditService
}

func NewCustomerServiceContextService(
	ticketService *TicketService,
	userRepo *repository.UserRepository,
	cartRepo *repository.CartRepository,
	wishlistRepo *repository.WishlistRepository,
	orderRepo *repository.OrderRepository,
	loyaltyRepo *repository.LoyaltyRepository,
	visitorProfileService *VisitorProfileService,
) *CustomerServiceContextService {
	return &CustomerServiceContextService{
		ticketService:         ticketService,
		userRepo:              userRepo,
		cartRepo:              cartRepo,
		wishlistRepo:          wishlistRepo,
		orderRepo:             orderRepo,
		loyaltyRepo:           loyaltyRepo,
		visitorProfileService: visitorProfileService,
	}
}

func (s *CustomerServiceContextService) ConfigureMediaService(resolver PublicMediaURLResolver) {
	if s == nil {
		return
	}
	s.mediaURLResolver = resolver
}

type CustomerServiceContext struct {
	Conversation    CustomerServiceContextConversation    `json:"conversation"`
	Customer        CustomerServiceContextCustomer        `json:"customer"`
	Contact         CustomerServiceContextContact         `json:"contact"`
	Cart            CustomerServiceContextCart            `json:"cart"`
	Wishlist        CustomerServiceContextWishlist        `json:"wishlist"`
	Orders          CustomerServiceContextOrders          `json:"orders"`
	ShippingAddress CustomerServiceContextShippingAddress `json:"shipping_address"`
	AfterSales      CustomerServiceContextAfterSales      `json:"after_sales"`
	Disputes        CustomerServiceContextDisputes        `json:"payment_disputes"`
	Browsing        CustomerServiceContextBrowsing        `json:"browsing"`
	Signals         CustomerServiceContextSignals         `json:"signals"`
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
	ID          uint                       `json:"id"`
	Email       string                     `json:"email"`
	Username    string                     `json:"username"`
	DisplayName string                     `json:"display_name"`
	FirstName   string                     `json:"first_name"`
	LastName    string                     `json:"last_name"`
	Role        string                     `json:"role"`
	Locale      string                     `json:"locale"`
	Status      string                     `json:"status"`
	MemberTier  *CustomerServiceMemberTier `json:"member_tier,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
}

type CustomerServiceMemberTier struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Icon            string `json:"icon,omitempty"`
	Color           string `json:"color,omitempty"`
	TotalPoints     int    `json:"total_points"`
	AvailablePoints int    `json:"available_points"`
}

type CustomerServiceConversationSummary struct {
	Type          string                     `json:"type"`
	Identity      string                     `json:"identity"`
	IdentityLabel string                     `json:"identity_label"`
	DisplayName   string                     `json:"display_name"`
	RegionLabel   string                     `json:"region_label"`
	RegionStatus  string                     `json:"region_status"`
	MemberTier    *CustomerServiceMemberTier `json:"member_tier,omitempty"`
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
	Email          string `json:"email"`
	EmailSource    string `json:"email_source"`
	Locale         string `json:"locale"`
	LocaleSource   string `json:"locale_source"`
	Timezone       string `json:"timezone"`
	TimezoneSource string `json:"timezone_source"`
}

type CustomerServiceContextCart struct {
	Available bool                             `json:"available"`
	Status    string                           `json:"status"`
	Reason    string                           `json:"reason,omitempty"`
	ItemCount int                              `json:"item_count"`
	Currency  string                           `json:"currency"`
	Total     string                           `json:"total"`
	Items     []CustomerServiceContextCartItem `json:"items"`
}

type CustomerServiceContextCartItem struct {
	ID                uint   `json:"id"`
	ProductID         uint   `json:"product_id"`
	VariantID         *uint  `json:"variant_id,omitempty"`
	Name              string `json:"name"`
	SKU               string `json:"sku"`
	Image             string `json:"image"`
	Quantity          int    `json:"quantity"`
	Currency          string `json:"currency"`
	Price             string `json:"price"`
	LineTotal         string `json:"line_total"`
	InventorySnapshot string `json:"inventory_snapshot"`
	VariantName       string `json:"variant_name"`
}

type CustomerServiceContextWishlist struct {
	Available bool                                 `json:"available"`
	Status    string                               `json:"status"`
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
	Available         bool                              `json:"available"`
	Status            string                            `json:"status"`
	Reason            string                            `json:"reason,omitempty"`
	Total             int64                             `json:"total"`
	FulfillmentStatus string                            `json:"fulfillment_status"`
	FulfillmentReason string                            `json:"fulfillment_reason,omitempty"`
	Items             []CustomerServiceContextOrderItem `json:"items"`
}

type CustomerServiceContextOrderItem struct {
	ID                       uint                              `json:"id"`
	OrderNumber              string                            `json:"order_number"`
	Status                   string                            `json:"status"`
	PaymentStatus            string                            `json:"payment_status"`
	ShippingStatus           string                            `json:"shipping_status"`
	Currency                 string                            `json:"currency"`
	SubtotalAmount           string                            `json:"subtotal_amount"`
	DiscountAmount           string                            `json:"discount_amount"`
	DiscountedSubtotalAmount string                            `json:"discounted_subtotal_amount"`
	ShippingFee              string                            `json:"shipping_fee"`
	TaxAmount                string                            `json:"tax_amount"`
	TotalAmount              string                            `json:"total_amount"`
	ItemCount                int                               `json:"item_count"`
	Items                    []CustomerServiceContextOrderLine `json:"items"`
	Shipments                []CustomerServiceContextShipment  `json:"shipments"`
	CreatedAt                time.Time                         `json:"created_at"`
}

type CustomerServiceContextOrderLine struct {
	ID                uint   `json:"id"`
	ProductID         uint   `json:"product_id"`
	VariantID         *uint  `json:"variant_id,omitempty"`
	Name              string `json:"name"`
	SKU               string `json:"sku,omitempty"`
	Thumbnail         string `json:"thumbnail,omitempty"`
	Quantity          int    `json:"quantity"`
	Currency          string `json:"currency"`
	UnitPrice         string `json:"unit_price"`
	SubtotalAmount    string `json:"subtotal_amount"`
	DiscountAmount    string `json:"discount_amount"`
	TotalAmount       string `json:"total_amount"`
	FulfillmentMode   string `json:"fulfillment_mode,omitempty"`
	InventorySnapshot string `json:"inventory_snapshot"`
	PricingSnapshot   string `json:"pricing_snapshot"`
}

type CustomerServiceContextShipment struct {
	ID                 uint       `json:"id"`
	Carrier            string     `json:"carrier,omitempty"`
	CarrierService     string     `json:"carrier_service,omitempty"`
	TrackingNumber     string     `json:"tracking_number"`
	RegistrationStatus string     `json:"registration_status,omitempty"`
	SyncStatus         string     `json:"sync_status,omitempty"`
	LastEventAt        *time.Time `json:"last_event_at,omitempty"`
}

type CustomerServiceContextShippingAddress struct {
	Available         bool   `json:"available"`
	Status            string `json:"status"`
	Reason            string `json:"reason,omitempty"`
	SourceOrderID     uint   `json:"source_order_id,omitempty"`
	SourceOrderNumber string `json:"source_order_number,omitempty"`
	RecipientName     string `json:"recipient_name,omitempty"`
	AddressLine       string `json:"address_line,omitempty"`
	City              string `json:"city,omitempty"`
	State             string `json:"state,omitempty"`
	PostalCode        string `json:"postal_code,omitempty"`
	Country           string `json:"country,omitempty"`
	PhonePresent      bool   `json:"phone_present"`
}

type CustomerServiceContextAfterSales struct {
	Available      bool                                   `json:"available"`
	Status         string                                 `json:"status"`
	Reason         string                                 `json:"reason,omitempty"`
	Items          []CustomerServiceContextAfterSalesItem `json:"items"`
	RefundStatus   string                                 `json:"refund_status"`
	RefundReason   string                                 `json:"refund_reason,omitempty"`
	Refunds        []CustomerServiceContextRefund         `json:"refunds"`
	WarrantyStatus string                                 `json:"warranty_status"`
	WarrantyReason string                                 `json:"warranty_reason,omitempty"`
	Warranty       []CustomerServiceContextWarrantyClaim  `json:"warranty_claims"`
}

type CustomerServiceContextAfterSalesItem struct {
	ID                 uint      `json:"id"`
	OrderID            uint      `json:"order_id"`
	OrderNumber        string    `json:"order_number,omitempty"`
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	ItemSummary        string    `json:"item_summary,omitempty"`
	CurrentHandlerID   uint      `json:"current_handler_id,omitempty"`
	CurrentHandlerName string    `json:"current_handler_name,omitempty"`
	Resolution         string    `json:"resolution,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CustomerServiceContextRefund struct {
	ID          uint       `json:"id"`
	OrderID     uint       `json:"order_id"`
	OrderNumber string     `json:"order_number,omitempty"`
	Status      string     `json:"status"`
	Reason      string     `json:"reason,omitempty"`
	Amount      string     `json:"amount"`
	Currency    string     `json:"currency"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type CustomerServiceContextWarrantyClaim struct {
	ID           uint      `json:"id"`
	OrderItemID  *uint     `json:"order_item_id,omitempty"`
	OrderNumber  string    `json:"order_number,omitempty"`
	IssueType    string    `json:"issue_type"`
	Status       string    `json:"status"`
	TirePressure string    `json:"tire_pressure,omitempty"`
	IsTubeless   bool      `json:"is_tubeless"`
	Resolution   string    `json:"resolution,omitempty"`
	ProcessedBy  uint      `json:"processed_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CustomerServiceContextDisputes struct {
	Available bool                                `json:"available"`
	Status    string                              `json:"status"`
	Reason    string                              `json:"reason,omitempty"`
	Items     []CustomerServiceContextDisputeItem `json:"items"`
}

type CustomerServiceContextDisputeItem struct {
	Provider            string     `json:"provider"`
	OrderID             uint       `json:"order_id"`
	OrderNumber         string     `json:"order_number,omitempty"`
	Status              string     `json:"status"`
	Reason              string     `json:"reason,omitempty"`
	Amount              string     `json:"amount"`
	Currency            string     `json:"currency"`
	EvidenceDueAt       *time.Time `json:"evidence_due_at,omitempty"`
	EvidenceSubmittedAt *time.Time `json:"evidence_submitted_at,omitempty"`
}

func paymentDisputeDisplayAmount(amountMinor int64, currencyCode string) string {
	amount, err := domainmoney.New(amountMinor, currencyCode)
	if err != nil {
		return "0"
	}
	value, err := amount.FormatMajor()
	if err != nil {
		return "0"
	}
	return value
}

type CustomerServiceContextBrowsing struct {
	Available bool                                 `json:"available"`
	Reason    string                               `json:"reason,omitempty"`
	Count     int                                  `json:"count"`
	Items     []CustomerServiceContextBrowsingItem `json:"items"`
}

type CustomerServiceContextBrowsingItem struct {
	ProductID    uint      `json:"product_id"`
	Name         string    `json:"name,omitempty"`
	SKU          string    `json:"sku,omitempty"`
	Thumbnail    string    `json:"thumbnail,omitempty"`
	Currency     string    `json:"currency,omitempty"`
	Price        string    `json:"price,omitempty"`
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
	defer s.auditContextRead(ticketID, agentUserID, canViewAll)

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
			EmailSource:    "not_captured",
			LocaleSource:   "not_captured",
			TimezoneSource: "not_captured",
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
			Available:         false,
			Reason:            "订单只对登录账号可用。",
			FulfillmentStatus: "unavailable",
			Items:             []CustomerServiceContextOrderItem{},
		},
		ShippingAddress: CustomerServiceContextShippingAddress{
			Available: false,
			Status:    "unavailable",
			Reason:    "收货地址只对登录账号的订单事实可用。",
		},
		AfterSales: CustomerServiceContextAfterSales{
			Available:    false,
			Status:       "unavailable",
			Reason:       "售后事实只对登录账号的订单可用。",
			Items:        []CustomerServiceContextAfterSalesItem{},
			RefundStatus: "unavailable", WarrantyStatus: "unavailable",
			Refunds:  []CustomerServiceContextRefund{},
			Warranty: []CustomerServiceContextWarrantyClaim{},
		},
		Disputes: CustomerServiceContextDisputes{
			Available: false,
			Status:    "unavailable",
			Reason:    "支付争议只对登录账号的订单可用。",
			Items:     []CustomerServiceContextDisputeItem{},
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
			s.applyVisitorTimezone(context, t.VisitorSessionHash, customerUserID)
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
			MemberTier:  s.memberTierSummary(customerUserID),
			CreatedAt:   account.CreatedAt,
		},
		IdentitySources: []CustomerServiceContextIdentityClaim{
			{Source: "customer_user_id", Value: strconv.FormatUint(uint64(customerUserID), 10), Status: "verified"},
			{Source: "visitor_session_hash", Value: hashPreview(t.VisitorSessionHash), Status: availabilityStatus(t.VisitorSessionHash)},
		},
	}
	context.Contact = CustomerServiceContextContact{
		Email:          account.Email,
		EmailSource:    "account",
		Locale:         account.Locale,
		LocaleSource:   "account",
		TimezoneSource: "not_captured",
	}
	context.Signals.EmailCapture = CustomerServiceContextSignal{Status: "captured", Value: account.Email, Reason: "来自登录账号。"}
	s.applyVisitorTimezone(context, t.VisitorSessionHash, customerUserID)

	context.Cart = s.customerCartContext(customerUserID)
	context.Wishlist = s.customerWishlistContext(customerUserID)
	context.Orders, context.ShippingAddress, context.AfterSales, context.Disputes = s.customerOperationalFactsContext(customerUserID)
	context.Browsing = s.customerBrowsingContext(customerUserID)
	s.applyVisitorProfileSignals(context, t.VisitorSessionHash)

	return context, nil
}

func (s *CustomerServiceContextService) auditContextRead(ticketID, agentUserID uint, canViewAll bool) {
	if s == nil || s.auditService == nil || ticketID == 0 {
		return
	}
	scope := "assigned"
	if canViewAll {
		scope = "all"
	}
	// Never include message bodies, raw visitor fingerprints, addresses, or
	// provider payloads in the audit record. The record only proves that the
	// already-authorized projection was read and which access scope applied.
	_ = s.auditService.CreateAuditLog(&audit.AuditLog{
		UserID:     agentUserID,
		Action:     "view",
		Resource:   "customer_service_context",
		ResourceID: ticketID,
		Method:     "GET",
		Path:       fmt.Sprintf("/api/admin/customer-service/conversations/%d/context", ticketID),
		Changes:    fmt.Sprintf(`{"scope":%q,"projection":"redacted"}`, scope),
		Status:     "success",
	})
}

func (s *CustomerServiceContextService) ConversationListSummary(t ticket.Ticket) CustomerServiceConversationSummary {
	summary := CustomerServiceConversationSummary{
		Type:          "visitor",
		Identity:      "visitor",
		IdentityLabel: "游客",
		DisplayName:   "匿名客户",
		RegionLabel:   "未知区域",
		RegionStatus:  "unknown",
	}

	if t.CustomerUserID != nil && *t.CustomerUserID > 0 {
		summary.Type = "member"
		summary.Identity = "member"
		summary.IdentityLabel = "会员"
		summary.DisplayName = "客户 " + strconv.FormatUint(uint64(*t.CustomerUserID), 10)
		if s != nil && s.userRepo != nil {
			if account, err := s.userRepo.FindByID(*t.CustomerUserID); err == nil && account != nil {
				if displayName := serviceDisplayName(account); displayName != "" {
					summary.DisplayName = displayName
				}
			}
		}
		summary.MemberTier = s.memberTierSummary(*t.CustomerUserID)
	} else if strings.TrimSpace(t.VisitorSessionHash) != "" {
		summary.DisplayName = "游客 " + hashPreview(t.VisitorSessionHash)
	}

	if label := s.conversationCoarseRegionLabel(t); label != "" {
		summary.RegionLabel = label
		summary.RegionStatus = "captured"
	}

	return summary
}

func (s *CustomerServiceContextService) memberTierSummary(userID uint) *CustomerServiceMemberTier {
	if s == nil || s.loyaltyRepo == nil || userID == 0 {
		return nil
	}

	userLoyalty, err := s.loyaltyRepo.FindUserLoyaltyByUserID(userID)
	if err != nil || userLoyalty == nil {
		return nil
	}

	level := s.memberLevelForLoyalty(userLoyalty)
	if level == nil {
		return nil
	}

	return &CustomerServiceMemberTier{
		ID:              level.ID,
		Name:            strings.TrimSpace(level.Name),
		Icon:            strings.TrimSpace(level.Icon),
		Color:           strings.TrimSpace(level.Color),
		TotalPoints:     userLoyalty.TotalPoints,
		AvailablePoints: userLoyalty.AvailablePoints,
	}
}

func (s *CustomerServiceContextService) memberLevelForLoyalty(userLoyalty *loyalty.UserLoyalty) *loyalty.MemberLevel {
	if s == nil || s.loyaltyRepo == nil || userLoyalty == nil {
		return nil
	}
	if userLoyalty.MemberLevelID > 0 {
		if level, err := s.loyaltyRepo.FindMemberLevelByID(userLoyalty.MemberLevelID); err == nil && level != nil {
			return level
		}
	}
	if level, err := s.loyaltyRepo.FindMemberLevelByPoints(userLoyalty.TotalPoints); err == nil && level != nil {
		return level
	}
	return nil
}

func (s *CustomerServiceContextService) conversationCoarseRegionLabel(t ticket.Ticket) string {
	if s == nil || s.visitorProfileService == nil {
		return ""
	}

	if strings.TrimSpace(t.VisitorSessionHash) != "" {
		if profile, err := s.visitorProfileService.FindByCustomerServiceVisitorHash(t.VisitorSessionHash); err == nil {
			if label := visitorProfileCoarseRegionLabel(profile); label != "" {
				return label
			}
		}
	}

	if t.CustomerUserID != nil && *t.CustomerUserID > 0 {
		if profile, err := s.visitorProfileService.FindByUserID(*t.CustomerUserID); err == nil {
			return visitorProfileCoarseRegionLabel(profile)
		}
	}

	return ""
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
	applyProfileTimezone(context, profile)

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

func (s *CustomerServiceContextService) applyVisitorTimezone(context *CustomerServiceContext, visitorSessionHash string, userID uint) {
	if context == nil || s == nil || s.visitorProfileService == nil {
		return
	}

	if strings.TrimSpace(visitorSessionHash) != "" {
		if profile, err := s.visitorProfileService.FindByCustomerServiceVisitorHash(visitorSessionHash); err == nil {
			if applyProfileTimezone(context, profile) {
				return
			}
		}
	}

	if userID > 0 {
		if profile, err := s.visitorProfileService.FindByUserID(userID); err == nil {
			applyProfileTimezone(context, profile)
		}
	}
}

func applyProfileTimezone(context *CustomerServiceContext, profile *visitor.Profile) bool {
	if context == nil || profile == nil {
		return false
	}

	timezone := strings.TrimSpace(profile.Timezone)
	if timezone == "" {
		return false
	}

	context.Contact.Timezone = timezone
	context.Contact.TimezoneSource = "visitor_profile"
	return true
}

func applyProfileLocationSignals(context *CustomerServiceContext, profileID uint, countryCode, region, city string) {
	parts := make([]string, 0, 3)
	if countryLabel := countryCodeDisplayName(countryCode); countryLabel != "" {
		parts = append(parts, countryLabel)
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

func visitorProfileCoarseRegionLabel(profile *visitor.Profile) string {
	if profile == nil {
		return ""
	}

	country := countryCodeDisplayName(profile.CountryCode)
	region := strings.TrimSpace(profile.Region)
	if mapped := locationAliasDisplayName(region); mapped != "" {
		region = mapped
	}

	if country != "" && region != "" && !sameLocationLabel(country, region) {
		return country + " / " + region
	}
	if country != "" {
		return country
	}
	return region
}

func countryCodeDisplayName(countryCode string) string {
	code := strings.ToUpper(strings.TrimSpace(countryCode))
	if code == "" {
		return ""
	}
	if mapped := locationAliasDisplayName(code); mapped != "" {
		return mapped
	}
	return code
}

func locationAliasDisplayName(value string) string {
	key := strings.ToUpper(strings.TrimSpace(value))
	key = strings.ReplaceAll(key, "_", " ")
	key = strings.Join(strings.Fields(key), " ")
	switch key {
	case "CN", "CHN", "CHINA", "MAINLAND CHINA", "中国", "中国大陆":
		return "中国大陆"
	case "TW", "TWN", "TAIWAN", "TAIWAN, PROVINCE OF CHINA", "中国台湾", "台湾":
		return "中国台湾"
	case "HK", "HKG", "HONG KONG", "中国香港", "香港":
		return "中国香港"
	case "MO", "MAC", "MACAO", "MACAU", "中国澳门", "澳门":
		return "中国澳门"
	case "US", "USA", "UNITED STATES", "UNITED STATES OF AMERICA":
		return "United States"
	case "JP", "JPN", "JAPAN":
		return "Japan"
	case "DE", "DEU", "GERMANY":
		return "Germany"
	case "GB", "GBR", "UK", "UNITED KINGDOM":
		return "United Kingdom"
	case "FR", "FRA", "FRANCE":
		return "France"
	case "CA", "CAN", "CANADA":
		return "Canada"
	case "AU", "AUS", "AUSTRALIA":
		return "Australia"
	default:
		return ""
	}
}

func sameLocationLabel(left, right string) bool {
	normalize := func(value string) string {
		value = strings.ToLower(strings.TrimSpace(value))
		value = strings.ReplaceAll(value, " ", "")
		value = strings.ReplaceAll(value, "-", "")
		value = strings.ReplaceAll(value, "_", "")
		value = strings.ReplaceAll(value, ",", "")
		return value
	}
	return normalize(left) != "" && normalize(left) == normalize(right)
}

func (s *CustomerServiceContextService) customerCartContext(userID uint) CustomerServiceContextCart {
	result := CustomerServiceContextCart{Available: true, Status: "available", Currency: product.DefaultPriceCurrency, Items: []CustomerServiceContextCartItem{}}
	if s == nil || s.cartRepo == nil || userID == 0 {
		result.Available = false
		result.Status = "unavailable"
		result.Reason = "购物车事实源不可用。"
		return result
	}
	cart, err := s.cartRepo.FindByUserID(userID)
	if repository.IsRecordNotFound(err) {
		return result
	}
	if err != nil {
		result.Available = false
		result.Status = "error"
		result.Reason = err.Error()
		return result
	}
	summary, err := s.cartRepo.GetSummary(cart.ID)
	if err != nil {
		result.Available = false
		result.Status = "error"
		result.Reason = err.Error()
		return result
	}
	result.ItemCount = summary.ItemCount
	result.Currency = summary.TotalMoney.Currency().String()
	if total, totalErr := summary.TotalMoney.FormatMajor(); totalErr == nil {
		result.Total = total
	} else {
		result.Available = false
		result.Status = "error"
		result.Reason = totalErr.Error()
	}
	result.Items, err = customerCartItems(summary.Items, s.mediaURLResolver)
	if err != nil {
		result.Available = false
		result.Reason = err.Error()
		result.Items = []CustomerServiceContextCartItem{}
	}
	return result
}

func (s *CustomerServiceContextService) customerCartContextBySessionID(sessionID string) CustomerServiceContextCart {
	result := CustomerServiceContextCart{Available: true, Status: "available", Currency: product.DefaultPriceCurrency, Items: []CustomerServiceContextCartItem{}}
	if s == nil || s.cartRepo == nil || strings.TrimSpace(sessionID) == "" {
		result.Available = false
		result.Status = "unavailable"
		result.Reason = "购物车 session 未绑定。"
		return result
	}
	cart, err := s.cartRepo.FindBySessionID(strings.TrimSpace(sessionID))
	if repository.IsRecordNotFound(err) {
		result.Available = false
		result.Status = "unavailable"
		result.Reason = "visitor profile 已绑定购物车 session，但当前 session 没有购物车记录。"
		return result
	}
	if err != nil {
		result.Available = false
		result.Status = "error"
		result.Reason = err.Error()
		return result
	}
	summary, err := s.cartRepo.GetSummary(cart.ID)
	if err != nil {
		result.Available = false
		result.Status = "error"
		result.Reason = err.Error()
		return result
	}
	result.ItemCount = summary.ItemCount
	result.Currency = summary.TotalMoney.Currency().String()
	if total, totalErr := summary.TotalMoney.FormatMajor(); totalErr == nil {
		result.Total = total
	} else {
		result.Available = false
		result.Reason = totalErr.Error()
	}
	result.Items, err = customerCartItems(summary.Items, s.mediaURLResolver)
	if err != nil {
		result.Available = false
		result.Status = "error"
		result.Reason = err.Error()
		result.Items = []CustomerServiceContextCartItem{}
	}
	return result
}

func (s *CustomerServiceContextService) customerWishlistContext(userID uint) CustomerServiceContextWishlist {
	result := CustomerServiceContextWishlist{Available: true, Status: "available", Items: []CustomerServiceContextWishlistItem{}}
	if s == nil || s.wishlistRepo == nil || userID == 0 {
		result.Available = false
		result.Status = "unavailable"
		result.Reason = "心愿单事实源不可用。"
		return result
	}
	items, err := s.wishlistRepo.ListByUserID(userID)
	if err != nil {
		result.Available = false
		result.Status = "error"
		result.Reason = err.Error()
		return result
	}
	result.Count = len(items)
	result.Items = customerWishlistItems(items, 8, s.mediaURLResolver)
	return result
}

func (s *CustomerServiceContextService) customerOrdersContext(userID uint) CustomerServiceContextOrders {
	result, _, _, _ := s.customerOperationalFactsContext(userID)
	return result
}

func (s *CustomerServiceContextService) customerOperationalFactsContext(userID uint) (
	CustomerServiceContextOrders,
	CustomerServiceContextShippingAddress,
	CustomerServiceContextAfterSales,
	CustomerServiceContextDisputes,
) {
	ordersResult := CustomerServiceContextOrders{Available: false, Status: "unavailable", FulfillmentStatus: "unavailable", Reason: "订单事实源未配置。", Items: []CustomerServiceContextOrderItem{}}
	addressResult := CustomerServiceContextShippingAddress{Available: false, Status: "unavailable", Reason: "订单事实源未配置。"}
	afterSalesResult := CustomerServiceContextAfterSales{Available: false, Status: "unavailable", Reason: "售后事实源未配置。", Items: []CustomerServiceContextAfterSalesItem{}, RefundStatus: "unavailable", WarrantyStatus: "unavailable", Refunds: []CustomerServiceContextRefund{}, Warranty: []CustomerServiceContextWarrantyClaim{}}
	disputesResult := CustomerServiceContextDisputes{Available: false, Status: "unavailable", Reason: "支付争议事实源未配置。", Items: []CustomerServiceContextDisputeItem{}}
	if s == nil || s.orderRepo == nil || userID == 0 {
		return ordersResult, addressResult, afterSalesResult, disputesResult
	}

	orders, total, err := s.orderRepo.FindByUserID(userID, 1, 5)
	if err != nil {
		ordersResult.Reason = err.Error()
		ordersResult.Status = "error"
		addressResult.Status, addressResult.Reason = "error", "订单事实源读取失败。"
		afterSalesResult.Status, afterSalesResult.Reason = "error", "订单事实源读取失败。"
		disputesResult.Status, disputesResult.Reason = "error", "订单事实源读取失败。"
		return ordersResult, addressResult, afterSalesResult, disputesResult
	}
	ordersResult.Available, ordersResult.Status, ordersResult.Total = true, "available", total
	ordersResult.Items, ordersResult.FulfillmentStatus, ordersResult.FulfillmentReason = s.customerOrderItemsWithFacts(orders)
	if ordersResult.FulfillmentStatus == "unavailable" {
		ordersResult.FulfillmentReason = "物流事实源未配置。"
	}
	if len(orders) == 0 {
		ordersResult.Reason = "该账号暂无订单。"
	}
	addressResult = customerShippingAddressContext(orders)

	if s.afterSalesRepo == nil {
		afterSalesResult.Reason = "售后事实源未配置。"
	} else {
		afterSalesResult.Available, afterSalesResult.Status = true, "available"
		if s.paymentRepo == nil {
			afterSalesResult.RefundStatus, afterSalesResult.RefundReason = "unavailable", "退款事实源未配置。"
		} else {
			afterSalesResult.RefundStatus = "available"
		}
		if s.warrantyRepo == nil {
			afterSalesResult.WarrantyStatus, afterSalesResult.WarrantyReason = "unavailable", "保修事实源未配置。"
		} else {
			afterSalesResult.WarrantyStatus = "available"
		}
		for _, orderRecord := range orders {
			cases, caseErr := s.afterSalesRepo.FindByOrderID(orderRecord.ID, "")
			if caseErr != nil {
				afterSalesResult.Available, afterSalesResult.Status = false, "error"
				afterSalesResult.Reason = caseErr.Error()
				break
			}
			for _, record := range cases {
				itemSummary := make([]string, 0, len(record.Items))
				for _, item := range record.Items {
					label := strings.TrimSpace(item.ProductName)
					if label == "" {
						label = fmt.Sprintf("产品 %d", item.ProductID)
					}
					itemSummary = append(itemSummary, fmt.Sprintf("%s x%d", label, item.Quantity))
				}
				currentHandlerID := record.UpdatedBy
				if currentHandlerID == 0 {
					currentHandlerID = record.CreatedBy
				}
				afterSalesResult.Items = append(afterSalesResult.Items, CustomerServiceContextAfterSalesItem{
					ID: record.ID, OrderID: record.OrderID, OrderNumber: orderRecord.OrderNumber,
					Type: record.Type, Status: record.Status, Reason: record.Reason,
					ItemSummary: strings.Join(itemSummary, ", "), CurrentHandlerID: currentHandlerID,
					CurrentHandlerName: s.customerServiceHandlerName(currentHandlerID), Resolution: record.Resolution,
					CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
				})
			}
			if s.paymentRepo != nil {
				refunds, refundErr := s.paymentRepo.FindRefundsByOrderID(orderRecord.ID)
				if refundErr != nil {
					afterSalesResult.RefundStatus, afterSalesResult.RefundReason = "error", refundErr.Error()
				} else {
					for _, refund := range refunds {
						afterSalesResult.Refunds = append(afterSalesResult.Refunds, customerRefundContext(refund, orderRecord.OrderNumber))
					}
				}
			}
		}
		if len(afterSalesResult.Items) == 0 {
			afterSalesResult.Reason = "该账号暂无售后请求。"
		}
	}

	if s.warrantyRepo == nil {
		// Warranty is a separate fact source from generic RMA cases. Keep this
		// explicit so the UI does not turn an unconfigured table into “no claims”.
		if afterSalesResult.Status == "available" {
			afterSalesResult.WarrantyStatus, afterSalesResult.WarrantyReason = "unavailable", "保修事实源未配置；通用售后事实仍可用。"
		}
	} else {
		claims, claimErr := s.warrantyRepo.FindWarrantyClaimsByUserID(userID, 20)
		if claimErr != nil {
			afterSalesResult.WarrantyStatus, afterSalesResult.WarrantyReason = "error", claimErr.Error()
		} else {
			for _, claim := range claims {
				afterSalesResult.Warranty = append(afterSalesResult.Warranty, CustomerServiceContextWarrantyClaim{
					ID: claim.ID, OrderItemID: claim.OrderItemID, OrderNumber: claim.OrderNumber,
					IssueType: claim.IssueType, Status: claim.Status, TirePressure: claim.TirePressure,
					IsTubeless: claim.IsTubeless, Resolution: claim.Resolution, ProcessedBy: claim.ProcessedBy,
					CreatedAt: claim.CreatedAt, UpdatedAt: claim.UpdatedAt,
				})
			}
		}
	}

	if s.paymentRepo == nil {
		disputesResult.Reason = "支付争议事实源未配置。"
	} else {
		disputesResult.Available, disputesResult.Status = true, "available"
		for _, orderRecord := range orders {
			stripeDisputes, stripeErr := s.paymentRepo.ListStripeDisputesByOrderID(orderRecord.ID)
			paypalDisputes, paypalErr := s.paymentRepo.ListPayPalDisputesByOrderID(orderRecord.ID)
			if stripeErr != nil || paypalErr != nil {
				disputesResult.Available, disputesResult.Status = false, "error"
				if stripeErr != nil {
					disputesResult.Reason = stripeErr.Error()
				} else {
					disputesResult.Reason = paypalErr.Error()
				}
				break
			}
			for _, dispute := range stripeDisputes {
				disputeAmount := paymentDisputeDisplayAmount(dispute.AmountMinor, dispute.Currency)
				disputesResult.Items = append(disputesResult.Items, CustomerServiceContextDisputeItem{
					Provider: "stripe", OrderID: orderRecord.ID, OrderNumber: orderRecord.OrderNumber,
					Status: dispute.Status, Reason: dispute.Reason, Amount: disputeAmount, Currency: dispute.Currency,
					EvidenceDueAt: dispute.EvidenceDueAt, EvidenceSubmittedAt: dispute.EvidenceSubmittedAt,
				})
			}
			for _, dispute := range paypalDisputes {
				disputeAmount := paymentDisputeDisplayAmount(dispute.AmountMinor, dispute.Currency)
				disputesResult.Items = append(disputesResult.Items, CustomerServiceContextDisputeItem{
					Provider: "paypal", OrderID: orderRecord.ID, OrderNumber: orderRecord.OrderNumber,
					Status: dispute.Status, Reason: dispute.Reason, Amount: disputeAmount, Currency: dispute.Currency,
					EvidenceSubmittedAt: dispute.EvidenceSubmittedAt,
				})
			}
		}
		if len(disputesResult.Items) == 0 {
			disputesResult.Reason = "该账号暂无支付争议。"
		}
	}

	return ordersResult, addressResult, afterSalesResult, disputesResult
}

func customerShippingAddressContext(orders []order.Order) CustomerServiceContextShippingAddress {
	result := CustomerServiceContextShippingAddress{Available: false, Status: "unavailable", Reason: "暂无可用收货地址。"}
	for _, orderRecord := range orders {
		address := orderRecord.ShippingAddress
		if strings.TrimSpace(address.Address1) == "" && strings.TrimSpace(address.City) == "" && strings.TrimSpace(address.Country) == "" {
			continue
		}
		name := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(address.FirstName), strings.TrimSpace(address.LastName)}, " "))
		line := strings.TrimSpace(address.Address1)
		if strings.TrimSpace(address.Address2) != "" {
			line = strings.TrimSpace(line + " " + strings.TrimSpace(address.Address2))
		}
		result = CustomerServiceContextShippingAddress{
			Available: true, Status: "available", SourceOrderID: orderRecord.ID, SourceOrderNumber: orderRecord.OrderNumber,
			RecipientName: name, AddressLine: line, City: strings.TrimSpace(address.City), State: strings.TrimSpace(address.State),
			PostalCode: strings.TrimSpace(address.PostalCode), Country: strings.TrimSpace(address.Country), PhonePresent: strings.TrimSpace(address.Phone) != "",
		}
		return result
	}
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
	products := make([]product.Product, 0)
	if s != nil && s.productRepo != nil {
		productIDs := make([]uint, 0, len(items))
		seen := make(map[uint]struct{}, len(items))
		for _, item := range items {
			if item.ProductID == 0 {
				continue
			}
			if _, ok := seen[item.ProductID]; ok {
				continue
			}
			seen[item.ProductID] = struct{}{}
			productIDs = append(productIDs, item.ProductID)
		}
		// Product metadata is an optional enrichment. Preserve the browsing
		// facts when a stale/deleted product or a read-side failure is present.
		if loaded, loadErr := s.productRepo.FindProductsByIDsForCustomerContext(productIDs); loadErr == nil {
			products = loaded
		}
	}
	result.Items = customerBrowsingItems(items, products, s.mediaURLResolver)
	return result
}

func customerCartItems(items []product.CartItem, resolvers ...PublicMediaURLResolver) ([]CustomerServiceContextCartItem, error) {
	result := make([]CustomerServiceContextCartItem, 0, len(items))
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("customer cart item %d quantity must be greater than zero", item.ID)
		}
		name := "Unknown product"
		sku := ""
		image := ""
		if item.Product != nil {
			name = item.Product.Name
			sku = item.Product.DisplaySKU()
			image = firstProductImage(item.Product, resolvers...)
		}
		variantName := ""
		if item.Variant != nil {
			if strings.TrimSpace(item.Variant.SKU) != "" {
				sku = item.Variant.SKU
			}
			variantName = strings.TrimSpace(item.Variant.OptionValues)
		}
		unitMoney, err := item.PriceMoney()
		if err != nil {
			return nil, fmt.Errorf("customer cart item %d price: %w", item.ID, err)
		}
		lineMoney, err := unitMoney.MultiplyInt(int64(item.Quantity))
		if err != nil {
			return nil, fmt.Errorf("calculate customer cart item %d total: %w", item.ID, err)
		}
		lineTotal, err := lineMoney.FormatMajor()
		if err != nil {
			return nil, fmt.Errorf("format customer cart item %d total: %w", item.ID, err)
		}
		unitPrice, err := unitMoney.FormatMajor()
		if err != nil {
			return nil, fmt.Errorf("format customer cart item %d price: %w", item.ID, err)
		}
		result = append(result, CustomerServiceContextCartItem{
			ID:                item.ID,
			ProductID:         item.ProductID,
			VariantID:         item.VariantID,
			Name:              name,
			SKU:               sku,
			Image:             image,
			Quantity:          item.Quantity,
			Currency:          strings.TrimSpace(item.Currency),
			Price:             unitPrice,
			LineTotal:         lineTotal,
			InventorySnapshot: "not_captured",
			VariantName:       variantName,
		})
	}
	return result, nil
}

func customerWishlistItems(items []wishlist.Item, limit int, resolvers ...PublicMediaURLResolver) []CustomerServiceContextWishlistItem {
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
			image = firstProductImage(item.Product, resolvers...)
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
		subtotalAmount, discountAmount, shippingFee, taxAmount, totalAmount := "0", "0", "0", "0", "0"
		if value, err := item.SubtotalMoney(); err == nil {
			subtotalAmount, _ = value.FormatMajor()
		}
		if value, err := item.DiscountMoney(); err == nil {
			discountAmount, _ = value.FormatMajor()
		}
		if value, err := item.ShippingFeeMoney(); err == nil {
			shippingFee, _ = value.FormatMajor()
		}
		if value, err := item.TaxMoney(); err == nil {
			taxAmount, _ = value.FormatMajor()
		}
		if value, err := item.TotalMoney(); err == nil {
			totalAmount, _ = value.FormatMajor()
		}
		discountedSubtotal := "0"
		// Customer-service payloads are a display/read model, but derive the
		// discounted subtotal from the immutable minor-unit snapshot so floating
		// point arithmetic never influences an order fact.
		if subtotalMoney, subtotalErr := item.SubtotalMoney(); subtotalErr == nil {
			if discountMoney, discountErr := item.DiscountMoney(); discountErr == nil {
				if netMoney, netErr := subtotalMoney.Subtract(discountMoney); netErr == nil {
					if netMoney.AmountMinor() < 0 {
						netMoney, _ = domainmoney.New(0, subtotalMoney.Currency().String())
					}
					discountedSubtotal, _ = netMoney.FormatMajor()
				}
			}
		}
		result = append(result, CustomerServiceContextOrderItem{
			ID:                       item.ID,
			OrderNumber:              item.OrderNumber,
			Status:                   item.Status,
			PaymentStatus:            item.PaymentStatus,
			ShippingStatus:           item.ShippingStatus,
			Currency:                 strings.TrimSpace(item.Currency),
			SubtotalAmount:           subtotalAmount,
			DiscountAmount:           discountAmount,
			DiscountedSubtotalAmount: discountedSubtotal,
			ShippingFee:              shippingFee,
			TaxAmount:                taxAmount,
			TotalAmount:              totalAmount,
			ItemCount:                len(item.Items),
			Items:                    customerOrderLines(item.Items, nil),
			Shipments:                []CustomerServiceContextShipment{},
			CreatedAt:                item.CreatedAt,
		})
	}
	return result
}

func (s *CustomerServiceContextService) customerOrderItemsWithFacts(items []order.Order) ([]CustomerServiceContextOrderItem, string, string) {
	result := customerOrderItems(items)
	for index := range items {
		result[index].Items = customerOrderLines(items[index].Items, nil)
	}

	if s != nil && s.productRepo != nil {
		productIDs := make([]uint, 0)
		seen := make(map[uint]struct{})
		for _, record := range items {
			for _, line := range record.Items {
				if line.ProductID == 0 {
					continue
				}
				if _, ok := seen[line.ProductID]; ok {
					continue
				}
				seen[line.ProductID] = struct{}{}
				productIDs = append(productIDs, line.ProductID)
			}
		}
		if products, err := s.productRepo.FindProductsByIDsForCustomerContext(productIDs); err == nil {
			images := make(map[uint]string, len(products))
			for index := range products {
				images[products[index].ID] = firstProductImage(&products[index], s.mediaURLResolver)
			}
			for orderIndex := range result {
				for lineIndex := range result[orderIndex].Items {
					result[orderIndex].Items[lineIndex].Thumbnail = images[result[orderIndex].Items[lineIndex].ProductID]
				}
			}
		}
	}

	if s == nil || s.shippingRepo == nil {
		return result, "unavailable", "物流事实源未配置。"
	}
	for index, record := range items {
		shipments, err := s.shippingRepo.FindTrackingShipmentsByOrderID(record.ID)
		if err != nil {
			return result, "error", err.Error()
		}
		result[index].Shipments = customerTrackingShipments(shipments)
	}
	return result, "available", ""
}

func customerOrderLines(items []order.OrderItem, imageByProduct map[uint]string) []CustomerServiceContextOrderLine {
	result := make([]CustomerServiceContextOrderLine, 0, len(items))
	for _, item := range items {
		currency := strings.TrimSpace(item.Currency)
		line := CustomerServiceContextOrderLine{
			ID: item.ID, ProductID: item.ProductID, VariantID: item.VariantID,
			Name: strings.TrimSpace(item.ProductName), SKU: strings.TrimSpace(item.SKU), Quantity: item.Quantity,
			Currency: currency, FulfillmentMode: order.NormalizeFulfillmentMode(item.FulfillmentMode),
			InventorySnapshot: "not_captured", PricingSnapshot: snapshotAvailability(item.PricingSnapshotData),
		}
		if imageByProduct != nil {
			line.Thumbnail = imageByProduct[item.ProductID]
		}
		if money, err := item.PriceMoney(); err == nil {
			line.UnitPrice, _ = money.FormatMajor()
		}
		if money, err := item.SubtotalMoney(); err == nil {
			line.SubtotalAmount, _ = money.FormatMajor()
		}
		if money, err := item.DiscountMoney(); err == nil {
			line.DiscountAmount, _ = money.FormatMajor()
		}
		if money, err := item.TotalMoney(); err == nil {
			line.TotalAmount, _ = money.FormatMajor()
		}
		result = append(result, line)
	}
	return result
}

func customerTrackingShipments(shipments []shippingdomain.TrackingShipment) []CustomerServiceContextShipment {
	result := make([]CustomerServiceContextShipment, 0, len(shipments))
	for _, shipment := range shipments {
		carrier := ""
		if shipment.Carrier != nil {
			carrier = strings.TrimSpace(shipment.Carrier.Name)
		}
		carrierService := ""
		if shipment.CarrierService != nil {
			carrierService = strings.TrimSpace(shipment.CarrierService.ServiceName)
		}
		result = append(result, CustomerServiceContextShipment{
			ID: shipment.ID, Carrier: carrier, CarrierService: carrierService,
			TrackingNumber: strings.TrimSpace(shipment.TrackingNumber), RegistrationStatus: strings.TrimSpace(shipment.RegistrationStatus),
			SyncStatus: strings.TrimSpace(shipment.SyncStatus), LastEventAt: shipment.LastEventAt,
		})
	}
	return result
}

func snapshotAvailability(raw []byte) string {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "{}" || !json.Valid([]byte(value)) {
		return "not_captured"
	}
	return "captured"
}

func maxFloat(value, floor float64) float64 {
	if value < floor {
		return floor
	}
	return value
}

func (s *CustomerServiceContextService) customerServiceHandlerName(userID uint) string {
	if userID == 0 || s == nil || s.userRepo == nil {
		return ""
	}
	userRecord, err := s.userRepo.FindByID(userID)
	if err != nil {
		return ""
	}
	return serviceDisplayName(userRecord)
}

func customerRefundContext(refund payment.Refund, orderNumber string) CustomerServiceContextRefund {
	// Customer-service payloads are a display/read model, but the source of
	// truth is the immutable minor-unit refund snapshot. Never consult the
	// removed major-unit projection, even when a historical row is malformed.
	amount := "0"
	if money, err := refund.AmountMoney(); err == nil {
		amount, _ = money.FormatMajor()
	}
	return CustomerServiceContextRefund{
		ID: refund.ID, OrderID: refund.OrderID, OrderNumber: orderNumber,
		Status: strings.TrimSpace(refund.Status), Reason: strings.TrimSpace(refund.Reason),
		Amount: amount, Currency: strings.TrimSpace(refund.Currency), CreatedAt: refund.CreatedAt, CompletedAt: refund.CompletedAt,
	}
}

func customerBrowsingItems(items []user.BrowsingHistory, products []product.Product, resolvers ...PublicMediaURLResolver) []CustomerServiceContextBrowsingItem {
	result := make([]CustomerServiceContextBrowsingItem, 0, len(items))
	productByID := make(map[uint]*product.Product, len(products))
	for index := range products {
		productByID[products[index].ID] = &products[index]
	}
	for _, item := range items {
		browsing := CustomerServiceContextBrowsingItem{
			ProductID:    item.ProductID,
			ViewCount:    item.ViewCount,
			LastViewedAt: item.LastViewedAt,
		}
		if viewedProduct := productByID[item.ProductID]; viewedProduct != nil {
			browsing.Name = strings.TrimSpace(viewedProduct.Name)
			browsing.SKU = strings.TrimSpace(viewedProduct.DisplaySKU())
			browsing.Thumbnail = firstProductImage(viewedProduct, resolvers...)
			browsing.Currency = strings.TrimSpace(viewedProduct.DisplayPriceCurrency())
			if variant := viewedProduct.StartingPriceVariant(); variant != nil {
				if money, err := variant.PriceMoney(); err == nil {
					browsing.Price, _ = money.FormatMajor()
				}
			} else if money, err := viewedProduct.PriceMoney(); err == nil {
				browsing.Price, _ = money.FormatMajor()
			}
		}
		result = append(result, browsing)
	}
	return result
}

func firstProductImage(item *product.Product, resolvers ...PublicMediaURLResolver) string {
	if item == nil {
		return ""
	}
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible && strings.TrimSpace(media.URL) != "" {
			var resolver PublicMediaURLResolver
			if len(resolvers) > 0 {
				resolver = resolvers[0]
			}
			return canonicalPublicMediaURL(resolver, media.URL)
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
