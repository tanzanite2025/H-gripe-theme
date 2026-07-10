package order

import orderdomain "tanzanite/internal/domain/order"

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Items           []OrderItemRequest `json:"items" binding:"required,min=1"`
	ShippingAddress AddressRequest     `json:"shipping_address" binding:"required"`
	BillingAddress  AddressRequest     `json:"billing_address"`
	PaymentMethod   string             `json:"payment_method" binding:"required"`
	ShippingMethod  string             `json:"shipping_method" binding:"required"`
	CouponCode      string             `json:"coupon_code"`
	PointsToUse     int                `json:"points_to_use"`
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
