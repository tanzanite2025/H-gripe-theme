package order

import (
	orderdomain "commerce-platform/internal/domain/order"
	"encoding/json"
	"strings"
	"time"
)

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Items                        []OrderItemRequest `json:"items" binding:"required,min=1"`
	ShippingAddress              AddressRequest     `json:"shipping_address" binding:"required"`
	BillingAddress               *AddressRequest    `json:"billing_address,omitempty" binding:"omitempty"`
	PaymentMethod                string             `json:"payment_method" binding:"required"`
	ShippingMethod               string             `json:"shipping_method" binding:"required"`
	ExpectedTotalMinor           *int64             `json:"expected_total_minor" binding:"required"`
	ShippingQuoteID              string             `json:"shipping_quote_id" binding:"required"`
	SelectedQuotePlanID          string             `json:"selected_quote_plan_id" binding:"required"`
	CouponCode                   string             `json:"coupon_code"`
	Notes                        string             `json:"notes"`
	DisplayCurrency              string             `json:"display_currency"`
	PolicyDisclosureAcknowledged bool               `json:"policy_disclosure_acknowledged"`
	ClientRisk                   *ClientRiskRequest `json:"client_risk,omitempty"`
}

type ClientRiskRequest struct {
	IPCountry      string `json:"ip_country,omitempty"`
	BillingCountry string `json:"billing_country,omitempty"` // retained for client compatibility; server ignores it
	VPNDetected    bool   `json:"vpn_detected,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
}

type OrderItemRequest struct {
	ProductID uint  `json:"product_id" binding:"required"`
	VariantID *uint `json:"variant_id"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

type AddressRequest struct {
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Company    string `json:"company"`
	Address1   string `json:"address1" binding:"required"`
	Address2   string `json:"address2"`
	City       string `json:"city" binding:"required"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code" binding:"required"`
	Country    string `json:"country" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
}

// PublicOrderResponse is the customer-facing order contract. It deliberately
// omits database IDs so browser clients use order_number as the only order
// reference.
type PublicOrderResponse struct {
	OrderNumber           string                   `json:"order_number"`
	Status                string                   `json:"status"`
	PaymentMethod         string                   `json:"payment_method"`
	PaymentStatus         string                   `json:"payment_status"`
	ShippingMethod        string                   `json:"shipping_method"`
	ShippingStatus        string                   `json:"shipping_status"`
	FulfillmentMode       string                   `json:"fulfillment_mode"`
	ProductionStatus      string                   `json:"production_status"`
	SignatureRequired     bool                     `json:"signature_required"`
	SubtotalMinor         int64                    `json:"subtotal_minor"`
	ShippingFeeMinor      int64                    `json:"shipping_fee_minor"`
	TaxMinor              int64                    `json:"tax_minor"`
	DiscountMinor         int64                    `json:"discount_minor"`
	TotalMinor            int64                    `json:"total_minor"`
	Currency              string                   `json:"currency"`
	CouponCode            string                   `json:"coupon_code"`
	ShippingAddress       orderdomain.Address      `json:"shipping_address"`
	BillingAddress        orderdomain.Address      `json:"billing_address"`
	CustomerNote          string                   `json:"customer_note"`
	Items                 []PublicOrderItem        `json:"items"`
	CreatedAt             time.Time                `json:"created_at"`
	UpdatedAt             time.Time                `json:"updated_at"`
	PaidAt                *time.Time               `json:"paid_at"`
	ShippedAt             *time.Time               `json:"shipped_at"`
	CompletedAt           *time.Time               `json:"completed_at"`
	CancelledAt           *time.Time               `json:"cancelled_at"`
	ProductionStartedAt   *time.Time               `json:"production_started_at"`
	ProductionCompletedAt *time.Time               `json:"production_completed_at"`
	TrackingShipments     []PublicTrackingShipment `json:"tracking_shipments,omitempty"`
}

type PublicTrackingShipment struct {
	TrackingNumber      string     `json:"tracking_number"`
	ProviderCarrierCode string     `json:"provider_carrier_code"`
	RegistrationStatus  string     `json:"registration_status"`
	SyncStatus          string     `json:"sync_status"`
	LastEventAt         *time.Time `json:"last_event_at,omitempty"`
	Enabled             bool       `json:"enabled"`
}

type PublicOrderItem struct {
	// ItemIndex is the stable zero-based position in the public order payload.
	// It lets customer workflows refer to an order line without exposing the
	// internal database order-item ID.
	ItemIndex       int             `json:"item_index"`
	ProductID       uint            `json:"product_id"`
	VariantID       *uint           `json:"variant_id"`
	ProductName     string          `json:"product_name"`
	SKU             string          `json:"sku"`
	Configuration   json.RawMessage `json:"configuration,omitempty"`
	FulfillmentMode string          `json:"fulfillment_mode"`
	Quantity        int             `json:"quantity"`
	Currency        string          `json:"currency"`
	PriceMinor      int64           `json:"price_minor"`
	SubtotalMinor   int64           `json:"subtotal_minor"`
	DiscountMinor   int64           `json:"discount_minor"`
	TaxMinor        int64           `json:"tax_minor"`
	TotalMinor      int64           `json:"total_minor"`
}

func publicOrderResponse(item orderdomain.Order) PublicOrderResponse {
	fulfillmentMode := orderdomain.NormalizeFulfillmentMode(item.FulfillmentMode)
	if item.FulfillmentMode == "" {
		fulfillmentMode = orderdomain.ResolveFulfillmentMode(item.Items)
	}
	productionStatus := orderdomain.NormalizeProductionStatus(item.ProductionStatus)
	if item.ProductionStatus == "" {
		productionStatus = orderdomain.DefaultProductionStatus(fulfillmentMode)
	}

	subtotalMinor := int64(0)
	if value, err := item.SubtotalMoney(); err == nil {
		subtotalMinor = value.AmountMinor()
	}
	shippingMinor := int64(0)
	if value, err := item.ShippingFeeMoney(); err == nil {
		shippingMinor = value.AmountMinor()
	}
	taxMinor := int64(0)
	if value, err := item.TaxMoney(); err == nil {
		taxMinor = value.AmountMinor()
	}
	discountMinor := int64(0)
	if value, err := item.DiscountMoney(); err == nil {
		discountMinor = value.AmountMinor()
	}
	totalMinor := int64(0)
	if value, err := item.TotalMoney(); err == nil {
		totalMinor = value.AmountMinor()
	}
	return PublicOrderResponse{
		OrderNumber:           item.OrderNumber,
		Status:                item.Status,
		PaymentMethod:         item.PaymentMethod,
		PaymentStatus:         item.PaymentStatus,
		ShippingMethod:        item.ShippingMethod,
		ShippingStatus:        item.ShippingStatus,
		FulfillmentMode:       fulfillmentMode,
		ProductionStatus:      productionStatus,
		SignatureRequired:     item.SignatureRequired,
		SubtotalMinor:         subtotalMinor,
		ShippingFeeMinor:      shippingMinor,
		TaxMinor:              taxMinor,
		DiscountMinor:         discountMinor,
		TotalMinor:            totalMinor,
		Currency:              item.Currency,
		CouponCode:            item.CouponCode,
		ShippingAddress:       item.ShippingAddress,
		BillingAddress:        item.BillingAddress,
		CustomerNote:          item.CustomerNote,
		Items:                 publicOrderItems(item.Items),
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
		PaidAt:                item.PaidAt,
		ShippedAt:             item.ShippedAt,
		CompletedAt:           item.CompletedAt,
		CancelledAt:           item.CancelledAt,
		ProductionStartedAt:   item.ProductionStartedAt,
		ProductionCompletedAt: item.ProductionCompletedAt,
	}
}

func publicOrderResponses(items []orderdomain.Order) []PublicOrderResponse {
	result := make([]PublicOrderResponse, 0, len(items))
	for _, item := range items {
		result = append(result, publicOrderResponse(item))
	}
	return result
}

func publicOrderItems(items []orderdomain.OrderItem) []PublicOrderItem {
	result := make([]PublicOrderItem, 0, len(items))
	for itemIndex, item := range items {
		priceMinor := int64(0)
		if value, err := item.PriceMoney(); err == nil {
			priceMinor = value.AmountMinor()
		}
		subtotalMinor := int64(0)
		if value, err := item.SubtotalMoney(); err == nil {
			subtotalMinor = value.AmountMinor()
		}
		discountMinor := int64(0)
		if value, err := item.DiscountMoney(); err == nil {
			discountMinor = value.AmountMinor()
		}
		taxMinor := int64(0)
		if value, err := item.TaxAmountMoney(); err == nil {
			taxMinor = value.AmountMinor()
		}
		totalMinor := int64(0)
		if value, err := item.TotalMoney(); err == nil {
			totalMinor = value.AmountMinor()
		}
		result = append(result, PublicOrderItem{
			ItemIndex:       itemIndex,
			ProductID:       item.ProductID,
			VariantID:       item.VariantID,
			ProductName:     item.ProductName,
			SKU:             item.SKU,
			Configuration:   append(json.RawMessage(nil), item.ConfigurationSnapshotData...),
			FulfillmentMode: item.FulfillmentMode,
			Quantity:        item.Quantity,
			Currency:        item.Currency,
			PriceMinor:      priceMinor,
			SubtotalMinor:   subtotalMinor,
			DiscountMinor:   discountMinor,
			TaxMinor:        taxMinor,
			TotalMinor:      totalMinor,
		})
	}
	return result
}

func addressFromRequest(req AddressRequest) orderdomain.Address {
	return orderdomain.Address{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Company:    req.Company,
		Address1:   req.Address1,
		Address2:   req.Address2,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		Phone:      req.Phone,
		Email:      req.Email,
	}
}

func billingCountryFromRequest(shippingCountry string, req *AddressRequest) string {
	if req != nil {
		if country := strings.TrimSpace(req.Country); country != "" {
			return country
		}
	}
	return strings.TrimSpace(shippingCountry)
}

func billingAddressFromRequest(shippingAddr orderdomain.Address, req *AddressRequest) orderdomain.Address {
	if req == nil {
		return shippingAddr
	}
	return addressFromRequest(*req)
}
