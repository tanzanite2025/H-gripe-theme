package payment

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/plutov/paypal/v4"
)

type fakePayPalWebhookVerifier struct {
	status  string
	err     error
	seen    *http.Request
	seenCtx context.Context
	seenID  string
}

func (f *fakePayPalWebhookVerifier) VerifyWebhookSignature(ctx context.Context, httpReq *http.Request, webhookID string) (*paypal.VerifyWebhookResponse, error) {
	f.seenCtx = ctx
	f.seen = httpReq
	f.seenID = webhookID
	if f.err != nil {
		return nil, f.err
	}
	return &paypal.VerifyWebhookResponse{VerificationStatus: f.status}, nil

}

func TestVerifyPayPalWebhookUsesOfficialVerifierAndParsesEvent(t *testing.T) {
	payload := []byte(`{"id":"evt_paypal_1","event_type":"CHECKOUT.ORDER.COMPLETED","resource":{"id":"order_1"}}`)
	headers := validPayPalWebhookHeaders()
	verifier := &fakePayPalWebhookVerifier{status: "SUCCESS"}

	event, err := VerifyPayPalWebhook(context.Background(), &Config{
		Type:          GatewayPayPal,
		APIKey:        "client-id",
		SecretKey:     "secret",
		WebhookSecret: "webhook-id",
	}, headers, payload, verifier)
	if err != nil {
		t.Fatalf("VerifyPayPalWebhook() error = %v", err)
	}
	if event.ID != "evt_paypal_1" || event.EventType != "CHECKOUT.ORDER.COMPLETED" {
		t.Fatalf("unexpected PayPal event: %#v", event)
	}
	if verifier.seenID != "webhook-id" {
		t.Fatalf("expected webhook id to be passed to verifier, got %q", verifier.seenID)
	}
	if verifier.seen == nil || verifier.seen.Header.Get("PAYPAL-TRANSMISSION-ID") != headers.Get("PAYPAL-TRANSMISSION-ID") {
		t.Fatalf("expected PayPal transmission headers to be preserved")
	}

}

func TestVerifyPayPalWebhookFailsClosedOnMissingTransmissionHeader(t *testing.T) {
	headers := validPayPalWebhookHeaders()
	headers.Del("PAYPAL-TRANSMISSION-SIG")

	_, err := VerifyPayPalWebhook(context.Background(), &Config{WebhookSecret: "webhook-id"}, headers, []byte(`{}`), &fakePayPalWebhookVerifier{status: "SUCCESS"})
	if err == nil || !strings.Contains(err.Error(), "PAYPAL-TRANSMISSION-SIG") {
		t.Fatalf("expected missing transmission signature error, got %v", err)
	}

}

func TestVerifyPayPalWebhookFailsClosedWhenOfficialVerifierRejects(t *testing.T) {
	_, err := VerifyPayPalWebhook(context.Background(), &Config{WebhookSecret: "webhook-id"}, validPayPalWebhookHeaders(), []byte(`{"id":"evt_1","event_type":"CHECKOUT.ORDER.COMPLETED"}`), &fakePayPalWebhookVerifier{status: "FAILURE"})
	if err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Fatalf("expected PayPal verifier rejection, got %v", err)
	}

	_, err = VerifyPayPalWebhook(context.Background(), &Config{WebhookSecret: "webhook-id"}, validPayPalWebhookHeaders(), []byte(`{"id":"evt_1","event_type":"CHECKOUT.ORDER.COMPLETED"}`), &fakePayPalWebhookVerifier{err: errors.New("paypal unavailable")})
	if err == nil || !strings.Contains(err.Error(), "signature verification failed") {
		t.Fatalf("expected PayPal verifier error, got %v", err)
	}

}

func TestVerifyPayPalWebhookMarksTransientVerifierTimeoutAsRetryable(t *testing.T) {
	_, err := VerifyPayPalWebhook(context.Background(), &Config{WebhookSecret: "webhook-id"}, validPayPalWebhookHeaders(), []byte(`{"id":"evt_1","event_type":"CHECKOUT.ORDER.COMPLETED"}`), &fakePayPalWebhookVerifier{err: context.DeadlineExceeded})
	if !errors.Is(err, ErrPayPalWebhookVerificationUnavailable) {
		t.Fatalf("expected transient verification sentinel, got %v", err)
	}
}

