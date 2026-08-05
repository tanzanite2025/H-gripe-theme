package payment

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/smartwalle/alipay/v3"
)

type AlipayWebhookNotification struct {
	AppID       string
	OutTradeNo  string
	TradeNo     string
	TradeStatus string
	TotalAmount string
	Currency    string
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
		AppID:       values.Get("app_id"),
		OutTradeNo:  values.Get("out_trade_no"),
		TradeNo:     values.Get("trade_no"),
		TradeStatus: values.Get("trade_status"),
		TotalAmount: values.Get("total_amount"),
		Currency:    values.Get("currency"),
	}
	if notification.Currency == "" {
		notification.Currency = "CNY"
	}
	return notification, nil
}
