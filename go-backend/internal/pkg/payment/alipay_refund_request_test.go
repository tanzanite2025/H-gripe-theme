package payment

import "testing"

func TestBuildAlipayRefundRequestUsesProviderTradeNo(t *testing.T) {
	req, err := buildAlipayRefundRequest("ALIPAY-TRADE-1", 12.34, "rf_1", RefundOptions{
		ProviderTransactionID: "ALIPAY-TRADE-1",
		MerchantOrderNumber:   "ORD-1",
		Reason:                "customer returned item",
	})
	if err != nil {
		t.Fatalf("buildAlipayRefundRequest() error = %v", err)
	}
	if req.TradeNo != "ALIPAY-TRADE-1" {
		t.Fatalf("TradeNo = %q, want ALIPAY-TRADE-1", req.TradeNo)
	}
	if req.OutTradeNo != "" {
		t.Fatalf("OutTradeNo = %q, want empty", req.OutTradeNo)
	}
	if req.RefundAmount != "12.34" {
		t.Fatalf("RefundAmount = %q, want 12.34", req.RefundAmount)
	}
	if req.OutRequestNo != "rf_1" {
		t.Fatalf("OutRequestNo = %q, want rf_1", req.OutRequestNo)
	}
	if req.RefundReason != "customer returned item" {
		t.Fatalf("RefundReason = %q", req.RefundReason)
	}
}

func TestBuildAlipayRefundRequestRequiresMerchantOrderNumber(t *testing.T) {
	_, err := buildAlipayRefundRequest("ALIPAY-TRADE-1", 12.34, "rf_1", RefundOptions{
		ProviderTransactionID: "ALIPAY-TRADE-1",
	})
	if err == nil {
		t.Fatalf("buildAlipayRefundRequest() expected error")
	}
}
