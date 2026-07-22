package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/tracking"
	"tanzanite/internal/repository"
	"time"
)

type ShippingService struct {
	shippingRepo *repository.ShippingRepository
	productRepo  *repository.ProductRepository
	trackingRun  TrackingPollingRunState
	webhookRun   TrackingWebhookRunState
	trackingMu   sync.RWMutex
}

type ShippingCalculationInput struct {
	TemplateID uint
	Weight     float64
	Quantity   int
	Amount     float64
	Country    string
}

type ShippingQuoteItemInput struct {
	ProductID     uint    `json:"product_id"`
	VariantID     *uint   `json:"variant_id,omitempty"`
	ProductTypeID *uint   `json:"product_type_id,omitempty"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	WeightGrams   int     `json:"weight_grams"`
}

type ShippingQuoteInput struct {
	Country  string                   `json:"country"`
	Amount   float64                  `json:"amount"`
	Currency string                   `json:"currency,omitempty"`
	Items    []ShippingQuoteItemInput `json:"items"`
}

type ShippingQuoteItem struct {
	ProductID            uint    `json:"product_id"`
	VariantID            *uint   `json:"variant_id,omitempty"`
	ProductTypeID        *uint   `json:"product_type_id,omitempty"`
	TemplateID           uint    `json:"template_id"`
	TemplateName         string  `json:"template_name"`
	PackagingRuleID      *uint   `json:"packaging_rule_id,omitempty"`
	PackagingRuleName    string  `json:"packaging_rule_name,omitempty"`
	Quantity             int     `json:"quantity"`
	UnitPrice            float64 `json:"unit_price"`
	Amount               float64 `json:"amount"`
	WeightGrams          int     `json:"weight_grams"`
	PackagingWeightGrams int     `json:"packaging_weight_grams"`
	ChargeWeightGrams    int     `json:"charge_weight_grams"`
	ShippingFee          float64 `json:"shipping_fee"`
	FreeShipping         bool    `json:"free_shipping"`
}

type ShippingQuote struct {
	ShippingFee    float64               `json:"shipping_fee"`
	FreeShipping   bool                  `json:"free_shipping"`
	Currency       string                `json:"currency,omitempty"`
	Source         string                `json:"source,omitempty"`
	Items          []ShippingQuoteItem   `json:"items,omitempty"`
	Options        []ShippingQuoteOption `json:"options,omitempty"`
	SelectedOption *ShippingQuoteOption  `json:"selected_option,omitempty"`
}

type ShippingQuoteOption struct {
	CarrierID             uint    `json:"carrier_id"`
	CarrierName           string  `json:"carrier_name"`
	CarrierCode           string  `json:"carrier_code"`
	CarrierServiceID      uint    `json:"carrier_service_id"`
	ServiceCode           string  `json:"service_code"`
	ServiceName           string  `json:"service_name"`
	RouteName             string  `json:"route_name,omitempty"`
	TemplateID            uint    `json:"template_id"`
	TemplateName          string  `json:"template_name"`
	Currency              string  `json:"currency,omitempty"`
	BillingMode           string  `json:"billing_mode"`
	ActualWeightGrams     int     `json:"actual_weight_grams"`
	VolumetricWeightGrams int     `json:"volumetric_weight_grams"`
	ChargeWeightGrams     int     `json:"charge_weight_grams"`
	BillableWeightGrams   int     `json:"billable_weight_grams"`
	BaseFee               float64 `json:"base_fee"`
	FuelSurcharge         float64 `json:"fuel_surcharge"`
	RemoteSurcharge       float64 `json:"remote_surcharge"`
	ShippingFee           float64 `json:"shipping_fee"`
	FreeShipping          bool    `json:"free_shipping"`
	EtaMinDays            int     `json:"eta_min_days"`
	EtaMaxDays            int     `json:"eta_max_days"`
	SortOrder             int     `json:"sort_order"`
}

type TrackingCarrierResolutionInput struct {
	ProviderID       uint
	CarrierID        *uint
	CarrierServiceID *uint
}

type TrackingCarrierResolution struct {
	Provider            *shipping.TrackingProviderConfig
	Carrier             *shipping.Carrier
	CarrierService      *shipping.CarrierService
	Mapping             *shipping.TrackingCarrierMapping
	ProviderCarrierCode string
	ProviderCarrierName string
}

type TrackingSyncInput struct {
	OrderID                  uint
	ProviderID               uint
	TrackingNumber           string
	ProviderCarrierCode      string
	CarrierID                *uint
	CarrierServiceID         *uint
	TrackingCarrierMappingID *uint
}

type TrackingShipmentInput struct {
	OrderID                  uint
	TrackingProviderID       uint
	TrackingNumber           string
	ProviderCarrierCode      string
	CarrierID                *uint
	CarrierServiceID         *uint
	TrackingCarrierMappingID *uint
}

type TrackingShipmentListFilter struct {
	SyncStatus          string
	RegistrationStatus  string
	TrackingNumber      string
	ProviderCarrierCode string
	Keyword             string
	OrderID             uint
	ProviderID          uint
	CarrierID           uint
	CarrierServiceID    uint
	Enabled             *bool
	DueOnly             bool
	Limit               int
}

type TrackingSyncResult struct {
	TrackingNumber string                     `json:"tracking_number"`
	Carrier        string                     `json:"carrier"`
	Status         string                     `json:"status"`
	StatusCode     int                        `json:"status_code"`
	Events         []shipping.TrackingEvent   `json:"events"`
	Shipment       *shipping.TrackingShipment `json:"shipment,omitempty"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

type TrackingShipmentSyncFailure struct {
	OrderID        uint   `json:"order_id"`
	TrackingNumber string `json:"tracking_number"`
	Error          string `json:"error"`
}

type TrackingShipmentSyncBatchResult struct {
	Matched int                           `json:"matched"`
	Synced  int                           `json:"synced"`
	Failed  int                           `json:"failed"`
	Results []TrackingSyncResult          `json:"results"`
	Errors  []TrackingShipmentSyncFailure `json:"errors"`
}

type TrackingPollingRunState struct {
	Enabled         bool                          `json:"enabled"`
	Running         bool                          `json:"running"`
	Interval        string                        `json:"interval"`
	IntervalSeconds int                           `json:"interval_seconds"`
	BatchLimit      int                           `json:"batch_limit"`
	LastStartedAt   *time.Time                    `json:"last_started_at,omitempty"`
	LastFinishedAt  *time.Time                    `json:"last_finished_at,omitempty"`
	LastDurationMs  int64                         `json:"last_duration_ms"`
	LastMatched     int                           `json:"last_matched"`
	LastSynced      int                           `json:"last_synced"`
	LastFailed      int                           `json:"last_failed"`
	LastError       string                        `json:"last_error"`
	LastErrors      []TrackingShipmentSyncFailure `json:"last_errors,omitempty"`
}

type TrackingWebhookEventInput struct {
	Status      string    `json:"status"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	EventTime   time.Time `json:"event_time"`
}

type TrackingWebhookInput struct {
	ProviderID          uint                        `json:"provider_id"`
	TrackingNumber      string                      `json:"tracking_number"`
	ProviderCarrierCode string                      `json:"provider_carrier_code"`
	Status              string                      `json:"status"`
	StatusCode          int                         `json:"status_code"`
	Events              []TrackingWebhookEventInput `json:"events"`
}

type TrackingWebhookResult struct {
	Shipment *shipping.TrackingShipment `json:"shipment"`
	Events   []shipping.TrackingEvent   `json:"events"`
}

type TrackingWebhookRunState struct {
	LastReceivedAt       *time.Time `json:"last_received_at,omitempty"`
	LastFinishedAt       *time.Time `json:"last_finished_at,omitempty"`
	LastDurationMs       int64      `json:"last_duration_ms"`
	LastProviderCode     string     `json:"last_provider_code"`
	LastProviderID       uint       `json:"last_provider_id"`
	LastTrackingNumber   string     `json:"last_tracking_number"`
	LastCarrierCode      string     `json:"last_carrier_code"`
	LastOrderID          uint       `json:"last_order_id"`
	LastEventCount       int        `json:"last_event_count"`
	LastHTTPStatus       int        `json:"last_http_status"`
	LastAccepted         bool       `json:"last_accepted"`
	LastSignatureChecked bool       `json:"last_signature_checked"`
	LastSignatureValid   bool       `json:"last_signature_valid"`
	LastError            string     `json:"last_error"`
}

type resolvedShippingItem struct {
	ShippingQuoteItemInput
	Amount               float64
	Template             *shipping.ShippingTemplate
	PackagingRule        *shipping.PackagingRule
	PackagingWeightGrams int
	ChargeWeightGrams    int
}

type shippingQuoteGroup struct {
	Template         *shipping.ShippingTemplate
	ItemIndexes      []int
	Amount           float64
	Quantity         int
	TotalWeightGrams int
}

var (
	ErrTrackingProviderRequired       = errors.New("tracking provider is required")
	ErrTrackingLocalTargetRequired    = errors.New("carrier or carrier service is required")
	ErrTrackingProviderDisabled       = errors.New("tracking provider is disabled")
	ErrTrackingCarrierDisabled        = errors.New("carrier is disabled")
	ErrTrackingCarrierServiceDisabled = errors.New("carrier service is disabled")
	ErrTrackingCarrierMappingMissing  = errors.New("tracking carrier mapping is not configured")
	ErrTrackingOrderRequired          = errors.New("tracking order id is required")
	ErrTrackingCarrierCodeRequired    = errors.New("tracking provider carrier code is required")
	ErrTrackingProviderAPIKeyMissing  = errors.New("tracking provider api key is required")
	ErrTrackingProviderBaseURLMissing = errors.New("tracking provider base url is required")
	ErrTrackingProviderUnsupported    = errors.New("tracking provider is not supported")
)

const (
	trackingRegistrationPending = "pending"
	trackingRegistrationFailed  = "failed"
	trackingRegistrationSynced  = "registered"

	trackingSyncPending = "pending"

	defaultTrackingPollingIntervalMinutes = 60
)

func NewShippingService(shippingRepo *repository.ShippingRepository, productRepo ...*repository.ProductRepository) *ShippingService {
	service := &ShippingService{shippingRepo: shippingRepo}
	if len(productRepo) > 0 {
		service.productRepo = productRepo[0]
	}
	return service
}

func (s *ShippingService) ConfigureTrackingPolling(enabled bool, interval time.Duration, batchLimit int) {
	if s == nil {
		return
	}

	s.trackingMu.Lock()
	defer s.trackingMu.Unlock()

	s.trackingRun.Enabled = enabled
	s.trackingRun.Interval = interval.String()
	s.trackingRun.IntervalSeconds = int(interval.Seconds())
	s.trackingRun.BatchLimit = batchLimit
	if !enabled {
		s.trackingRun.Running = false
	}
}

func (s *ShippingService) MarkTrackingPollingStarted(startedAt time.Time) {
	if s == nil {
		return
	}

	s.trackingMu.Lock()
	defer s.trackingMu.Unlock()

	s.trackingRun.Running = true
	s.trackingRun.LastStartedAt = &startedAt
	s.trackingRun.LastFinishedAt = nil
	s.trackingRun.LastDurationMs = 0
	s.trackingRun.LastError = ""
}

func (s *ShippingService) MarkTrackingPollingFinished(startedAt time.Time, result *TrackingShipmentSyncBatchResult, err error) {
	if s == nil {
		return
	}

	finishedAt := time.Now()

	s.trackingMu.Lock()
	defer s.trackingMu.Unlock()

	s.trackingRun.Running = false
	s.trackingRun.LastStartedAt = &startedAt
	s.trackingRun.LastFinishedAt = &finishedAt
	s.trackingRun.LastDurationMs = finishedAt.Sub(startedAt).Milliseconds()
	if result != nil {
		s.trackingRun.LastMatched = result.Matched
		s.trackingRun.LastSynced = result.Synced
		s.trackingRun.LastFailed = result.Failed
		s.trackingRun.LastErrors = append([]TrackingShipmentSyncFailure(nil), result.Errors...)
	} else {
		s.trackingRun.LastMatched = 0
		s.trackingRun.LastSynced = 0
		s.trackingRun.LastFailed = 0
		s.trackingRun.LastErrors = nil
	}
	if err != nil {
		s.trackingRun.LastError = err.Error()
	} else {
		s.trackingRun.LastError = ""
	}
}

func (s *ShippingService) TrackingPollingState() TrackingPollingRunState {
	if s == nil {
		return TrackingPollingRunState{}
	}

	s.trackingMu.RLock()
	defer s.trackingMu.RUnlock()

	state := s.trackingRun
	state.LastErrors = append([]TrackingShipmentSyncFailure(nil), s.trackingRun.LastErrors...)
	return state
}

func (s *ShippingService) RecordTrackingWebhookRun(state TrackingWebhookRunState) {
	if s == nil {
		return
	}

	s.trackingMu.Lock()
	defer s.trackingMu.Unlock()

	s.webhookRun = state
}

func (s *ShippingService) TrackingWebhookState() TrackingWebhookRunState {
	if s == nil {
		return TrackingWebhookRunState{}
	}

	s.trackingMu.RLock()
	defer s.trackingMu.RUnlock()

	return s.webhookRun
}

func (s *ShippingService) ListTemplates() ([]shipping.ShippingTemplate, error) {
	return s.shippingRepo.FindAllTemplates()
}

func (s *ShippingService) GetTemplate(id uint) (*shipping.ShippingTemplate, error) {
	return s.shippingRepo.FindTemplateByID(id)
}

func (s *ShippingService) CreateTemplate(template *shipping.ShippingTemplate) error {
	return s.shippingRepo.CreateTemplateWithRules(template, template.Rules)
}

func (s *ShippingService) UpdateTemplate(template *shipping.ShippingTemplate) error {
	return s.shippingRepo.UpdateTemplateWithRules(template, template.Rules)
}

func (s *ShippingService) DeleteTemplate(id uint) error {
	return s.shippingRepo.DeleteTemplate(id)
}

func (s *ShippingService) ListTemplateBindings() ([]shipping.ShippingTemplateBinding, error) {
	return s.shippingRepo.FindAllTemplateBindings()
}

func (s *ShippingService) GetTemplateBinding(id uint) (*shipping.ShippingTemplateBinding, error) {
	return s.shippingRepo.FindTemplateBindingByID(id)
}

func (s *ShippingService) CreateTemplateBinding(binding *shipping.ShippingTemplateBinding) error {
	return s.shippingRepo.CreateTemplateBinding(binding)
}

func (s *ShippingService) UpdateTemplateBinding(binding *shipping.ShippingTemplateBinding) error {
	return s.shippingRepo.UpdateTemplateBinding(binding)
}

func (s *ShippingService) DeleteTemplateBinding(id uint) error {
	return s.shippingRepo.DeleteTemplateBinding(id)
}

func (s *ShippingService) CreateTemplateRule(templateID uint, rule *shipping.ShippingRule) error {
	rule.TemplateID = templateID
	return s.shippingRepo.CreateRule(rule)
}

func (s *ShippingService) UpdateTemplateRule(templateID uint, rule *shipping.ShippingRule) error {
	rule.TemplateID = templateID
	return s.shippingRepo.UpdateRuleForTemplate(rule)
}

func (s *ShippingService) DeleteTemplateRule(templateID uint, ruleID uint) error {
	return s.shippingRepo.DeleteRuleForTemplate(templateID, ruleID)
}

func (s *ShippingService) CalculateShipping(input ShippingCalculationInput) (*ShippingQuote, error) {
	template, err := s.GetTemplate(input.TemplateID)
	if err != nil {
		return nil, err
	}

	if template.FreeShipping && input.Amount >= template.FreeThreshold {
		return &ShippingQuote{ShippingFee: 0, FreeShipping: true}, nil
	}

	value := input.Weight
	switch template.Type {
	case "quantity":
		value = float64(input.Quantity)
	case "price", "amount":
		value = input.Amount
	}

	shippingFee := template.DefaultFee
	for _, rule := range template.Rules {
		if shippingRuleMatchesCountry(rule.Region, input.Country) && value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue) {
			shippingFee = calculateRuleFee(rule, value)
			break
		}
	}

	return &ShippingQuote{ShippingFee: roundMoney(shippingFee), FreeShipping: false}, nil
}

func (s *ShippingService) QuoteCart(input ShippingQuoteInput) (*ShippingQuote, error) {
	if s.productRepo == nil {
		return nil, errors.New("shipping quote product repository is not configured")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("shipping quote requires at least one item")
	}

	items := make([]ShippingQuoteItemInput, 0, len(input.Items))
	var amount float64
	for _, item := range input.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product ID %d", item.ProductID)
		}

		product, variant, err := s.productRepo.FindPurchasableVariant(item.ProductID, item.VariantID)
		if err != nil {
			return nil, fmt.Errorf("product ID %d is not available for shipping quote: %w", item.ProductID, err)
		}
		if variant == nil {
			return nil, fmt.Errorf("product ID %d has no purchasable SKU", item.ProductID)
		}
		if variant.Weight <= 0 {
			return nil, fmt.Errorf("shipping weight is missing for SKU %s", variant.SKU)
		}

		resolvedVariantID := variant.ID
		unitPrice := variant.EffectivePrice()
		amount += unitPrice * float64(item.Quantity)
		items = append(items, ShippingQuoteItemInput{
			ProductID:     product.ID,
			VariantID:     &resolvedVariantID,
			ProductTypeID: product.ProductTypeID,
			Quantity:      item.Quantity,
			UnitPrice:     unitPrice,
			WeightGrams:   variant.Weight,
		})
	}

	input.Items = items
	input.Amount = amount
	return s.QuoteResolvedItems(input)
}

func (s *ShippingService) QuoteResolvedItems(input ShippingQuoteInput) (*ShippingQuote, error) {
	country := strings.ToUpper(strings.TrimSpace(input.Country))
	if country == "" {
		return nil, errors.New("shipping country is required")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("shipping quote requires at least one item")
	}

	bindings, err := s.shippingRepo.FindEnabledTemplateBindingsWithTemplates()
	if err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		return nil, errors.New("no enabled shipping template binding is configured")
	}

	productIDs := uniqueShippingQuoteProductIDs(input.Items)
	packagingRulesByProduct, err := s.shippingRepo.FindActivePackagingRulesByProductIDs(productIDs)
	if err != nil {
		return nil, err
	}

	resolvedItems := make([]resolvedShippingItem, 0, len(input.Items))
	var cartAmount float64
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return nil, errors.New("shipping quote item product_id is required")
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product ID %d", item.ProductID)
		}
		if item.WeightGrams <= 0 {
			if item.VariantID != nil {
				return nil, fmt.Errorf("shipping weight is missing for variant ID %d", *item.VariantID)
			}
			return nil, fmt.Errorf("shipping weight is missing for product ID %d", item.ProductID)
		}

		template := selectShippingTemplateForItem(bindings, item)
		if template == nil {
			return nil, fmt.Errorf("no enabled shipping template binding matches product ID %d", item.ProductID)
		}

		amount := item.UnitPrice * float64(item.Quantity)
		packagingRule := packagingRulesByProduct[item.ProductID]
		packagingWeightGrams := packagingRuleWeightGrams(packagingRule)
		chargeWeightGrams := item.WeightGrams + packagingWeightGrams
		cartAmount += amount
		resolvedItems = append(resolvedItems, resolvedShippingItem{
			ShippingQuoteItemInput: item,
			Amount:                 amount,
			Template:               template,
			PackagingRule:          packagingRule,
			PackagingWeightGrams:   packagingWeightGrams,
			ChargeWeightGrams:      chargeWeightGrams,
		})
	}

	if input.Amount > 0 {
		cartAmount = input.Amount
	}

	groups := make(map[uint]*shippingQuoteGroup)
	quoteItems := make([]ShippingQuoteItem, len(resolvedItems))
	for index, item := range resolvedItems {
		templateID := item.Template.ID
		group := groups[templateID]
		if group == nil {
			group = &shippingQuoteGroup{Template: item.Template}
			groups[templateID] = group
		}

		group.ItemIndexes = append(group.ItemIndexes, index)
		group.Amount += item.Amount
		group.Quantity += item.Quantity
		group.TotalWeightGrams += item.ChargeWeightGrams * item.Quantity

		var packagingRuleID *uint
		var packagingRuleName string
		if item.PackagingRule != nil {
			packagingRuleID = uintPtr(item.PackagingRule.ID)
			packagingRuleName = item.PackagingRule.RuleName
		}

		quoteItems[index] = ShippingQuoteItem{
			ProductID:            item.ProductID,
			VariantID:            item.VariantID,
			ProductTypeID:        item.ProductTypeID,
			TemplateID:           item.Template.ID,
			TemplateName:         item.Template.Name,
			PackagingRuleID:      packagingRuleID,
			PackagingRuleName:    packagingRuleName,
			Quantity:             item.Quantity,
			UnitPrice:            item.UnitPrice,
			Amount:               roundMoney(item.Amount),
			WeightGrams:          item.WeightGrams,
			PackagingWeightGrams: item.PackagingWeightGrams,
			ChargeWeightGrams:    item.ChargeWeightGrams,
		}
	}

	var shippingFee float64
	freeShipping := false
	for _, group := range groups {
		groupFee, groupFree := calculateTemplateShippingFee(
			group.Template,
			country,
			group.TotalWeightGrams,
			group.Quantity,
			group.Amount,
			cartAmount,
		)
		if groupFree {
			freeShipping = true
		}
		shippingFee += groupFee
		distributeGroupFee(group, resolvedItems, quoteItems, groupFee, groupFree)
	}

	shippingFee = roundMoney(shippingFee)
	if shippingFee > 0 {
		freeShipping = false
	}

	currency := strings.TrimSpace(input.Currency)
	if currency == "" {
		currency = "USD"
	}
	currency = strings.ToUpper(currency)

	options, err := s.quoteCarrierServiceOptions(country, currency, resolvedItems, groups, cartAmount)
	if err != nil {
		return nil, err
	}

	source := "template"
	var selectedOption *ShippingQuoteOption
	if len(options) > 0 {
		source = "carrier_service"
		selectedOption = &options[0]
		shippingFee = selectedOption.ShippingFee
		freeShipping = selectedOption.FreeShipping
		if group := singleShippingQuoteGroup(groups); group != nil {
			distributeGroupFee(group, resolvedItems, quoteItems, selectedOption.ShippingFee, selectedOption.FreeShipping)
		}
	}

	return &ShippingQuote{
		ShippingFee:    shippingFee,
		FreeShipping:   freeShipping,
		Currency:       currency,
		Source:         source,
		Items:          quoteItems,
		Options:        options,
		SelectedOption: selectedOption,
	}, nil
}

func (s *ShippingService) ListCarriers(enabledOnly bool) ([]shipping.Carrier, error) {
	return s.shippingRepo.FindAllCarriers(enabledOnly)
}

func (s *ShippingService) GetCarrier(id uint) (*shipping.Carrier, error) {
	return s.shippingRepo.FindCarrierByID(id)
}

func (s *ShippingService) CreateCarrier(carrier *shipping.Carrier) error {
	return s.shippingRepo.CreateCarrier(carrier)
}

func (s *ShippingService) UpdateCarrier(carrier *shipping.Carrier) error {
	return s.shippingRepo.UpdateCarrier(carrier)
}

func (s *ShippingService) DeleteCarrier(id uint) error {
	return s.shippingRepo.DeleteCarrier(id)
}

func (s *ShippingService) ListTrackingProviderConfigs(enabledOnly bool) ([]shipping.TrackingProviderConfig, error) {
	return s.shippingRepo.FindAllTrackingProviderConfigs(enabledOnly)
}

func (s *ShippingService) GetTrackingProviderConfig(id uint) (*shipping.TrackingProviderConfig, error) {
	return s.shippingRepo.FindTrackingProviderConfigByID(id)
}

func (s *ShippingService) GetTrackingProviderConfigByCode(providerCode string) (*shipping.TrackingProviderConfig, error) {
	return s.shippingRepo.FindTrackingProviderConfigByCode(providerCode)
}

func (s *ShippingService) CreateTrackingProviderConfig(provider *shipping.TrackingProviderConfig) error {
	return s.shippingRepo.CreateTrackingProviderConfig(provider)
}

func (s *ShippingService) UpdateTrackingProviderConfig(provider *shipping.TrackingProviderConfig) error {
	return s.shippingRepo.UpdateTrackingProviderConfig(provider)
}

func (s *ShippingService) DeleteTrackingProviderConfig(id uint) error {
	return s.shippingRepo.DeleteTrackingProviderConfig(id)
}

func (s *ShippingService) NewTrackingClientForProvider(providerID uint) (tracking.TrackingService, error) {
	if providerID == 0 {
		return nil, ErrTrackingProviderRequired
	}

	provider, err := s.GetTrackingProviderConfig(providerID)
	if err != nil {
		return nil, err
	}
	return newTrackingClientFromProvider(provider)
}

func (s *ShippingService) GetTrackingShipmentByOrderID(orderID uint) (*shipping.TrackingShipment, error) {
	if orderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	return s.shippingRepo.FindTrackingShipmentByOrderID(orderID)
}

func (s *ShippingService) ListTrackingShipments(filter TrackingShipmentListFilter) ([]shipping.TrackingShipment, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	if filter.Limit > 500 {
		filter.Limit = 500
	}

	return s.shippingRepo.FindAllTrackingShipments(repository.TrackingShipmentFilter{
		SyncStatus:          strings.ToLower(strings.TrimSpace(filter.SyncStatus)),
		RegistrationStatus:  strings.ToLower(strings.TrimSpace(filter.RegistrationStatus)),
		TrackingNumber:      strings.TrimSpace(filter.TrackingNumber),
		ProviderCarrierCode: strings.TrimSpace(filter.ProviderCarrierCode),
		Keyword:             strings.TrimSpace(filter.Keyword),
		OrderID:             filter.OrderID,
		ProviderID:          filter.ProviderID,
		CarrierID:           filter.CarrierID,
		CarrierServiceID:    filter.CarrierServiceID,
		Enabled:             filter.Enabled,
		DueOnly:             filter.DueOnly,
		Limit:               filter.Limit,
	})
}

func (s *ShippingService) SyncDueTrackingShipments(ctx context.Context, limit int) (*TrackingShipmentSyncBatchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	shipments, err := s.shippingRepo.FindDueTrackingShipments(limit, time.Now())
	if err != nil {
		return nil, err
	}

	batch := &TrackingShipmentSyncBatchResult{
		Matched: len(shipments),
		Results: make([]TrackingSyncResult, 0, len(shipments)),
		Errors:  make([]TrackingShipmentSyncFailure, 0),
	}
	for _, shipment := range shipments {
		result, err := s.SyncTracking(ctx, TrackingSyncInput{
			OrderID:                  shipment.OrderID,
			ProviderID:               shipment.TrackingProviderID,
			TrackingNumber:           shipment.TrackingNumber,
			ProviderCarrierCode:      shipment.ProviderCarrierCode,
			CarrierID:                shipment.CarrierID,
			CarrierServiceID:         shipment.CarrierServiceID,
			TrackingCarrierMappingID: shipment.TrackingCarrierMappingID,
		})
		if err != nil {
			batch.Failed++
			batch.Errors = append(batch.Errors, TrackingShipmentSyncFailure{
				OrderID:        shipment.OrderID,
				TrackingNumber: shipment.TrackingNumber,
				Error:          err.Error(),
			})
			continue
		}

		batch.Synced++
		batch.Results = append(batch.Results, *result)
	}

	return batch, nil
}

func (s *ShippingService) ApplyTrackingWebhook(input TrackingWebhookInput) (*TrackingWebhookResult, error) {
	if input.ProviderID == 0 {
		return nil, ErrTrackingProviderRequired
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if trackingNumber == "" {
		return nil, ErrTrackingNumberRequired
	}

	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	shipment, err := s.shippingRepo.FindTrackingShipmentByProviderTrackingNumber(input.ProviderID, trackingNumber, providerCarrierCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(shipment.ProviderCarrierCode) != "" {
		providerCarrierCode = shipment.ProviderCarrierCode
	}

	events := trackingWebhookEventsToDomainEvents(shipment.OrderID, trackingNumber, providerCarrierCode, input)
	if err := s.shippingRepo.ReplaceTrackingEvents(shipment.OrderID, trackingNumber, events); err != nil {
		return nil, err
	}
	if err := s.shippingRepo.UpdateTrackingShipmentSyncSuccess(shipment.OrderID, len(events), latestTrackingEventTime(events), nil); err != nil {
		return nil, err
	}

	updatedShipment, err := s.GetTrackingShipmentByOrderID(shipment.OrderID)
	if err != nil {
		return nil, err
	}

	return &TrackingWebhookResult{
		Shipment: updatedShipment,
		Events:   events,
	}, nil
}

func (s *ShippingService) UpsertTrackingShipment(input TrackingShipmentInput) (*shipping.TrackingShipment, error) {
	if input.OrderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	if input.TrackingProviderID == 0 {
		return nil, ErrTrackingProviderRequired
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if trackingNumber == "" {
		return nil, ErrTrackingNumberRequired
	}

	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	if providerCarrierCode == "" {
		return nil, ErrTrackingCarrierCodeRequired
	}

	shipment := &shipping.TrackingShipment{
		OrderID:                  input.OrderID,
		TrackingProviderID:       input.TrackingProviderID,
		TrackingNumber:           trackingNumber,
		ProviderCarrierCode:      providerCarrierCode,
		CarrierID:                input.CarrierID,
		CarrierServiceID:         input.CarrierServiceID,
		TrackingCarrierMappingID: input.TrackingCarrierMappingID,
		RegistrationStatus:       trackingRegistrationPending,
		SyncStatus:               trackingSyncPending,
		EventCount:               0,
		LastError:                "",
		Enabled:                  true,
	}
	if err := s.shippingRepo.UpsertTrackingShipment(shipment); err != nil {
		return nil, err
	}
	return s.GetTrackingShipmentByOrderID(input.OrderID)
}

func (s *ShippingService) UpsertAndMaybeRegisterTrackingShipment(ctx context.Context, input TrackingShipmentInput) (*shipping.TrackingShipment, error) {
	shipment, err := s.UpsertTrackingShipment(input)
	if err != nil {
		return nil, err
	}

	provider, err := s.GetTrackingProviderConfig(input.TrackingProviderID)
	if err != nil {
		return nil, err
	}
	if !provider.AutoRegister {
		return shipment, nil
	}

	if err := s.RegisterTrackingShipment(ctx, TrackingSyncInput{
		OrderID:                  input.OrderID,
		ProviderID:               input.TrackingProviderID,
		TrackingNumber:           input.TrackingNumber,
		ProviderCarrierCode:      input.ProviderCarrierCode,
		CarrierID:                input.CarrierID,
		CarrierServiceID:         input.CarrierServiceID,
		TrackingCarrierMappingID: input.TrackingCarrierMappingID,
	}); err != nil {
		return nil, err
	}
	return s.GetTrackingShipmentByOrderID(input.OrderID)
}

func (s *ShippingService) RegisterTrackingShipment(ctx context.Context, input TrackingSyncInput) error {
	if input.OrderID == 0 {
		return ErrTrackingOrderRequired
	}
	if input.ProviderID == 0 {
		return ErrTrackingProviderRequired
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if trackingNumber == "" {
		return ErrTrackingNumberRequired
	}
	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	if providerCarrierCode == "" {
		return ErrTrackingCarrierCodeRequired
	}

	existing, err := s.GetTrackingShipmentByOrderID(input.OrderID)
	if err == nil &&
		existing.TrackingProviderID == input.ProviderID &&
		existing.TrackingNumber == trackingNumber &&
		existing.ProviderCarrierCode == providerCarrierCode &&
		existing.RegistrationStatus == trackingRegistrationSynced {
		return nil
	}
	if err != nil && !repository.IsRecordNotFound(err) {
		return err
	}

	client, err := s.NewTrackingClientForProvider(input.ProviderID)
	if err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentRegistrationStatus(input.OrderID, trackingRegistrationFailed, err.Error())
		return err
	}

	registrar, ok := client.(tracking.TrackingRegistrar)
	if !ok {
		return nil
	}

	if err := registrar.RegisterTrackings(ctx, []tracking.TrackingRequest{{TrackingNumber: trackingNumber, Carrier: providerCarrierCode}}); err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentRegistrationStatus(input.OrderID, trackingRegistrationFailed, err.Error())
		return err
	}

	return s.shippingRepo.UpdateTrackingShipmentRegistrationStatus(input.OrderID, trackingRegistrationSynced, "")
}

func (s *ShippingService) SyncTracking(ctx context.Context, input TrackingSyncInput) (*TrackingSyncResult, error) {
	if input.OrderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	if strings.TrimSpace(input.TrackingNumber) == "" {
		return nil, ErrTrackingNumberRequired
	}
	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	if providerCarrierCode == "" {
		return nil, ErrTrackingCarrierCodeRequired
	}

	provider, err := s.GetTrackingProviderConfig(input.ProviderID)
	if err != nil {
		return nil, err
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if _, err := s.ensureTrackingShipmentForSync(input, trackingNumber, providerCarrierCode); err != nil {
		return nil, err
	}
	if provider.AutoRegister {
		if err := s.RegisterTrackingShipment(ctx, input); err != nil {
			_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
			return nil, err
		}
	}
	if err := s.shippingRepo.UpdateTrackingShipmentSyncing(input.OrderID); err != nil {
		return nil, err
	}

	client, err := newTrackingClientFromProvider(provider)
	if err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
		return nil, err
	}

	info, err := client.Track(ctx, trackingNumber, providerCarrierCode)
	if err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
		return nil, err
	}

	events := trackingInfoToDomainEvents(input.OrderID, trackingNumber, providerCarrierCode, info)
	if err := s.shippingRepo.ReplaceTrackingEvents(input.OrderID, trackingNumber, events); err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
		return nil, err
	}
	if err := s.shippingRepo.UpdateTrackingShipmentSyncSuccess(input.OrderID, len(events), latestTrackingEventTime(events), nextTrackingSyncAt(provider, time.Now())); err != nil {
		return nil, err
	}
	shipment, err := s.GetTrackingShipmentByOrderID(input.OrderID)
	if err != nil {
		return nil, err
	}

	result := &TrackingSyncResult{
		TrackingNumber: trackingNumber,
		Carrier:        providerCarrierCode,
		Events:         events,
		Shipment:       shipment,
	}
	if info != nil {
		result.TrackingNumber = info.TrackingNumber
		if result.TrackingNumber == "" {
			result.TrackingNumber = trackingNumber
		}
		if strings.TrimSpace(info.Carrier) != "" {
			result.Carrier = info.Carrier
		}
		result.Status = info.Status
		result.StatusCode = info.StatusCode
		result.UpdatedAt = info.UpdatedAt
	}

	return result, nil
}

func (s *ShippingService) ListTrackingCarrierMappings(enabledOnly bool) ([]shipping.TrackingCarrierMapping, error) {
	return s.shippingRepo.FindAllTrackingCarrierMappings(enabledOnly)
}

func (s *ShippingService) GetTrackingCarrierMapping(id uint) (*shipping.TrackingCarrierMapping, error) {
	return s.shippingRepo.FindTrackingCarrierMappingByID(id)
}

func (s *ShippingService) CreateTrackingCarrierMapping(mapping *shipping.TrackingCarrierMapping) error {
	return s.shippingRepo.CreateTrackingCarrierMapping(mapping)
}

func (s *ShippingService) UpdateTrackingCarrierMapping(mapping *shipping.TrackingCarrierMapping) error {
	return s.shippingRepo.UpdateTrackingCarrierMapping(mapping)
}

func (s *ShippingService) DeleteTrackingCarrierMapping(id uint) error {
	return s.shippingRepo.DeleteTrackingCarrierMapping(id)
}

func (s *ShippingService) ResolveTrackingCarrier(input TrackingCarrierResolutionInput) (*TrackingCarrierResolution, error) {
	if input.ProviderID == 0 {
		return nil, ErrTrackingProviderRequired
	}
	if !hasPositiveID(input.CarrierID) && !hasPositiveID(input.CarrierServiceID) {
		return nil, ErrTrackingLocalTargetRequired
	}

	provider, err := s.GetTrackingProviderConfig(input.ProviderID)
	if err != nil {
		return nil, err
	}
	if !provider.Enabled {
		return nil, fmt.Errorf("%w: %s", ErrTrackingProviderDisabled, provider.ProviderName)
	}

	var carrier *shipping.Carrier
	var carrierID *uint
	if hasPositiveID(input.CarrierID) {
		carrierID = input.CarrierID
	}

	var carrierService *shipping.CarrierService
	if hasPositiveID(input.CarrierServiceID) {
		carrierService, err = s.GetCarrierService(*input.CarrierServiceID)
		if err != nil {
			return nil, err
		}
		if !carrierService.Enabled {
			return nil, fmt.Errorf("%w: %s", ErrTrackingCarrierServiceDisabled, carrierService.ServiceName)
		}
		if carrierService.Carrier != nil {
			carrier = carrierService.Carrier
		}
		serviceCarrierID := carrierService.CarrierID
		carrierID = &serviceCarrierID

		mapping, err := s.shippingRepo.FindEnabledTrackingCarrierMappingByCarrierService(provider.ID, carrierService.ID)
		if err == nil {
			return trackingCarrierResolution(provider, carrier, carrierService, mapping), nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	if hasPositiveID(carrierID) {
		if carrier == nil || carrier.ID != *carrierID {
			carrier, err = s.GetCarrier(*carrierID)
			if err != nil {
				return nil, err
			}
		}
		if !carrier.Enabled {
			return nil, fmt.Errorf("%w: %s", ErrTrackingCarrierDisabled, carrier.Name)
		}

		mapping, err := s.shippingRepo.FindEnabledTrackingCarrierMappingByCarrier(provider.ID, carrier.ID)
		if err == nil {
			return trackingCarrierResolution(provider, carrier, carrierService, mapping), nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	return nil, ErrTrackingCarrierMappingMissing
}

func (s *ShippingService) ListCarrierServices(enabledOnly bool) ([]shipping.CarrierService, error) {
	return s.shippingRepo.FindAllCarrierServices(enabledOnly)
}

func (s *ShippingService) GetCarrierService(id uint) (*shipping.CarrierService, error) {
	return s.shippingRepo.FindCarrierServiceByID(id)
}

func (s *ShippingService) CreateCarrierService(service *shipping.CarrierService) error {
	return s.shippingRepo.CreateCarrierService(service)
}

func (s *ShippingService) UpdateCarrierService(service *shipping.CarrierService) error {
	return s.shippingRepo.UpdateCarrierService(service)
}

func (s *ShippingService) DeleteCarrierService(id uint) error {
	return s.shippingRepo.DeleteCarrierService(id)
}

func (s *ShippingService) ListZones() ([]shipping.ShippingZone, error) {
	return s.shippingRepo.FindAllZones()
}

func (s *ShippingService) GetZone(id uint) (*shipping.ShippingZone, error) {
	return s.shippingRepo.FindZoneByID(id)
}

func (s *ShippingService) CreateZone(zone *shipping.ShippingZone) error {
	return s.shippingRepo.CreateZone(zone)
}

func (s *ShippingService) UpdateZone(zone *shipping.ShippingZone) error {
	return s.shippingRepo.UpdateZone(zone)
}

func (s *ShippingService) DeleteZone(id uint) error {
	return s.shippingRepo.DeleteZone(id)
}

func (s *ShippingService) GetTrackingEventsByTrackingNumber(trackingNumber string) ([]shipping.TrackingEvent, error) {
	return s.shippingRepo.FindTrackingEventsByTrackingNumber(trackingNumber)
}

func (s *ShippingService) GetTrackingEventsByOrderID(orderID uint) ([]shipping.TrackingEvent, error) {
	return s.shippingRepo.FindTrackingEventsByOrderID(orderID)
}

func (s *ShippingService) ListPackagingRules() ([]shipping.PackagingRule, error) {
	return s.shippingRepo.FindAllPackagingRules()
}

func (s *ShippingService) GetPackagingRule(id uint) (*shipping.PackagingRule, error) {
	return s.shippingRepo.FindPackagingRuleByID(id)
}

func (s *ShippingService) CreatePackagingRule(rule *shipping.PackagingRule) error {
	return s.shippingRepo.CreatePackagingRule(rule)
}

func (s *ShippingService) UpdatePackagingRule(rule *shipping.PackagingRule) error {
	return s.shippingRepo.UpdatePackagingRule(rule)
}

func (s *ShippingService) DeletePackagingRule(id uint) error {
	return s.shippingRepo.DeletePackagingRule(id)
}

func (s *ShippingService) CreatePackagingRuleApply(apply *shipping.PackagingRuleApply) error {
	if apply == nil {
		return errors.New("packaging rule apply is required")
	}
	if apply.RuleID == 0 {
		return errors.New("packaging rule id is required")
	}
	if apply.ProductID == 0 {
		return errors.New("product id is required")
	}

	if _, err := s.shippingRepo.FindPackagingRuleByID(apply.RuleID); err != nil {
		if repository.IsRecordNotFound(err) {
			return fmt.Errorf("packaging rule ID %d does not exist", apply.RuleID)
		}
		return err
	}

	if s.productRepo != nil {
		if _, err := s.productRepo.FindByID(apply.ProductID); err != nil {
			if repository.IsRecordNotFound(err) {
				return fmt.Errorf("product ID %d does not exist", apply.ProductID)
			}
			return err
		}
	}

	existing, err := s.shippingRepo.FindPackagingRuleApplyByProductID(apply.ProductID)
	if err == nil && existing != nil && existing.ID > 0 {
		if existing.RuleID == apply.RuleID {
			return errors.New("packaging rule already applies to this product")
		}
		return errors.New("product already has a packaging rule")
	}
	if err != nil && !repository.IsRecordNotFound(err) {
		return err
	}

	return s.shippingRepo.CreatePackagingRuleApply(apply)
}

func (s *ShippingService) DeletePackagingRuleApply(id uint) error {
	return s.shippingRepo.DeletePackagingRuleApply(id)
}

func (s *ShippingService) GetProductPackagingRules(productID uint) ([]shipping.PackagingRule, error) {
	return s.shippingRepo.FindPackagingRulesByProductID(productID)
}

func (s *ShippingService) quoteCarrierServiceOptions(
	country string,
	currency string,
	resolvedItems []resolvedShippingItem,
	groups map[uint]*shippingQuoteGroup,
	cartAmount float64,
) ([]ShippingQuoteOption, error) {
	group := singleShippingQuoteGroup(groups)
	if group == nil {
		return nil, nil
	}

	carrierServices, err := s.shippingRepo.FindEnabledCarrierServicesWithTemplates()
	if err != nil {
		return nil, err
	}

	options := make([]ShippingQuoteOption, 0, len(carrierServices))
	for i := range carrierServices {
		option, ok := buildCarrierServiceQuoteOption(carrierServices[i], group, resolvedItems, country, currency, cartAmount)
		if ok {
			options = append(options, option)
		}
	}

	sort.SliceStable(options, func(i, j int) bool {
		if options[i].ShippingFee != options[j].ShippingFee {
			return options[i].ShippingFee < options[j].ShippingFee
		}
		if options[i].SortOrder != options[j].SortOrder {
			return options[i].SortOrder < options[j].SortOrder
		}
		return options[i].CarrierServiceID < options[j].CarrierServiceID
	})

	return options, nil
}

func buildCarrierServiceQuoteOption(
	service shipping.CarrierService,
	group *shippingQuoteGroup,
	resolvedItems []resolvedShippingItem,
	country string,
	currency string,
	cartAmount float64,
) (ShippingQuoteOption, bool) {
	if group == nil || service.Template == nil || !service.Enabled || !service.Template.Enabled {
		return ShippingQuoteOption{}, false
	}
	if service.TemplateID == nil || *service.TemplateID != group.Template.ID {
		return ShippingQuoteOption{}, false
	}
	if service.Carrier == nil || !service.Carrier.Enabled {
		return ShippingQuoteOption{}, false
	}
	if !carrierServiceMatchesCountry(service.Countries, country) {
		return ShippingQuoteOption{}, false
	}

	quoteCurrency := strings.ToUpper(strings.TrimSpace(currency))
	serviceCurrency := strings.ToUpper(strings.TrimSpace(service.Currency))
	if serviceCurrency == "" {
		serviceCurrency = quoteCurrency
	}
	if quoteCurrency != "" && serviceCurrency != "" && serviceCurrency != quoteCurrency {
		return ShippingQuoteOption{}, false
	}

	actualWeightGrams := group.TotalWeightGrams
	volumetricWeightGrams, hasVolumetricWeight := carrierServiceVolumetricWeightGrams(service, resolvedItems)
	chargeWeightGrams := actualWeightGrams
	switch service.BillingMode {
	case "volumetric_weight":
		if !hasVolumetricWeight {
			return ShippingQuoteOption{}, false
		}
		chargeWeightGrams = volumetricWeightGrams
	case "greater_of_actual_and_volumetric":
		if !hasVolumetricWeight {
			return ShippingQuoteOption{}, false
		}
		chargeWeightGrams = maxInt(actualWeightGrams, volumetricWeightGrams)
	default:
		chargeWeightGrams = actualWeightGrams
	}

	billableWeightGrams := carrierServiceBillableWeightGrams(chargeWeightGrams, service)
	baseFee, freeShipping := calculateTemplateShippingFee(
		service.Template,
		country,
		billableWeightGrams,
		group.Quantity,
		group.Amount,
		cartAmount,
	)

	fuelSurcharge := 0.0
	remoteSurcharge := 0.0
	shippingFee := baseFee
	if freeShipping {
		shippingFee = 0
	} else {
		fuelSurcharge = roundMoney(baseFee * service.FuelSurchargePercent / 100)
		remoteSurcharge = roundMoney(service.RemoteSurcharge)
		shippingFee = roundMoney(baseFee + fuelSurcharge + remoteSurcharge)
	}

	return ShippingQuoteOption{
		CarrierID:             service.Carrier.ID,
		CarrierName:           service.Carrier.Name,
		CarrierCode:           service.Carrier.Code,
		CarrierServiceID:      service.ID,
		ServiceCode:           service.ServiceCode,
		ServiceName:           service.ServiceName,
		RouteName:             service.RouteName,
		TemplateID:            service.Template.ID,
		TemplateName:          service.Template.Name,
		Currency:              serviceCurrency,
		BillingMode:           service.BillingMode,
		ActualWeightGrams:     actualWeightGrams,
		VolumetricWeightGrams: volumetricWeightGrams,
		ChargeWeightGrams:     chargeWeightGrams,
		BillableWeightGrams:   billableWeightGrams,
		BaseFee:               baseFee,
		FuelSurcharge:         fuelSurcharge,
		RemoteSurcharge:       remoteSurcharge,
		ShippingFee:           shippingFee,
		FreeShipping:          freeShipping,
		EtaMinDays:            service.EtaMinDays,
		EtaMaxDays:            service.EtaMaxDays,
		SortOrder:             service.SortOrder,
	}, true
}

func singleShippingQuoteGroup(groups map[uint]*shippingQuoteGroup) *shippingQuoteGroup {
	if len(groups) != 1 {
		return nil
	}
	for _, group := range groups {
		return group
	}
	return nil
}

func carrierServiceMatchesCountry(countriesValue string, country string) bool {
	normalizedCountry := strings.ToUpper(strings.TrimSpace(country))
	if normalizedCountry == "" {
		return false
	}

	countries := normalizeShippingRegions(countriesValue)
	if len(countries) == 0 {
		return true
	}

	for _, candidate := range countries {
		switch candidate {
		case "*", "ALL", "GLOBAL", "WORLDWIDE":
			return true
		default:
			if candidate == normalizedCountry {
				return true
			}
		}
	}
	return false
}

func carrierServiceVolumetricWeightGrams(service shipping.CarrierService, resolvedItems []resolvedShippingItem) (int, bool) {
	if service.VolumetricDivisor <= 0 {
		return 0, false
	}

	total := 0
	for _, item := range resolvedItems {
		itemVolumetricWeight, ok := packagingRuleVolumetricWeightGrams(item.PackagingRule, service.VolumetricDivisor)
		if !ok {
			return 0, false
		}
		total += itemVolumetricWeight * item.Quantity
	}
	return total, total > 0
}

func packagingRuleVolumetricWeightGrams(rule *shipping.PackagingRule, divisor int) (int, bool) {
	if rule == nil || divisor <= 0 || rule.BoxLength <= 0 || rule.BoxWidth <= 0 || rule.BoxHeight <= 0 {
		return 0, false
	}
	volumetricKg := rule.BoxLength * rule.BoxWidth * rule.BoxHeight / float64(divisor)
	if volumetricKg <= 0 {
		return 0, false
	}
	return int(math.Ceil(volumetricKg * 1000)), true
}

func carrierServiceBillableWeightGrams(chargeWeightGrams int, service shipping.CarrierService) int {
	billableWeightGrams := maxInt(chargeWeightGrams, service.MinChargeWeightGrams)
	if billableWeightGrams <= 0 {
		return 0
	}

	firstWeightGrams := service.FirstWeightGrams
	additionalWeightGrams := service.AdditionalWeightGrams
	if firstWeightGrams > 0 {
		if billableWeightGrams <= firstWeightGrams {
			return firstWeightGrams
		}
		if additionalWeightGrams > 0 {
			return firstWeightGrams + ceilDivInt(billableWeightGrams-firstWeightGrams, additionalWeightGrams)*additionalWeightGrams
		}
		return billableWeightGrams
	}
	if additionalWeightGrams > 0 {
		return ceilDivInt(billableWeightGrams, additionalWeightGrams) * additionalWeightGrams
	}
	return billableWeightGrams
}

func ceilDivInt(value int, divisor int) int {
	if value <= 0 {
		return 0
	}
	return (value + divisor - 1) / divisor
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func uniqueShippingQuoteProductIDs(items []ShippingQuoteItemInput) []uint {
	seen := make(map[uint]struct{})
	productIDs := make([]uint, 0, len(items))
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
	return productIDs
}

func packagingRuleWeightGrams(rule *shipping.PackagingRule) int {
	if rule == nil || rule.BoxWeight <= 0 {
		return 0
	}
	return int(math.Round(rule.BoxWeight * 1000))
}

func selectShippingTemplateForItem(bindings []shipping.ShippingTemplateBinding, item ShippingQuoteItemInput) *shipping.ShippingTemplate {
	var selected *shipping.ShippingTemplateBinding
	selectedScore := 0

	for i := range bindings {
		binding := &bindings[i]
		if !binding.Enabled || binding.Template == nil || !binding.Template.Enabled {
			continue
		}

		score := shippingBindingMatchScore(binding, item)
		if score == 0 {
			continue
		}

		if selected == nil ||
			score > selectedScore ||
			(score == selectedScore && binding.Priority > selected.Priority) ||
			(score == selectedScore && binding.Priority == selected.Priority && binding.ID > selected.ID) {
			selected = binding
			selectedScore = score
		}
	}

	if selected == nil {
		return nil
	}
	return selected.Template
}

func shippingBindingMatchScore(binding *shipping.ShippingTemplateBinding, item ShippingQuoteItemInput) int {
	switch binding.Scope {
	case "variant":
		if item.VariantID != nil && binding.VariantID != nil && *item.VariantID == *binding.VariantID {
			return 4
		}
	case "product":
		if binding.ProductID != nil && item.ProductID == *binding.ProductID {
			return 3
		}
	case "product_type":
		if item.ProductTypeID != nil && binding.ProductTypeID != nil && *item.ProductTypeID == *binding.ProductTypeID {
			return 2
		}
	case "default":
		return 1
	}
	return 0
}

func calculateTemplateShippingFee(
	template *shipping.ShippingTemplate,
	country string,
	totalWeightGrams int,
	quantity int,
	amount float64,
	cartAmount float64,
) (float64, bool) {
	if template == nil {
		return 0, false
	}

	if template.FreeShipping && cartAmount >= template.FreeThreshold {
		return 0, true
	}

	value := float64(totalWeightGrams) / 1000
	switch template.Type {
	case "quantity", "items":
		value = float64(quantity)
	case "price", "amount":
		value = amount
	}

	shippingFee := template.DefaultFee
	for _, rule := range template.Rules {
		if shippingRuleMatchesCountry(rule.Region, country) && shippingRuleMatchesValue(rule, value) {
			shippingFee = calculateRuleFee(rule, value)
			break
		}
	}

	return roundMoney(shippingFee), false
}

func shippingRuleMatchesValue(rule shipping.ShippingRule, value float64) bool {
	return value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue)
}

func calculateRuleFee(rule shipping.ShippingRule, value float64) float64 {
	fee := rule.Fee
	if rule.Additional > 0 && value > rule.MinValue {
		fee += math.Ceil(value-rule.MinValue) * rule.Additional
	}
	return fee
}

func shippingRuleMatchesCountry(region string, country string) bool {
	normalizedCountry := strings.ToUpper(strings.TrimSpace(country))
	if normalizedCountry == "" {
		return false
	}

	regions := normalizeShippingRegions(region)
	if len(regions) == 0 {
		return true
	}

	for _, candidate := range regions {
		switch candidate {
		case "*", "ALL", "GLOBAL", "WORLDWIDE":
			return true
		default:
			if candidate == normalizedCountry {
				return true
			}
		}
	}
	return false
}

func normalizeShippingRegions(region string) []string {
	trimmed := strings.TrimSpace(region)
	if trimmed == "" {
		return nil
	}

	var parsed []string
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		return normalizeRegionList(parsed)
	}

	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\r' || r == '\t'
	})
	return normalizeRegionList(parts)
}

func normalizeRegionList(regions []string) []string {
	normalized := make([]string, 0, len(regions))
	for _, region := range regions {
		candidate := strings.ToUpper(strings.TrimSpace(region))
		if candidate != "" {
			normalized = append(normalized, candidate)
		}
	}
	return normalized
}

func distributeGroupFee(
	group *shippingQuoteGroup,
	resolvedItems []resolvedShippingItem,
	quoteItems []ShippingQuoteItem,
	groupFee float64,
	groupFree bool,
) {
	if group == nil || len(group.ItemIndexes) == 0 {
		return
	}

	var totalBasis float64
	for _, index := range group.ItemIndexes {
		totalBasis += shippingFeeDistributionBasis(group.Template.Type, resolvedItems[index])
	}

	remaining := roundMoney(groupFee)
	for position, index := range group.ItemIndexes {
		itemFee := 0.0
		if position == len(group.ItemIndexes)-1 {
			itemFee = remaining
		} else if totalBasis > 0 {
			itemFee = roundMoney(groupFee * shippingFeeDistributionBasis(group.Template.Type, resolvedItems[index]) / totalBasis)
			remaining = roundMoney(remaining - itemFee)
		} else {
			itemFee = roundMoney(groupFee / float64(len(group.ItemIndexes)))
			remaining = roundMoney(remaining - itemFee)
		}

		quoteItems[index].ShippingFee = itemFee
		quoteItems[index].FreeShipping = groupFree
	}
}

func shippingFeeDistributionBasis(templateType string, item resolvedShippingItem) float64 {
	switch templateType {
	case "quantity", "items":
		return float64(item.Quantity)
	case "price", "amount":
		return item.Amount
	default:
		return float64(item.ChargeWeightGrams * item.Quantity)
	}
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func hasPositiveID(id *uint) bool {
	return id != nil && *id > 0
}

func uintPtr(id uint) *uint {
	return &id
}

func (s *ShippingService) ensureTrackingShipmentForSync(input TrackingSyncInput, trackingNumber string, providerCarrierCode string) (*shipping.TrackingShipment, error) {
	existing, err := s.GetTrackingShipmentByOrderID(input.OrderID)
	if err == nil && existing.TrackingProviderID == input.ProviderID && existing.TrackingNumber == trackingNumber && existing.ProviderCarrierCode == providerCarrierCode {
		return existing, nil
	}
	if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}

	carrierID := input.CarrierID
	carrierServiceID := input.CarrierServiceID
	trackingCarrierMappingID := input.TrackingCarrierMappingID
	if err == nil {
		if !hasPositiveID(carrierID) {
			carrierID = existing.CarrierID
		}
		if !hasPositiveID(carrierServiceID) {
			carrierServiceID = existing.CarrierServiceID
		}
		if !hasPositiveID(trackingCarrierMappingID) {
			trackingCarrierMappingID = existing.TrackingCarrierMappingID
		}
	}

	return s.UpsertTrackingShipment(TrackingShipmentInput{
		OrderID:                  input.OrderID,
		TrackingProviderID:       input.ProviderID,
		TrackingNumber:           trackingNumber,
		ProviderCarrierCode:      providerCarrierCode,
		CarrierID:                carrierID,
		CarrierServiceID:         carrierServiceID,
		TrackingCarrierMappingID: trackingCarrierMappingID,
	})
}

func latestTrackingEventTime(events []shipping.TrackingEvent) *time.Time {
	var latest time.Time
	for _, event := range events {
		if event.EventTime.After(latest) {
			latest = event.EventTime
		}
	}
	if latest.IsZero() {
		return nil
	}
	return &latest
}

func nextTrackingSyncAt(provider *shipping.TrackingProviderConfig, now time.Time) *time.Time {
	if provider == nil || !provider.PollingEnabled {
		return nil
	}

	intervalMinutes := provider.PollingIntervalMinutes
	if intervalMinutes <= 0 {
		intervalMinutes = defaultTrackingPollingIntervalMinutes
	}

	next := now.Add(time.Duration(intervalMinutes) * time.Minute)
	return &next
}

func newTrackingClientFromProvider(provider *shipping.TrackingProviderConfig) (tracking.TrackingService, error) {
	if provider == nil || provider.ID == 0 {
		return nil, ErrTrackingProviderRequired
	}
	if !provider.Enabled {
		return nil, fmt.Errorf("%w: %s", ErrTrackingProviderDisabled, provider.ProviderName)
	}

	providerCode := strings.ToLower(strings.TrimSpace(provider.ProviderCode))
	if providerCode == "mock" {
		return tracking.NewMockTrackingService(), nil
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, ErrTrackingProviderAPIKeyMissing
	}

	baseURL := strings.TrimRight(strings.TrimSpace(provider.BaseURL), "/")
	if baseURL == "" {
		return nil, ErrTrackingProviderBaseURLMissing
	}

	timeoutSeconds := provider.RequestTimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 15
	}

	switch providerCode {
	case "17track", "track17":
		return tracking.NewTrackingService(&tracking.Config{
			Provider: provider.ProviderCode,
			APIKey:   apiKey,
			BaseURL:  baseURL,
			Timeout:  time.Duration(timeoutSeconds) * time.Second,
		}), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrTrackingProviderUnsupported, provider.ProviderCode)
	}
}

func trackingInfoToDomainEvents(orderID uint, trackingNumber string, carrierCode string, info *tracking.TrackingInfo) []shipping.TrackingEvent {
	if info == nil {
		return nil
	}

	effectiveTrackingNumber := strings.TrimSpace(info.TrackingNumber)
	if effectiveTrackingNumber == "" {
		effectiveTrackingNumber = trackingNumber
	}
	effectiveCarrier := strings.TrimSpace(info.Carrier)
	if effectiveCarrier == "" {
		effectiveCarrier = carrierCode
	}

	events := make([]shipping.TrackingEvent, 0, len(info.Events))
	syncTime := time.Now()
	for _, event := range info.Events {
		eventTime := event.Time
		if eventTime.IsZero() {
			eventTime = syncTime
		}
		events = append(events, shipping.TrackingEvent{
			OrderID:             orderID,
			TrackingNumber:      effectiveTrackingNumber,
			ProviderCarrierCode: effectiveCarrier,
			Status:              strings.TrimSpace(event.Status),
			Location:            trackingLocationString(event.Location),
			Description:         strings.TrimSpace(event.Description),
			EventTime:           eventTime,
		})
	}

	if len(events) == 0 && strings.TrimSpace(info.Status) != "" {
		eventTime := info.UpdatedAt
		if eventTime.IsZero() {
			eventTime = syncTime
		}
		events = append(events, shipping.TrackingEvent{
			OrderID:             orderID,
			TrackingNumber:      effectiveTrackingNumber,
			ProviderCarrierCode: effectiveCarrier,
			Status:              strings.TrimSpace(info.Status),
			Description:         strings.TrimSpace(info.Status),
			EventTime:           eventTime,
		})
	}

	return events
}

func trackingWebhookEventsToDomainEvents(orderID uint, trackingNumber string, carrierCode string, input TrackingWebhookInput) []shipping.TrackingEvent {
	events := make([]shipping.TrackingEvent, 0, len(input.Events))
	syncTime := time.Now()
	for _, event := range input.Events {
		eventTime := event.EventTime
		if eventTime.IsZero() {
			eventTime = syncTime
		}
		events = append(events, shipping.TrackingEvent{
			OrderID:             orderID,
			TrackingNumber:      trackingNumber,
			ProviderCarrierCode: carrierCode,
			Status:              strings.TrimSpace(event.Status),
			Location:            strings.TrimSpace(event.Location),
			Description:         strings.TrimSpace(event.Description),
			EventTime:           eventTime,
		})
	}

	if len(events) == 0 && strings.TrimSpace(input.Status) != "" {
		events = append(events, shipping.TrackingEvent{
			OrderID:             orderID,
			TrackingNumber:      trackingNumber,
			ProviderCarrierCode: carrierCode,
			Status:              strings.TrimSpace(input.Status),
			Description:         strings.TrimSpace(input.Status),
			EventTime:           syncTime,
		})
	}

	return events
}

func trackingLocationString(location *tracking.Location) string {
	if location == nil {
		return ""
	}

	parts := []string{}
	for _, part := range []string{location.City, location.State, location.Country, location.ZipCode} {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, ", ")
}

func trackingCarrierResolution(
	provider *shipping.TrackingProviderConfig,
	carrier *shipping.Carrier,
	carrierService *shipping.CarrierService,
	mapping *shipping.TrackingCarrierMapping,
) *TrackingCarrierResolution {
	return &TrackingCarrierResolution{
		Provider:            provider,
		Carrier:             carrier,
		CarrierService:      carrierService,
		Mapping:             mapping,
		ProviderCarrierCode: mapping.ProviderCarrierCode,
		ProviderCarrierName: mapping.ProviderCarrierName,
	}
}
