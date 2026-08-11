package service

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/product"
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
	loyaltyRepo           *repository.LoyaltyRepository
	visitorProfileService *VisitorProfileService
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

type CustomerServiceRegionAnalyticsInput struct {
	Date                  string
	TimezoneOffsetMinutes int
	AgentUserID           uint
	CanViewAll            bool
}

type CustomerServiceRegionAnalytics struct {
	Date                  string                               `json:"date"`
	TimezoneOffsetMinutes int                                  `json:"timezone_offset_minutes"`
	WindowStart           time.Time                            `json:"window_start"`
	WindowEnd             time.Time                            `json:"window_end"`
	TotalConversations    int                                  `json:"total_conversations"`
	KnownRegionCount      int                                  `json:"known_region_count"`
	UnknownRegionCount    int                                  `json:"unknown_region_count"`
	Regions               []CustomerServiceRegionAnalyticsItem `json:"regions"`
}

type CustomerServiceRegionAnalyticsItem struct {
	RegionLabel  string  `json:"region_label"`
	RegionStatus string  `json:"region_status"`
	Count        int     `json:"count"`
	MemberCount  int     `json:"member_count"`
	VisitorCount int     `json:"visitor_count"`
	Percent      float64 `json:"percent"`
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
			MemberTier:  s.memberTierSummary(customerUserID),
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

func (s *CustomerServiceContextService) RegionAnalyticsForAgent(input CustomerServiceRegionAnalyticsInput) (*CustomerServiceRegionAnalytics, error) {
	start, end, date := customerServiceAnalyticsWindow(input.Date, input.TimezoneOffsetMinutes)
	result := &CustomerServiceRegionAnalytics{
		Date:                  date,
		TimezoneOffsetMinutes: input.TimezoneOffsetMinutes,
		WindowStart:           start,
		WindowEnd:             end,
		Regions:               []CustomerServiceRegionAnalyticsItem{},
	}
	if s == nil || s.ticketService == nil {
		return result, nil
	}

	conversations, err := s.ticketService.ListCustomerServiceConversationsInWindowForAgent(
		start,
		end,
		input.AgentUserID,
		input.CanViewAll,
	)
	if err != nil {
		return nil, err
	}

	regionMap := map[string]*CustomerServiceRegionAnalyticsItem{}
	for _, conversation := range conversations {
		summary := s.ConversationListSummary(conversation)
		label := strings.TrimSpace(summary.RegionLabel)
		status := strings.TrimSpace(summary.RegionStatus)
		if label == "" {
			label = "未知区域"
		}
		if status == "" {
			status = "unknown"
		}

		item := regionMap[label]
		if item == nil {
			item = &CustomerServiceRegionAnalyticsItem{
				RegionLabel:  label,
				RegionStatus: status,
			}
			regionMap[label] = item
		}
		item.Count++
		if summary.Identity == "member" || summary.Type == "member" {
			item.MemberCount++
		} else {
			item.VisitorCount++
		}
	}

	result.TotalConversations = len(conversations)
	for _, item := range regionMap {
		if item.RegionStatus == "captured" {
			result.KnownRegionCount += item.Count
		} else {
			result.UnknownRegionCount += item.Count
		}
		if result.TotalConversations > 0 {
			item.Percent = math.Round((float64(item.Count)/float64(result.TotalConversations))*1000) / 10
		}
		result.Regions = append(result.Regions, *item)
	}

	sort.SliceStable(result.Regions, func(i, j int) bool {
		if result.Regions[i].Count == result.Regions[j].Count {
			return result.Regions[i].RegionLabel < result.Regions[j].RegionLabel
		}
		return result.Regions[i].Count > result.Regions[j].Count
	})

	return result, nil
}

func customerServiceAnalyticsWindow(date string, timezoneOffsetMinutes int) (time.Time, time.Time, string) {
	location := time.FixedZone("admin-local", timezoneOffsetMinutes*60)
	date = strings.TrimSpace(date)
	if date == "" {
		date = time.Now().In(location).Format("2006-01-02")
	}

	startLocal, err := time.ParseInLocation("2006-01-02", date, location)
	if err != nil {
		startLocal = time.Now().In(location)
		startLocal = time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day(), 0, 0, 0, 0, location)
		date = startLocal.Format("2006-01-02")
	}

	endLocal := startLocal.AddDate(0, 0, 1)
	return startLocal.UTC(), endLocal.UTC(), date
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
