package tracking

import (
	"fmt"
	"regexp"
	"strings"
)

// validateTrackingNumber 验证物流单号
func validateTrackingNumber(trackingNumber string) error {
	if trackingNumber == "" {
		return fmt.Errorf("tracking number cannot be empty")
	}

	// 移除空格
	trackingNumber = strings.TrimSpace(trackingNumber)

	// 检查长度 (一般物流单号在 8-40 个字符之间)
	if len(trackingNumber) < 8 || len(trackingNumber) > 40 {
		return fmt.Errorf("invalid tracking number length: must be between 8 and 40 characters")
	}

	// 检查是否只包含字母数字和连字符
	validChars := regexp.MustCompile(`^[A-Za-z0-9\-]+$`)
	if !validChars.MatchString(trackingNumber) {
		return fmt.Errorf("invalid tracking number format: only alphanumeric characters and hyphens allowed")
	}

	return nil
}

// validateBatchTrackingRequest 验证批量追踪请求
func validateBatchTrackingRequest(trackings []TrackingRequest) error {
	if len(trackings) == 0 {
		return fmt.Errorf("no tracking requests provided")
	}

	if len(trackings) > 100 {
		return fmt.Errorf("too many tracking requests: maximum 100 allowed")
	}

	for i, req := range trackings {
		if err := validateTrackingNumber(req.TrackingNumber); err != nil {
			return fmt.Errorf("invalid tracking request at index %d: %w", i, err)
		}
	}

	return nil
}
