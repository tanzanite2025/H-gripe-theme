package payment

import "testing"

func TestUnsafeGatewayWebhookVerificationFailsClosed(t *testing.T) {
	alipayGateway := &alipayGatewayImpl{config: &Config{WebhookSecret: "alipay_public_key"}}
	if ok, err := alipayGateway.VerifyWebhook([]byte(`{}`), "signature"); ok || err == nil {
		t.Fatalf("expected Alipay webhook verification to fail closed, ok=%v err=%v", ok, err)
	}

	wechatGateway := &wechatGatewayImpl{}
	if ok, err := wechatGateway.VerifyWebhook([]byte(`{}`), "signature"); ok || err == nil {
		t.Fatalf("expected WeChat webhook verification to fail closed, ok=%v err=%v", ok, err)
	}
}
