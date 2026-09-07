package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/order"
	productdomain "commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrderServiceHighValueUsesFinalOrderTotalAfterDiscounts(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedHighValuePolicyProduct(t, db, "SKU-HIGH-VALUE-DISCOUNT", "High value product", 800, 5)
	seedCoupon(t, db, "HIGH-VALUE-100", "fixed", 100, 1)
	seedCoupon(t, db, "HIGH-VALUE-50", "fixed", 50, 1)

	belowThreshold, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"HIGH-VALUE-100",
		0,
	)
	require.NoError(t, err)
	require.NotNil(t, belowThreshold)
	assert.InDelta(t, 700, belowThreshold.TotalAmount, 0.001)
	assert.False(t, belowThreshold.SignatureRequired)

	atThreshold, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"HIGH-VALUE-50",
		0,
	)
	require.NoError(t, err)
	require.NotNil(t, atThreshold)
	assert.InDelta(t, 750, atThreshold.TotalAmount, 0.001)
	assert.True(t, atThreshold.SignatureRequired)

	var savedBelowThreshold order.Order
	require.NoError(t, db.First(&savedBelowThreshold, belowThreshold.ID).Error)
	assert.False(t, savedBelowThreshold.SignatureRequired)

	var savedAtThreshold order.Order
	require.NoError(t, db.First(&savedAtThreshold, atThreshold.ID).Error)
	assert.True(t, savedAtThreshold.SignatureRequired)
}

func TestOrderServiceHighValueIsBasedOnOrderTotalAcrossMultipleItems(t *testing.T) {
	db, orderService := newTestOrderService(t)
	firstProduct := seedHighValuePolicyProduct(t, db, "SKU-HIGH-VALUE-FIRST", "First item", 400, 5)
	secondProduct := seedHighValuePolicyProduct(t, db, "SKU-HIGH-VALUE-SECOND", "Second item", 350, 5)

	createdOrder, err := orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{
			{ProductID: firstProduct.ID, Quantity: 1},
			{ProductID: secondProduct.ID, Quantity: 1},
		},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.InDelta(t, 750, createdOrder.TotalAmount, 0.001)
	assert.True(t, createdOrder.SignatureRequired)

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.Len(t, savedOrder.Items, 2)
	assert.InDelta(t, 400, savedOrder.Items[0].Total, 0.001)
	assert.InDelta(t, 350, savedOrder.Items[1].Total, 0.001)
	assert.True(t, savedOrder.SignatureRequired)
}

func TestOrderServiceHighValueIgnoresProductNameAndCategory(t *testing.T) {
	db, orderService := newTestOrderService(t)
	require.NoError(t, db.AutoMigrate(&productdomain.ProductCategory{}))

	wheelCategory := productdomain.ProductCategory{
		Name:      "Wheels",
		Slug:      "high-value-policy-wheels",
		IsEnabled: true,
	}
	ordinaryCategory := productdomain.ProductCategory{
		Name:      "Accessories",
		Slug:      "high-value-policy-accessories",
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&wheelCategory).Error)
	require.NoError(t, db.Create(&ordinaryCategory).Error)

	wheelProduct := seedHighValuePolicyProductWithCategory(
		t,
		db,
		"SKU-HIGH-VALUE-WHEEL",
		"Custom Wheelset Configuration",
		750,
		5,
		wheelCategory.ID,
	)
	ordinaryProduct := seedHighValuePolicyProductWithCategory(
		t,
		db,
		"SKU-HIGH-VALUE-ORDINARY",
		"Standard Accessory",
		750,
		5,
		ordinaryCategory.ID,
	)

	for _, productRecord := range []productdomain.Product{wheelProduct, ordinaryProduct} {
		createdOrder, err := orderService.CreateOrder(
			context.Background(),
			42,
			[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
			testAddress(),
			testAddress(),
			"card",
			"standard",
			"",
			0,
		)

		require.NoError(t, err)
		require.NotNil(t, createdOrder)
		assert.InDelta(t, 750, createdOrder.TotalAmount, 0.001)
		assert.True(t, createdOrder.SignatureRequired, "product %q should use the same order-total policy", productRecord.Name)
	}
}

func seedHighValuePolicyProduct(
	t *testing.T,
	db *gorm.DB,
	sku string,
	name string,
	price float64,
	stock int,
) productdomain.Product {
	return seedHighValuePolicyProductWithCategory(t, db, sku, name, price, stock, 0)
}

func seedHighValuePolicyProductWithCategory(
	t *testing.T,
	db *gorm.DB,
	sku string,
	name string,
	price float64,
	stock int,
	categoryID uint,
) productdomain.Product {
	t.Helper()

	shippingTemplateID := seedOrderTestShippingTemplateID(t, db)
	record := productdomain.Product{
		ProductCategoryID:  &categoryID,
		ShippingTemplateID: shippingTemplateID,
		SKU:                sku,
		Name:               name,
		Slug:               slugForHighValuePolicyTest(sku),
		Currency:           "USD",
		Price:              price,
		Stock:              stock,
		Status:             "active",
		Locale:             "en",
	}
	if categoryID == 0 {
		record.ProductCategoryID = nil
	}
	require.NoError(t, db.Create(&record).Error)
	require.NoError(t, db.Create(&productdomain.ProductVariant{
		ProductID:    record.ID,
		SKU:          sku,
		Title:        "Default",
		OptionValues: "{}",
		Currency:     "USD",
		Price:        price,
		Stock:        stock,
		Weight:       9000,
		IsDefault:    true,
		IsActive:     true,
	}).Error)
	return record
}

func slugForHighValuePolicyTest(sku string) string {
	switch sku {
	case "SKU-HIGH-VALUE-DISCOUNT":
		return "high-value-policy-discount"
	case "SKU-HIGH-VALUE-FIRST":
		return "high-value-policy-first"
	case "SKU-HIGH-VALUE-SECOND":
		return "high-value-policy-second"
	case "SKU-HIGH-VALUE-WHEEL":
		return "high-value-policy-wheel"
	case "SKU-HIGH-VALUE-ORDINARY":
		return "high-value-policy-ordinary"
	default:
		return "high-value-policy-product"
	}
}
