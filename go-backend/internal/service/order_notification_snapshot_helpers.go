package service

import (
	"net/url"
	"strings"

	"commerce-platform/internal/domain/order"
)

func orderCustomerName(orderRecord *order.Order) string {
	if orderRecord == nil {
		return ""
	}
	return strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(orderRecord.ShippingAddress.FirstName),
		strings.TrimSpace(orderRecord.ShippingAddress.LastName),
	}, " "))
}

func resolveOrderTrackingURL(template, trackingNumber string) string {
	template = strings.TrimSpace(template)
	trackingNumber = strings.TrimSpace(trackingNumber)
	if template == "" || trackingNumber == "" {
		return ""
	}

	resolved := strings.ReplaceAll(template, "{tracking_number}", url.PathEscape(trackingNumber))
	parsed, err := url.Parse(resolved)
	if err != nil || parsed.Host == "" {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return parsed.String()
}
