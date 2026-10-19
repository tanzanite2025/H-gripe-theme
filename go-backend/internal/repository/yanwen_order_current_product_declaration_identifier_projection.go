package repository

import (
	"fmt"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/product"
)

// Read current catalog identifiers in one query. The successful Yanwen request
// stores its own immutable snapshot; later catalog edits do not rewrite it.
func (r *OrderRepository) findCurrentProductDeclarationIdentifiersForYanwenOrder(items []order.OrderItem) (map[uint]product.Product, error) {
	identifiers := make(map[uint]product.Product, len(items))
	productIDs := make([]uint, 0, len(items))
	for _, item := range items {
		if item.ProductID != 0 {
			productIDs = append(productIDs, item.ProductID)
		}
	}
	if len(productIDs) == 0 {
		return identifiers, nil
	}
	var products []product.Product
	if err := r.db.Select("id", "merchant_product_identifier", "manufacturer_product_identifier", "standard_product_identifier").
		Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("读取燕文商品申报识别码失败: %w", err)
	}
	for _, record := range products {
		identifiers[record.ID] = record
	}
	return identifiers, nil
}
