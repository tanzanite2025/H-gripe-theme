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

type VerifiedWechatWebhook struct {
	ID          string
	EventType   string
	Transaction WechatWebhookTransaction
	Plaintext   string
}

func VerifyWechatWebhook(ctx context.Context, config *Config, headers http.Header, payload []byte) (VerifiedWechatWebhook, error) {
	if config == nil {
		return VerifiedWechatWebhook{}, fmt.Errorf("wechat config is required")
	}
	if strings.TrimSpace(config.WechatAPIv3Key) == "" {
		return VerifiedWechatWebhook{}, fmt.Errorf("wechat api_v3_key is not configured")
	}
	for _, key := range []string{"Wechatpay-Timestamp", "Wechatpay-Nonce", "Wechatpay-Signature", "Wechatpay-Serial"} {
		if strings.TrimSpace(headers.Get(key)) == "" {
			return VerifiedWechatWebhook{}, fmt.Errorf("missing wechat webhook header %s", key)
		}
	}

	verifier, err := newWechatPayVerifier(config)
	if err != nil {
		return VerifiedWechatWebhook{}, err
	}
	handler, err := notify.NewRSANotifyHandler(config.WechatAPIv3Key, verifier)
	if err != nil {
		return VerifiedWechatWebhook{}, fmt.Errorf("failed to create wechat notify handler: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://wechatpay.local/webhook", bytes.NewReader(payload))
	if err != nil {
		return VerifiedWechatWebhook{}, err
	}
	req.Header = headers.Clone()

	var transaction WechatWebhookTransaction
	notifyReq, err := handler.ParseNotifyRequest(ctx, req, &transaction)
	if err != nil {
		return VerifiedWechatWebhook{}, fmt.Errorf("wechat webhook verification failed: %w", err)
	}
	return VerifiedWechatWebhook{
		ID:          notifyReq.ID,
		EventType:   notifyReq.EventType,
		Transaction: transaction,
		Plaintext:   notifyReq.Resource.Plaintext,
	}, nil
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
