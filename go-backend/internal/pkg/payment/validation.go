package payment

import (
	"fmt"
	"regexp"
	"strings"

	"commerce-platform/internal/domain/currency"
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
	req.Currency = currency.NormalizeCode(req.Currency)

	if !currency.IsValidCode(req.Currency) || !currency.IsCatalogCode(req.Currency) {
		return fmt.Errorf("unsupported currency code")
	}

	if req.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}

	if req.Customer == nil {
		return fmt.Errorf("customer information is required")
	}

	normalizedBIN, err := NormalizeCardBIN(req.CardBIN)
	if err != nil {
		return err
	}
	req.CardBIN = normalizedBIN

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

// NormalizeCardBIN accepts only a 6- or 8-digit issuer BIN. It never accepts
// a full card number, expiration date, security code, or any other PAN data.
func NormalizeCardBIN(value string) (string, error) {
	value = strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(value))
	if value == "" {
		return "", nil
	}
	if len(value) != 6 && len(value) != 8 {
		return "", fmt.Errorf("card BIN must contain exactly 6 or 8 digits")
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return "", fmt.Errorf("card BIN must contain exactly 6 or 8 digits")
		}
	}
	return value, nil
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
