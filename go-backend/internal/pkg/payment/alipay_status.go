package payment

// GetAlipayPaymentStatus 将支付宝状态映射为统一状态
func GetAlipayPaymentStatus(alipayStatus string) string {
	switch alipayStatus {
	case "WAIT_BUYER_PAY":
		return "pending"
	case "TRADE_SUCCESS":
		return "succeeded"
	case "TRADE_FINISHED":
		return "completed"
	case "TRADE_CLOSED":
		return "closed"
	default:
		return alipayStatus
	}
}
