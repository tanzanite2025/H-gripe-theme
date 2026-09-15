package payment

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/smartwalle/alipay/v3"
)

type AlipayWebhookNotification struct {
	NotifyID     string
	NotifyType   string
	AppID        string
	OutTradeNo   string
	TradeNo      string
	OutRequestNo string
	OutBizNo     string
	TradeStatus  string
	RefundStatus string
	RefundAmount string
	RefundFee    string
	TotalAmount  string
	Currency     string
}

func VerifyAlipayWebhook(ctx context.Context, config *Config, payload []byte) (AlipayWebhookNotification, error) {
	if config == nil {
		return AlipayWebhookNotification{}, fmt.Errorf("alipay config is required")
	}
	if strings.TrimSpace(config.APIKey) == "" || strings.TrimSpace(config.SecretKey) == "" {
		return AlipayWebhookNotification{}, fmt.Errorf("alipay app_id and private_key are required for webhook verification")
	}
	if strings.TrimSpace(config.WebhookSecret) == "" {
		return AlipayWebhookNotification{}, fmt.Errorf("alipay public_key is not configured")
	}
	values, err := url.ParseQuery(string(payload))
	if err != nil {
		return AlipayWebhookNotification{}, fmt.Errorf("invalid alipay notification form payload: %w", err)
	}
	if strings.TrimSpace(values.Get("sign")) == "" {
		return AlipayWebhookNotification{}, fmt.Errorf("alipay notification signature is required")
	}

	client, err := alipay.New(config.APIKey, config.SecretKey, strings.EqualFold(config.Environment, "production"))
	if err != nil {
		return AlipayWebhookNotification{}, fmt.Errorf("failed to create alipay verifier client: %w", err)
	}
	if err := client.LoadAliPayPublicKey(config.WebhookSecret); err != nil {
		return AlipayWebhookNotification{}, fmt.Errorf("failed to load alipay public key: %w", err)
	}
	if err := client.VerifySign(ctx, values); err != nil {
		return AlipayWebhookNotification{}, fmt.Errorf("alipay webhook signature verification failed: %w", err)
	}

	notification := AlipayWebhookNotification{
		NotifyID:     strings.TrimSpace(values.Get("notify_id")),
		NotifyType:   strings.TrimSpace(values.Get("notify_type")),
		AppID:        strings.TrimSpace(values.Get("app_id")),
		OutTradeNo:   values.Get("out_trade_no"),
		TradeNo:      values.Get("trade_no"),
		OutRequestNo: values.Get("out_request_no"),
		OutBizNo:     values.Get("out_biz_no"),
		TradeStatus:  values.Get("trade_status"),
		RefundStatus: values.Get("refund_status"),
		RefundAmount: values.Get("refund_amount"),
		RefundFee:    values.Get("refund_fee"),
		TotalAmount:  values.Get("total_amount"),
		Currency:     values.Get("currency"),
	}
	if err := ValidateAlipayWebhookMerchantIdentity(config, notification); err != nil {
		return AlipayWebhookNotification{}, err
	}
	if notification.Currency == "" {
		notification.Currency = "CNY"
	}
	return notification, nil
}

// ValidateAlipayWebhookMerchantIdentity binds a verified notification to the
// Alipay application configured for this merchant.
func ValidateAlipayWebhookMerchantIdentity(config *Config, notification AlipayWebhookNotification) error {
	if config == nil {
		return fmt.Errorf("alipay config is required")
	}

	expectedAppID := strings.TrimSpace(config.APIKey)
	if expectedAppID == "" {
		return fmt.Errorf("alipay app_id is not configured for webhook verification")
	}
	if strings.TrimSpace(notification.AppID) == "" {
		return fmt.Errorf("alipay notification app_id is required")
	}
	if strings.TrimSpace(notification.AppID) != expectedAppID {
		return fmt.Errorf("alipay notification app_id does not match configured app_id")
	}
	return nil
}
