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
	productRecord := seedHighValuePolicyProduct(t, db, "SKU-HIGH-VALUE-DISCOUNT", "High value product", 80000, 5)
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
	)
	require.NoError(t, err)
	require.NotNil(t, belowThreshold)
	assert.Equal(t, int64(70000), belowThreshold.TotalAmountMinor)
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
	)
	require.NoError(t, err)
	require.NotNil(t, atThreshold)
	assert.Equal(t, int64(75000), atThreshold.TotalAmountMinor)
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
	firstProduct := seedHighValuePolicyProduct(t, db, "SKU-HIGH-VALUE-FIRST", "First item", 40000, 5)
	secondProduct := seedHighValuePolicyProduct(t, db, "SKU-HIGH-VALUE-SECOND", "Second item", 35000, 5)

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
	)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.Equal(t, int64(75000), createdOrder.TotalAmountMinor)
	assert.True(t, createdOrder.SignatureRequired)

	var savedOrder order.Order
	require.NoError(t, db.Preload("Items").First(&savedOrder, createdOrder.ID).Error)
	require.Len(t, savedOrder.Items, 2)
	assert.Equal(t, int64(40000), savedOrder.Items[0].TotalMinor)
	assert.Equal(t, int64(35000), savedOrder.Items[1].TotalMinor)
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
		75000,
		5,
		wheelCategory.ID,
	)
	ordinaryProduct := seedHighValuePolicyProductWithCategory(
		t,
		db,
		"SKU-HIGH-VALUE-ORDINARY",
		"Standard Accessory",
		75000,
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
		)

		require.NoError(t, err)
		require.NotNil(t, createdOrder)
		assert.Equal(t, int64(75000), createdOrder.TotalAmountMinor)
		assert.True(t, createdOrder.SignatureRequired, "product %q should use the same order-total policy", productRecord.Name)
	}
}

func seedHighValuePolicyProduct(
	t *testing.T,
	db *gorm.DB,
	sku string,
	name string,
	priceMinor int64,
	stock int,
) productdomain.Product {
	return seedHighValuePolicyProductWithCategory(t, db, sku, name, priceMinor, stock, 0)
}

func seedHighValuePolicyProductWithCategory(
	t *testing.T,
	db *gorm.DB,
	sku string,
	name string,
	priceMinor int64,
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
		PriceMinor:         priceMinor,
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
		PriceMinor:   priceMinor,
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
