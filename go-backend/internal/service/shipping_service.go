package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/resilience"
	"commerce-platform/internal/repository"
	"errors"
	"sync"
	"time"
)

type ShippingService struct {
	shippingRepo    *repository.ShippingRepository
	productRepo     *repository.ProductRepository
	orderRepo       *repository.OrderRepository
	currencyPolicy  *CurrencyPolicyService
	trackingRun     TrackingPollingRunState
	webhookRun      TrackingWebhookRunState
	trackingMu      sync.RWMutex
	trackingRetry   resilience.HTTPRetryPolicy
	trackingBreaker resilience.CircuitController
}

type ShippingCalculationInput struct {
	TemplateID uint
	Weight     float64
	Quantity   int
	Amount     float64
	Country    string
}

type ShippingQuoteItemInput struct {
	ProductID                      uint    `json:"product_id"`
	VariantID                      *uint   `json:"variant_id,omitempty"`
	ProductSpecificationTemplateID *uint   `json:"product_specification_template_id,omitempty"`
	ShippingTemplateID             *uint   `json:"shipping_template_id,omitempty"`
	Quantity                       int     `json:"quantity"`
	UnitPrice                      float64 `json:"unit_price"`
	WeightGrams                    int     `json:"weight_grams"`
}

type ShippingQuoteInput struct {
	Country         string                   `json:"country"`
	Amount          float64                  `json:"amount"`
	Currency        string                   `json:"currency,omitempty"`
	DisplayCurrency string                   `json:"display_currency,omitempty"`
	Items           []ShippingQuoteItemInput `json:"items"`
}

type ShippingQuoteItem struct {
	ProductID                      uint    `json:"product_id"`
	VariantID                      *uint   `json:"variant_id,omitempty"`
	ProductSpecificationTemplateID *uint   `json:"product_specification_template_id,omitempty"`
	TemplateID                     uint    `json:"template_id"`
	TemplateName                   string  `json:"template_name"`
	PackagingRuleID                *uint   `json:"packaging_rule_id,omitempty"`
	PackagingRuleName              string  `json:"packaging_rule_name,omitempty"`
	Quantity                       int     `json:"quantity"`
	UnitPrice                      float64 `json:"unit_price"`
	Amount                         float64 `json:"amount"`
	WeightGrams                    int     `json:"weight_grams"`
	PackagingWeightGrams           int     `json:"packaging_weight_grams"`
	ChargeWeightGrams              int     `json:"charge_weight_grams"`
	ShippingFee                    float64 `json:"shipping_fee"`
	FreeShipping                   bool    `json:"free_shipping"`
}

type ShippingQuote struct {
	ShippingFee     float64                         `json:"shipping_fee"`
	FreeShipping    bool                            `json:"free_shipping"`
	Currency        string                          `json:"currency,omitempty"`
	DisplayPrice    *currency.DisplayPriceSnapshot  `json:"display_price,omitempty"`
	DisplayPrices   []currency.DisplayPriceSnapshot `json:"display_prices,omitempty"`
	DisplayCurrency string                          `json:"display_currency,omitempty"`
	Source          string                          `json:"source,omitempty"`
	Items           []ShippingQuoteItem             `json:"items,omitempty"`
	Options         []ShippingQuoteOption           `json:"options,omitempty"`
	SelectedOption  *ShippingQuoteOption            `json:"selected_option,omitempty"`
}

type ShippingQuoteOption struct {
	CarrierID             uint                            `json:"carrier_id"`
	CarrierName           string                          `json:"carrier_name"`
	CarrierCode           string                          `json:"carrier_code"`
	CarrierServiceID      uint                            `json:"carrier_service_id"`
	ServiceCode           string                          `json:"service_code"`
	ServiceName           string                          `json:"service_name"`
	RouteName             string                          `json:"route_name,omitempty"`
	TemplateID            uint                            `json:"template_id"`
	TemplateName          string                          `json:"template_name"`
	Currency              string                          `json:"currency,omitempty"`
	BillingMode           string                          `json:"billing_mode"`
	ActualWeightGrams     int                             `json:"actual_weight_grams"`
	VolumetricWeightGrams int                             `json:"volumetric_weight_grams"`
	ChargeWeightGrams     int                             `json:"charge_weight_grams"`
	BillableWeightGrams   int                             `json:"billable_weight_grams"`
	BaseFee               float64                         `json:"base_fee"`
	FuelSurcharge         float64                         `json:"fuel_surcharge"`
	RemoteSurcharge       float64                         `json:"remote_surcharge"`
	ShippingFee           float64                         `json:"shipping_fee"`
	DisplayPrice          *currency.DisplayPriceSnapshot  `json:"display_price,omitempty"`
	DisplayPrices         []currency.DisplayPriceSnapshot `json:"display_prices,omitempty"`
	FreeShipping          bool                            `json:"free_shipping"`
	EtaMinDays            int                             `json:"eta_min_days"`
	EtaMaxDays            int                             `json:"eta_max_days"`
	SortOrder             int                             `json:"sort_order"`
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
	Status                 string    `json:"status"`
	Location               string    `json:"location"`
	Description            string    `json:"description"`
	RecipientSignatureName string    `json:"recipient_signature_name,omitempty"`
	ProofOfDeliveryURL     string    `json:"proof_of_delivery_url,omitempty"`
	EventTime              time.Time `json:"event_time"`
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

func (s *ShippingService) ConfigureCurrencyPolicy(policy *CurrencyPolicyService) {
	if s == nil {
		return
	}
	s.currencyPolicy = policy
}

func (s *ShippingService) ConfigureOrderRepository(orderRepo *repository.OrderRepository) {
	if s == nil {
		return
	}
	s.orderRepo = orderRepo
}

func (s *ShippingService) ConfigureOutboundTrackingResilience(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) {
	if s == nil {
		return
	}
	s.trackingMu.Lock()
	defer s.trackingMu.Unlock()
	s.trackingRetry = retry
	s.trackingBreaker = breaker
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

func (s *ShippingService) GetProductPackagingRules(productID uint) ([]shipping.PackagingRule, error) {
	return s.shippingRepo.FindPackagingRulesByProductID(productID)
}
