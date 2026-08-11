package service

import (
	"fmt"

	"commerce-platform/internal/domain/product"
)

func (s *ProductService) enqueueMerchantProductUpsert(item *product.Product, reason string) error {
	if s == nil || s.merchantEvents == nil || item == nil {
		return nil
	}
	return s.merchantEvents.EnqueueProductUpsert(item.ID, reason)
}

func (s *ProductService) enqueueMerchantProductWithdraw(item *product.Product, reason string) error {
	if s == nil || s.merchantEvents == nil || item == nil {
		return nil
	}
	return s.merchantEvents.EnqueueProductWithdraw(item.ID, reason)
}

func (s *ProductService) enqueueMerchantProductChange(item *product.Product, reason string) error {
	if item == nil {
		return nil
	}
	if item.Status == "active" {
		return s.enqueueMerchantProductUpsert(item, reason)
	}
	return s.enqueueMerchantProductWithdraw(item, reason)
}

func merchantProductSourceChangeReason(input ProductUpdateInput) string {
	switch {
	case input.Status != nil:
		return "product_status_changed"
	case input.UpdateVariants:
		return "product_variants_changed"
	case input.UpdateMedia:
		return "product_media_changed"
	case input.Name != nil:
		return "product_name_changed"
	case input.Slug != nil:
		return "product_slug_changed"
	case input.Description != nil || input.ShortDesc != nil:
		return "product_description_changed"
	case input.UpdateCurrency:
		return "product_currency_changed"
	default:
		return "product_source_changed"
	}
}

func merchantProductUpdateAffectsChannel(input ProductUpdateInput) bool {
	return input.Status != nil ||
		input.Name != nil ||
		input.Slug != nil ||
		input.Description != nil ||
		input.ShortDesc != nil ||
		input.UpdateCurrency ||
		input.UpdateVariants ||
		input.UpdateMedia ||
		input.UpdateShippingTemplateID ||
		input.UpdateAfterSalesTemplateID ||
		input.UpdatePackagingTemplateID
}

func merchantProductCoreChanged(previous, next *product.Product) bool {
	if previous == nil || next == nil {
		return true
	}
	return previous.SKU != next.SKU ||
		previous.Name != next.Name ||
		previous.Slug != next.Slug ||
		previous.Description != next.Description ||
		previous.ShortDesc != next.ShortDesc ||
		previous.Currency != next.Currency ||
		previous.Price != next.Price ||
		!floatPointerEqual(previous.SalePrice, next.SalePrice) ||
		previous.Stock != next.Stock ||
		previous.Status != next.Status
}

func floatPointerEqual(left, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func requireMerchantEvent(err error, item *product.Product, reason string) error {
	if err == nil {
		return nil
	}
	if item == nil {
		return err
	}
	return fmt.Errorf("product %d saved but Merchant event was not queued: %w", item.ID, err)
}
