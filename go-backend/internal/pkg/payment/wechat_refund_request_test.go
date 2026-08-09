package payment

import "testing"

func TestBuildWechatRefundRequestUsesProviderTransactionID(t *testing.T) {
	req, err := buildWechatRefundRequest("WX-TXN-1", 12.34, "rf_1", RefundOptions{
		ProviderTransactionID: "WX-TXN-1",
		MerchantOrderNumber:   "ORD-1",
		OriginalAmount:        99.99,
		Currency:              "CNY",
		Reason:                "customer returned item",
	})
	if err != nil {
		t.Fatalf("buildWechatRefundRequest() error = %v", err)
	}
	if req.TransactionId == nil || *req.TransactionId != "WX-TXN-1" {
		t.Fatalf("TransactionId = %#v, want WX-TXN-1", req.TransactionId)
	}
	if req.OutTradeNo != nil {
		t.Fatalf("OutTradeNo = %#v, want nil", req.OutTradeNo)
	}
	if req.OutRefundNo == nil || *req.OutRefundNo != "rf_1" {
		t.Fatalf("OutRefundNo = %#v, want rf_1", req.OutRefundNo)
	}
	if req.Reason == nil || *req.Reason != "customer returned item" {
		t.Fatalf("Reason = %#v", req.Reason)
	}
	if req.Amount == nil || req.Amount.Refund == nil || *req.Amount.Refund != 1234 {
		t.Fatalf("Refund amount = %#v, want 1234", req.Amount)
	}
	if req.Amount.Total == nil || *req.Amount.Total != 9999 {
		t.Fatalf("Total amount = %#v, want 9999", req.Amount)
	}
}

func TestBuildWechatRefundRequestRequiresOriginalAmount(t *testing.T) {
	_, err := buildWechatRefundRequest("WX-TXN-1", 12.34, "rf_1", RefundOptions{
		ProviderTransactionID: "WX-TXN-1",
		MerchantOrderNumber:   "ORD-1",
		Currency:              "CNY",
	})
	if err == nil {
		t.Fatalf("buildWechatRefundRequest() expected error")
	}
}
