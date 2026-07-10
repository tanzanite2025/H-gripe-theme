package payment

import (
	"fmt"
	"regexp"
)

// validateConfig 验证支付网关配置
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Type == "" {
		return fmt.Errorf("gateway type is required")
	}

	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if config.SecretKey == "" && config.Type != GatewayStripe {
		return fmt.Errorf("secret key is required")
	}

	if config.Environment != "sandbox" && config.Environment != "production" {
		return fmt.Errorf("environment must be 'sandbox' or 'production'")
	}

	return nil
}

// ValidatePaymentRequest 验证支付请求
func ValidatePaymentRequest(req *PaymentRequest) error {
	if req == nil {
		return fmt.Errorf("payment request cannot be nil")
	}

	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	// 验证货币代码格式 (ISO 4217)
	currencyRegex := regexp.MustCompile(`^[A-Z]{3}$`)
	if !currencyRegex.MatchString(req.Currency) {
		return fmt.Errorf("invalid currency code: must be 3 uppercase letters")
	}

	if req.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}

	if req.Customer == nil {
		return fmt.Errorf("customer information is required")
	}

	if req.Customer.Email == "" {
		return fmt.Errorf("customer email is required")
	}

	// 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Customer.Email) {
		return fmt.Errorf("invalid customer email format")
	}

	return nil
}

// ValidateRefundAmount 验证退款金额
func ValidateRefundAmount(amount, originalAmount float64) error {
	if amount <= 0 {
		return fmt.Errorf("refund amount must be greater than 0")
	}

	if amount > originalAmount {
		return fmt.Errorf("refund amount cannot exceed original payment amount")
	}

	return nil
}