func TestVerifyPayPalWebhookUsesFiveSecondTimeout(t *testing.T) {
	verifier := &fakePayPalWebhookVerifier{status: "SUCCESS"}

	_, err := VerifyPayPalWebhook(context.Background(), &Config{
		WebhookSecret: "webhook-id",
	}, validPayPalWebhookHeaders(), []byte(`{"id":"evt_1","event_type":"CHECKOUT.ORDER.COMPLETED"}`), verifier)
	if err != nil {
		t.Fatalf("VerifyPayPalWebhook() error = %v", err)
	}
	if verifier.seenCtx == nil {
		t.Fatalf("expected webhook verifier to receive a context")
	}
	deadline, ok := verifier.seenCtx.Deadline()
	if !ok {
		t.Fatalf("expected webhook verifier context deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > 5*time.Second {
		t.Fatalf("unexpected webhook verification timeout remaining: %v", remaining)
	}
}

func TestNewPayPalVerificationClientUsesPooledHTTPClient(t *testing.T) {
	client, err := newPayPalVerificationClient(&Config{
		APIKey:        "client-id",
		SecretKey:     "secret",
		Environment:   "production",
		WebhookSecret: "webhook-id",
	})
	if err != nil {
		t.Fatalf("newPayPalVerificationClient() error = %v", err)
	}
	if client.Client == nil {
		t.Fatalf("expected configured HTTP client")
	}
	if client.Client.Timeout != paypalWebhookVerificationTimeout {
		t.Fatalf("unexpected http client timeout: %v", client.Client.Timeout)
	}
	transport, ok := client.Client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected http transport, got %T", client.Client.Transport)
	}
	if transport.MaxIdleConns != 50 || transport.MaxIdleConnsPerHost != 50 {
		t.Fatalf("unexpected transport idle conns config: %+v", transport)
	}
	if transport.IdleConnTimeout != 90*time.Second {
		t.Fatalf("unexpected transport idle conn timeout: %v", transport.IdleConnTimeout)
	}
}

func TestVerifyAlipayWebhookFailsClosedWithoutSignatureOrPublicKey(t *testing.T) {
	_, err := VerifyAlipayWebhook(context.Background(), &Config{
		Type:          GatewayAlipay,
		APIKey:        "app-id",
		SecretKey:     "private-key",
		WebhookSecret: "public-key",
	}, []byte("out_trade_no=TZN1001&trade_status=TRADE_SUCCESS"))
	if err == nil || !strings.Contains(err.Error(), "signature is required") {
		t.Fatalf("expected missing signature error, got %v", err)
	}

	_, err = VerifyAlipayWebhook(context.Background(), &Config{
		Type:      GatewayAlipay,
		APIKey:    "app-id",
		SecretKey: "private-key",
	}, []byte("sign=fake"))
	if err == nil || !strings.Contains(err.Error(), "public_key") {
		t.Fatalf("expected missing public key error, got %v", err)
	}

}

func TestValidateAlipayWebhookMerchantIdentity(t *testing.T) {
	config := &Config{APIKey: "app-id"}

	if err := ValidateAlipayWebhookMerchantIdentity(config, AlipayWebhookNotification{AppID: "app-id"}); err != nil {
		t.Fatalf("expected matching Alipay app_id to pass: %v", err)
	}
	if err := ValidateAlipayWebhookMerchantIdentity(config, AlipayWebhookNotification{}); err == nil || !strings.Contains(err.Error(), "notification app_id") {
		t.Fatalf("expected missing Alipay notification app_id error, got %v", err)
	}
	if err := ValidateAlipayWebhookMerchantIdentity(config, AlipayWebhookNotification{AppID: "attacker-app-id"}); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected cross-merchant Alipay app_id error, got %v", err)
	}
	if err := ValidateAlipayWebhookMerchantIdentity(&Config{}, AlipayWebhookNotification{AppID: "app-id"}); err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected missing configured Alipay app_id error, got %v", err)
	}
}

func TestVerifyWechatWebhookFailsClosedWithoutRequiredHeadersOrVerifierMaterial(t *testing.T) {
	_, err := VerifyWechatWebhook(context.Background(), &Config{Type: GatewayWechat}, http.Header{}, []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "api_v3_key") {
		t.Fatalf("expected missing api v3 key error, got %v", err)
	}

	_, err = VerifyWechatWebhook(context.Background(), &Config{
		Type:           GatewayWechat,
		WechatAPIv3Key: "0123456789abcdef0123456789abcdef",
	}, http.Header{}, []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "Wechatpay-Timestamp") {
		t.Fatalf("expected missing WeChat timestamp header error, got %v", err)
	}

	_, err = VerifyWechatWebhook(context.Background(), &Config{
		Type:           GatewayWechat,
		WechatAPIv3Key: "0123456789abcdef0123456789abcdef",
	}, validWechatWebhookHeaders(), []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "platform_certificate or platform_public_key") {
		t.Fatalf("expected missing WeChat platform verifier error, got %v", err)
	}

}

