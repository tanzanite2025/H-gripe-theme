package order

import (
	orderdomain "commerce-platform/internal/domain/order"
	"time"
)

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Items           []OrderItemRequest `json:"items" binding:"required,min=1"`
	ShippingAddress AddressRequest     `json:"shipping_address" binding:"required"`
	BillingAddress  AddressRequest     `json:"billing_address"`
	PaymentMethod   string             `json:"payment_method" binding:"required"`
	ShippingMethod  string             `json:"shipping_method" binding:"required"`
	CouponCode      string             `json:"coupon_code"`
	PointsToUse     int                `json:"points_to_use"`
	ClientRisk      *ClientRiskRequest `json:"client_risk,omitempty"`
}

type ClientRiskRequest struct {
	IPCountry      string `json:"ip_country,omitempty"`
	BillingCountry string `json:"billing_country,omitempty"`
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
	OrderNumber         string              `json:"order_number"`
	Status              string              `json:"status"`
	PaymentMethod       string              `json:"payment_method"`
	PaymentStatus       string              `json:"payment_status"`
	ShippingMethod      string              `json:"shipping_method"`
	ShippingStatus      string              `json:"shipping_status"`
	TrackingNumber      string              `json:"tracking_number"`
	ProviderCarrierCode string              `json:"provider_carrier_code"`
	ProviderCarrierName string              `json:"provider_carrier_name"`
	SubtotalAmount      float64             `json:"subtotal_amount"`
	ShippingFee         float64             `json:"shipping_fee"`
	TaxAmount           float64             `json:"tax_amount"`
	DiscountAmount      float64             `json:"discount_amount"`
	TotalAmount         float64             `json:"total_amount"`
	Currency            string              `json:"currency"`
	CouponCode          string              `json:"coupon_code"`
	PointsUsed          int                 `json:"points_used"`
	PointsValue         float64             `json:"points_value"`
	ShippingAddress     orderdomain.Address `json:"shipping_address"`
	BillingAddress      orderdomain.Address `json:"billing_address"`
	CustomerNote        string              `json:"customer_note"`
	Items               []PublicOrderItem   `json:"items"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	PaidAt              *time.Time          `json:"paid_at"`
	ShippedAt           *time.Time          `json:"shipped_at"`
	CompletedAt         *time.Time          `json:"completed_at"`
	CancelledAt         *time.Time          `json:"cancelled_at"`
}

type PublicOrderItem struct {
	ProductID   uint    `json:"product_id"`
	VariantID   *uint   `json:"variant_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Attributes  string  `json:"attributes"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
	Discount    float64 `json:"discount"`
	TaxAmount   float64 `json:"tax_amount"`
	Total       float64 `json:"total"`
}

func publicOrderResponse(item orderdomain.Order) PublicOrderResponse {
	return PublicOrderResponse{
		OrderNumber:         item.OrderNumber,
		Status:              item.Status,
		PaymentMethod:       item.PaymentMethod,
		PaymentStatus:       item.PaymentStatus,
		ShippingMethod:      item.ShippingMethod,
		ShippingStatus:      item.ShippingStatus,
		TrackingNumber:      item.TrackingNumber,
		ProviderCarrierCode: item.ProviderCarrierCode,
		ProviderCarrierName: item.ProviderCarrierName,
		SubtotalAmount:      item.SubtotalAmount,
		ShippingFee:         item.ShippingFee,
		TaxAmount:           item.TaxAmount,
		DiscountAmount:      item.DiscountAmount,
		TotalAmount:         item.TotalAmount,
		Currency:            item.Currency,
		CouponCode:          item.CouponCode,
		PointsUsed:          item.PointsUsed,
		PointsValue:         item.PointsValue,
		ShippingAddress:     item.ShippingAddress,
		BillingAddress:      item.BillingAddress,
		CustomerNote:        item.CustomerNote,
		Items:               publicOrderItems(item.Items),
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
		PaidAt:              item.PaidAt,
		ShippedAt:           item.ShippedAt,
		CompletedAt:         item.CompletedAt,
		CancelledAt:         item.CancelledAt,
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
	for _, item := range items {
		result = append(result, PublicOrderItem{
			ProductID:   item.ProductID,
			VariantID:   item.VariantID,
			ProductName: item.ProductName,
			SKU:         item.SKU,
			Attributes:  item.Attributes,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Subtotal,
			Discount:    item.Discount,
			TaxAmount:   item.TaxAmount,
			Total:       item.Total,
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

func billingAddressFromRequest(shippingAddr orderdomain.Address, req AddressRequest) orderdomain.Address {
	if req.FirstName == "" {
		return shippingAddr
	}
	return addressFromRequest(req)
}
