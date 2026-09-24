package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// verifyHMACSHA256 验证 HMAC SHA256 签名
func verifyHMACSHA256(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
