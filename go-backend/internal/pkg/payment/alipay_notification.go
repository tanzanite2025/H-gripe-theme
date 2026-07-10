package payment

import (
	"fmt"

	"github.com/smartwalle/alipay/v3"
)

// VerifyAlipayNotification 验证支付宝异步通知（推荐方法）
func VerifyAlipayNotification(client *alipay.Client, values map[string]string) (bool, error) {
	return false, fmt.Errorf("alipay notification verification requires SDK API update")
}