func TestWechatPlatformPublicKeyTakesPrecedenceOverCertificate(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	encodedPublicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal RSA public key: %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encodedPublicKey})

	config := &Config{
		WechatPayPlatformPublicKey:   string(publicKeyPEM),
		WechatPayPlatformPublicKeyID: "PUB_KEY_ID_123",
		WechatPayPlatformCertificate: "invalid legacy certificate PEM",
	}
	if _, err := newWechatPayVerifier(config); err != nil {
		t.Fatalf("expected configured platform public key to take precedence over certificate: %v", err)
	}

	config.WechatPayPlatformPublicKeyID = ""
	if _, err := newWechatPayVerifier(config); err == nil || !strings.Contains(err.Error(), "platform_public_key_id is required") {
		t.Fatalf("expected missing public key ID to fail closed instead of falling back to certificate, got %v", err)
	}
}

func TestValidateWechatWebhookMerchantIdentity(t *testing.T) {
	config := &Config{APIKey: "mch-id", WechatAppID: "app-id"}
	transaction := WechatWebhookTransaction{AppID: "app-id", MchID: "mch-id"}

	if err := ValidateWechatWebhookMerchantIdentity(config, transaction); err != nil {
		t.Fatalf("expected matching WeChat merchant identity to pass: %v", err)
	}

	if err := ValidateWechatWebhookMerchantIdentity(config, WechatWebhookTransaction{MchID: "mch-id"}); err == nil || !strings.Contains(err.Error(), "notification appid") {
		t.Fatalf("expected missing WeChat notification appid error, got %v", err)
	}
	if err := ValidateWechatWebhookMerchantIdentity(config, WechatWebhookTransaction{AppID: "app-id"}); err == nil || !strings.Contains(err.Error(), "notification mchid") {
		t.Fatalf("expected missing WeChat notification mchid error, got %v", err)
	}

	transaction.AppID = "attacker-app-id"
	if err := ValidateWechatWebhookMerchantIdentity(config, transaction); err == nil || !strings.Contains(err.Error(), "appid does not match") {
		t.Fatalf("expected cross-merchant WeChat appid error, got %v", err)
	}
	transaction = WechatWebhookTransaction{AppID: "app-id", MchID: "attacker-mch-id"}
	if err := ValidateWechatWebhookMerchantIdentity(config, transaction); err == nil || !strings.Contains(err.Error(), "mchid does not match") {
		t.Fatalf("expected cross-merchant WeChat mchid error, got %v", err)
	}
	if err := ValidateWechatWebhookMerchantIdentity(&Config{APIKey: "mch-id"}, transaction); err == nil || !strings.Contains(err.Error(), "app_id is not configured") {
		t.Fatalf("expected missing configured WeChat app_id error, got %v", err)
	}
	if err := ValidateWechatWebhookMerchantIdentity(&Config{WechatAppID: "app-id"}, transaction); err == nil || !strings.Contains(err.Error(), "mch_id is not configured") {
		t.Fatalf("expected missing configured WeChat mch_id error, got %v", err)
	}
}

func TestValidateWechatRefundWebhookMerchantIdentityAllowsMissingAppID(t *testing.T) {
	config := &Config{APIKey: "mch-id", WechatAppID: "app-id"}
	refund := WechatRefundNotification{MchID: "mch-id"}

	if err := ValidateWechatRefundWebhookMerchantIdentity(config, refund); err != nil {
		t.Fatalf("expected refund notification without appid to pass: %v", err)
	}
	if err := ValidateWechatRefundWebhookMerchantIdentity(config, WechatRefundNotification{AppID: "app-id"}); err == nil || !strings.Contains(err.Error(), "notification mchid") {
		t.Fatalf("expected missing refund notification mchid error, got %v", err)
	}
	if err := ValidateWechatRefundWebhookMerchantIdentity(config, WechatRefundNotification{MchID: "attacker-mch-id"}); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected cross-merchant refund mchid error, got %v", err)
	}
}

func validPayPalWebhookHeaders() http.Header {
	headers := http.Header{}
	headers.Set("PAYPAL-AUTH-ALGO", "SHA256withRSA")
	headers.Set("PAYPAL-CERT-URL", "https://api-m.paypal.com/certs/CERT-123")
	headers.Set("PAYPAL-TRANSMISSION-ID", "transmission-id")
	headers.Set("PAYPAL-TRANSMISSION-SIG", "signature")
	headers.Set("PAYPAL-TRANSMISSION-TIME", "2026-07-30T12:00:00Z")
	return headers
}

func validWechatWebhookHeaders() http.Header {
	headers := http.Header{}
	headers.Set("Wechatpay-Timestamp", "1760000000")
	headers.Set("Wechatpay-Nonce", "nonce")
	headers.Set("Wechatpay-Signature", "signature")
	headers.Set("Wechatpay-Serial", "serial")
	return headers
}
