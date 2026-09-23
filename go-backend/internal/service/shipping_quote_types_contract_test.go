package service

import (
	"reflect"
	"testing"
)

func TestShippingQuoteDisplayAmountsAreDecimalStrings(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeOf(ShippingQuoteItem{}),
		reflect.TypeOf(ShippingQuote{}),
		reflect.TypeOf(ShippingQuotePlan{}),
		reflect.TypeOf(ShippingQuoteLeg{}),
	}
	for _, typ := range types {
		for index := 0; index < typ.NumField(); index++ {
			field := typ.Field(index)
			if field.Name == "UnitPrice" || field.Name == "Amount" || field.Name == "ShippingFee" || field.Name == "BaseFee" || field.Name == "FuelSurcharge" || field.Name == "RemoteSurcharge" {
				t.Fatalf("legacy floating-point shipping amount field remains: %s.%s", typ.Name(), field.Name)
			}
		}
	}

	for _, field := range []struct {
		typ  reflect.Type
		name string
	}{
		{reflect.TypeOf(ShippingQuoteItem{}), "UnitPriceDecimal"},
		{reflect.TypeOf(ShippingQuoteItem{}), "AmountDecimal"},
		{reflect.TypeOf(ShippingQuoteItem{}), "ShippingFeeDecimal"},
		{reflect.TypeOf(ShippingQuote{}), "ShippingFeeDecimal"},
		{reflect.TypeOf(ShippingQuotePlan{}), "ShippingFeeDecimal"},
		{reflect.TypeOf(ShippingQuoteLeg{}), "BaseFeeDecimal"},
		{reflect.TypeOf(ShippingQuoteLeg{}), "FuelSurchargeDecimal"},
		{reflect.TypeOf(ShippingQuoteLeg{}), "RemoteSurchargeDecimal"},
		{reflect.TypeOf(ShippingQuoteLeg{}), "ShippingFeeDecimal"},
	} {
		value, ok := field.typ.FieldByName(field.name)
		if !ok {
			t.Fatalf("decimal shipping amount field is missing: %s.%s", field.typ.Name(), field.name)
		}
		if value.Type.Kind() != reflect.String {
			t.Fatalf("shipping amount field must be a decimal string: %s.%s is %s", field.typ.Name(), field.name, value.Type)
		}
	}
}
