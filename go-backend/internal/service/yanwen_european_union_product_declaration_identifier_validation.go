package service

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

func yanwenDestinationRequiresEuropeanUnionProductIdentifiers(country string) bool {
	switch strings.ToUpper(strings.TrimSpace(country)) {
	case "AT", "BE", "BG", "CY", "HR", "CZ", "DK", "EE", "FI", "FR", "DE", "GR", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PL", "PT", "RO", "SK", "SI", "ES", "SE":
		return true
	default:
		return false
	}
}

// Both the builder and the gateway enforce this rule, independently of channel
// tax settings. Apply it immediately, including FAT, to prevent invalid EU orders.
func normalizeAndValidateYanwenProductDeclarationIdentifiers(request *YanwenCreateOrderRequest) error {
	required := yanwenDestinationRequiresEuropeanUnionProductIdentifiers(request.ReceiverInfo.Country)
	products := append([]YanwenProductDeclare(nil), request.ParcelInfo.ProductList...)
	if required && len(products) == 0 {
		return errors.New("燕文欧盟订单缺少商品申报明细，无法校验 MPID 和 NSPID")
	}
	var problems []string
	for index := range products {
		item := &products[index]
		for _, field := range []struct {
			value    *string
			label    string
			required bool
		}{
			{&item.MerchantProductIdentifier, "商家产品识别码（MPID）", required},
			{&item.ManufacturerProductIdentifier, "制造商产品识别码（NSPID）", required},
			{&item.StandardProductIdentifier, "国际标准商品编码（SPID）", false},
		} {
			*field.value = strings.TrimSpace(*field.value)
			reference := fmt.Sprintf("订单 %s 商品第 %d 项（%s / SKU：%s）", request.OrderNumber, index+1, item.GoodsNameChinese, item.SKU)
			if field.required && *field.value == "" {
				problems = append(problems, reference+"缺少"+field.label)
			} else if utf8.RuneCountInString(*field.value) > 70 {
				problems = append(problems, reference+"的"+field.label+"超过 70 个字符")
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("燕文建单已阻止：%s。请在产品编辑的清关资料中补齐或修正后重试，本次未提交燕文", strings.Join(problems, "；"))
	}
	request.ParcelInfo.ProductList = products
	return nil
}
