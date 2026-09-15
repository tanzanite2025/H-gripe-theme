package payment

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
)

type WechatWebhookTransaction struct {
	AppID         string `json:"appid"`
	MchID         string `json:"mchid"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	Attach        string `json:"attach"`
	Amount        struct {
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

// WechatRefundNotification is the decrypted resource used by REFUND.*
// notifications. It is intentionally separate from the collection transaction
// schema: refund events do not contain trade_state and use refund-specific
// amount fields.
type WechatRefundNotification struct {
	AppID         string `json:"appid"`
	MchID         string `json:"mchid"`
	TransactionID string `json:"transaction_id"`
	OutTradeNo    string `json:"out_trade_no"`
	RefundID      string `json:"refund_id"`
	OutRefundNo   string `json:"out_refund_no"`
	RefundStatus  string `json:"refund_status"`
	Amount        struct {
		Total       int64  `json:"total"`
		Refund      int64  `json:"refund"`
		PayerTotal  int64  `json:"payer_total"`
		PayerRefund int64  `json:"payer_refund"`
		Currency    string `json:"currency"`
	} `json:"amount"`
}

type VerifiedWechatWebhook struct {
	ID          string
	EventType   string
	Transaction WechatWebhookTransaction
	Plaintext   string
}

func VerifyWechatWebhook(ctx context.Context, config *Config, headers http.Header, payload []byte) (VerifiedWechatWebhook, error) {
	var transaction WechatWebhookTransaction
	notifyReq, plaintext, err := verifyWechatNotification(ctx, config, headers, payload, &transaction)
	if err != nil {
		return VerifiedWechatWebhook{}, err
	}
	if err := ValidateWechatWebhookMerchantIdentity(config, transaction); err != nil {
		return VerifiedWechatWebhook{}, err
	}
	return VerifiedWechatWebhook{
		ID:          notifyReq.ID,
		EventType:   notifyReq.EventType,
		Transaction: transaction,
		Plaintext:   plaintext,
	}, nil
}

// VerifiedWechatRefundWebhook contains a verified and decrypted refund
// notification. Refund notifications are verified with the same RSA-V3/AES
// envelope, but are decoded into WechatRefundNotification rather than the
// payment transaction model.
type VerifiedWechatRefundWebhook struct {
	ID        string
	EventType string
	Refund    WechatRefundNotification
	Plaintext string
}

func VerifyWechatRefundWebhook(ctx context.Context, config *Config, headers http.Header, payload []byte) (VerifiedWechatRefundWebhook, error) {
	var refund WechatRefundNotification
	notifyReq, plaintext, err := verifyWechatNotification(ctx, config, headers, payload, &refund)
	if err != nil {
		return VerifiedWechatRefundWebhook{}, err
	}
	if err := ValidateWechatRefundWebhookMerchantIdentity(config, refund); err != nil {
		return VerifiedWechatRefundWebhook{}, err
	}
	return VerifiedWechatRefundWebhook{
		ID:        notifyReq.ID,
		EventType: notifyReq.EventType,
		Refund:    refund,
		Plaintext: plaintext,
	}, nil
}

func ValidateWechatRefundWebhookMerchantIdentity(config *Config, refund WechatRefundNotification) error {
	if config == nil {
		return fmt.Errorf("wechat config is required")
	}
	// Domestic refund resources are documented to carry mchid but not appid.
	// The envelope signature plus merchant id binds this resource to the
	// configured account; requiring appid here would reject every valid refund.
	if strings.TrimSpace(refund.MchID) == "" {
		return fmt.Errorf("wechat refund notification mchid is required")
	}
	if strings.TrimSpace(config.APIKey) != strings.TrimSpace(refund.MchID) {
		return fmt.Errorf("wechat refund notification mchid does not match configured mch_id")
	}
	return nil
}

func verifyWechatNotification(ctx context.Context, config *Config, headers http.Header, payload []byte, content interface{}) (*notify.Request, string, error) {
	if config == nil {
		return nil, "", fmt.Errorf("wechat config is required")
	}
	if strings.TrimSpace(config.WechatAPIv3Key) == "" {
		return nil, "", fmt.Errorf("wechat api_v3_key is not configured")
	}
	for _, key := range []string{"Wechatpay-Timestamp", "Wechatpay-Nonce", "Wechatpay-Signature", "Wechatpay-Serial"} {
		if strings.TrimSpace(headers.Get(key)) == "" {
			return nil, "", fmt.Errorf("missing wechat webhook header %s", key)
		}
	}

	verifier, err := newWechatPayVerifier(config)
	if err != nil {
		return nil, "", err
	}
	handler, err := notify.NewRSANotifyHandler(config.WechatAPIv3Key, verifier)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create wechat notify handler: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://wechatpay.local/webhook", bytes.NewReader(payload))
	if err != nil {
		return nil, "", err
	}
	req.Header = headers.Clone()

	notifyReq, err := handler.ParseNotifyRequest(ctx, req, content)
	if err != nil {
		return nil, "", fmt.Errorf("wechat webhook verification failed: %w", err)
	}
	if notifyReq == nil || notifyReq.Resource == nil {
		return nil, "", fmt.Errorf("wechat webhook resource is missing")
	}
	return notifyReq, notifyReq.Resource.Plaintext, nil
}

// ValidateWechatWebhookMerchantIdentity binds a verified transaction to the
// WeChat app and merchant configured for this merchant.
func ValidateWechatWebhookMerchantIdentity(config *Config, transaction WechatWebhookTransaction) error {
	if config == nil {
		return fmt.Errorf("wechat config is required")
	}

	expectedAppID := strings.TrimSpace(config.WechatAppID)
	if expectedAppID == "" {
		return fmt.Errorf("wechat app_id is not configured for webhook verification")
	}
	expectedMchID := strings.TrimSpace(config.APIKey)
	if expectedMchID == "" {
		return fmt.Errorf("wechat mch_id is not configured for webhook verification")
	}
	if strings.TrimSpace(transaction.AppID) == "" {
		return fmt.Errorf("wechat notification appid is required")
	}
	if strings.TrimSpace(transaction.MchID) == "" {
		return fmt.Errorf("wechat notification mchid is required")
	}
	if strings.TrimSpace(transaction.AppID) != expectedAppID {
		return fmt.Errorf("wechat notification appid does not match configured app_id")
	}
	if strings.TrimSpace(transaction.MchID) != expectedMchID {
		return fmt.Errorf("wechat notification mchid does not match configured mch_id")
	}
	return nil
}

func newWechatPayVerifier(config *Config) (auth.Verifier, error) {
	if strings.TrimSpace(config.WechatPayPlatformPublicKey) != "" {
		publicKey, err := parseRSAPublicKey(config.WechatPayPlatformPublicKey)
		if err != nil {
			return nil, fmt.Errorf("invalid wechat platform public key: %w", err)
		}
		if strings.TrimSpace(config.WechatPayPlatformPublicKeyID) == "" {
			return nil, fmt.Errorf("wechat platform_public_key_id is required when platform_public_key is configured")
		}
		return verifiers.NewSHA256WithRSAPubkeyVerifier(config.WechatPayPlatformPublicKeyID, *publicKey), nil
	}
	if strings.TrimSpace(config.WechatPayPlatformCertificate) == "" {
		return nil, fmt.Errorf("wechat platform_certificate or platform_public_key is required for webhook verification")
	}
	cert, err := parseWechatPayCertificate(config.WechatPayPlatformCertificate)
	if err != nil {
		return nil, err
	}
	return verifiers.NewSHA256WithRSAVerifier(core.NewCertificateMapWithList([]*x509.Certificate{cert})), nil
}

func parseWechatPayCertificate(value string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, fmt.Errorf("invalid wechat platform certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid wechat platform certificate: %w", err)
	}
	return cert, nil
}

func parseRSAPublicKey(value string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM public key")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}
	return publicKey, nil
}
