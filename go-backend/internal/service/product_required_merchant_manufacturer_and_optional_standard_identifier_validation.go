package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"commerce-platform/internal/domain/product"
)

// Product identifiers are carrier-independent catalog data. Never manufacture
// an identifier from a SKU or copy the merchant identifier into the factory code.
func normalizeAndValidateRequiredProductIdentifiers(record *product.Product) error {
	fields := []struct {
		value    *string
		label    string
		required bool
	}{
		{&record.MerchantProductIdentifier, "商家产品识别码（MPID）", true},
		{&record.ManufacturerProductIdentifier, "制造商产品识别码（NSPID）", true},
		{&record.StandardProductIdentifier, "国际标准商品编码（SPID）", false},
	}
	for _, field := range fields {
		*field.value = strings.TrimSpace(*field.value)
		if field.required && *field.value == "" {
			return fmt.Errorf("%w: 请填写%s", ErrProductCustomsInfoInvalid, field.label)
		}
		if utf8.RuneCountInString(*field.value) > 70 {
			return fmt.Errorf("%w: %s不能超过 70 个字符", ErrProductCustomsInfoInvalid, field.label)
		}
	}
	return nil
}
