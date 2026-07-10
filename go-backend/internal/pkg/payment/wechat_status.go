package payment

// GetWechatPaymentStatus 将微信支付状态映射为统一状态
func GetWechatPaymentStatus(wechatStatus string) string {
	switch wechatStatus {
	case "SUCCESS":
		return "succeeded"
	case "NOTPAY":
		return "pending"
	case "CLOSED":
		return "closed"
	case "REVOKED":
		return "revoked"
	case "USERPAYING":
		return "processing"
	case "PAYERROR":
		return "failed"
	case "REFUND":
		return "refunded"
	default:
		return wechatStatus
	}
}

// getStringValue 安全地获取字符串指针的值
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
