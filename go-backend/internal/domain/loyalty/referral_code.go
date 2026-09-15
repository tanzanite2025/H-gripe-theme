package loyalty

import (
	"crypto/rand"
	"errors"
	"strings"
)

const (
	ReferralCodeLength   = 8
	referralCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

var ErrInvalidReferralCode = errors.New("invalid referral code")

func NormalizeReferralCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func ValidateReferralCode(value string) error {
	value = NormalizeReferralCode(value)
	if len(value) < 6 || len(value) > 16 {
		return ErrInvalidReferralCode
	}
	for _, character := range value {
		if !strings.ContainsRune(referralCodeAlphabet, character) {
			return ErrInvalidReferralCode
		}
	}
	return nil
}

func GenerateReferralCode() (string, error) {
	buffer := make([]byte, ReferralCodeLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	for index := range buffer {
		buffer[index] = referralCodeAlphabet[int(buffer[index])%len(referralCodeAlphabet)]
	}
	return string(buffer), nil
}
