package service

import (
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/shipping"
)

func buildYanwenCreateOrderRequest(
	orderRecord *order.Order,
	productCode string,
	warehouseCode string,
	orderSource string,
	hasBattery bool,
	countryCode string,
) (YanwenCreateOrderRequest, string, int, int, error) {
	return buildYanwenCreateOrderRequestWithCustomsDeclaration(
		orderRecord,
		productCode,
		warehouseCode,
		orderSource,
		hasBattery,
		countryCode,
		YanwenCustomsDeclarationInput{},
	)
}

func buildYanwenCreateOrderRequestWithCustomsDeclaration(
	orderRecord *order.Order,
	productCode string,
	warehouseCode string,
	orderSource string,
	hasBattery bool,
	countryCode string,
	customsDeclaration YanwenCustomsDeclarationInput,
) (YanwenCreateOrderRequest, string, int, int, error) {
	return buildYanwenCreateOrderRequestFromFacts(
		newYanwenOrderFactsFromOrder(orderRecord),
		productCode,
		warehouseCode,
		orderSource,
		hasBattery,
		countryCode,
		customsDeclaration,
	)
}

func newYanwenOrderFactsFromOrder(orderRecord *order.Order) *shipping.YanwenOrderFacts {
	if orderRecord == nil {
		return nil
	}
	items := make([]shipping.YanwenOrderFactItem, 0, len(orderRecord.Items))
	for _, item := range orderRecord.Items {
		items = append(items, shipping.YanwenOrderFactItem{
			ProductName:            item.ProductName,
			SKU:                    item.SKU,
			Quantity:               item.Quantity,
			WeightGrams:            item.WeightGrams,
			HSCode:                 item.HSCode,
			CustomsDescription:     item.CustomsDescription,
			DeclaredValueMinor:     item.DeclaredValueMinor,
			DeclaredValueConfirmed: item.DeclaredValueConfirmed,
		})
	}
	return &shipping.YanwenOrderFacts{
		ID:            orderRecord.ID,
		OrderNumber:   orderRecord.OrderNumber,
		Status:        orderRecord.Status,
		PaymentStatus: orderRecord.PaymentStatus,
		Currency:      orderRecord.Currency,
		ShippingAddress: shipping.YanwenOrderFactAddress{
			FirstName:  orderRecord.ShippingAddress.FirstName,
			LastName:   orderRecord.ShippingAddress.LastName,
			Company:    orderRecord.ShippingAddress.Company,
			Address1:   orderRecord.ShippingAddress.Address1,
			Address2:   orderRecord.ShippingAddress.Address2,
			City:       orderRecord.ShippingAddress.City,
			State:      orderRecord.ShippingAddress.State,
			PostalCode: orderRecord.ShippingAddress.PostalCode,
			Country:    orderRecord.ShippingAddress.Country,
			Phone:      orderRecord.ShippingAddress.Phone,
			Email:      orderRecord.ShippingAddress.Email,
		},
		Items:  items,
		PaidAt: orderRecord.PaidAt,
	}
}
