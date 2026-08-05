package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const prefix = "v1:"

func EncryptString(value, masterKey string) (string, error) {
	if strings.TrimSpace(masterKey) == "" {
		return "", errors.New("secretbox master key is required")
	}
	if value == "" {
		return "", errors.New("secretbox value is required")
	}

	block, err := aes.NewCipher(deriveKey(masterKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(value), nil)
	payload := append(nonce, ciphertext...)
	return prefix + base64.RawStdEncoding.EncodeToString(payload), nil
}

func DecryptString(value, masterKey string) (string, error) {
	if strings.TrimSpace(masterKey) == "" {
		return "", errors.New("secretbox master key is required")
	}
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, prefix) {
		return "", errors.New("unsupported secretbox format")
	}

	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil {
		return "", fmt.Errorf("decode secretbox payload: %w", err)
	}

	block, err := aes.NewCipher(deriveKey(masterKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) <= gcm.NonceSize() {
		return "", errors.New("invalid secretbox payload")
	}

	plaintext, err := gcm.Open(nil, payload[:gcm.NonceSize()], payload[gcm.NonceSize():], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secretbox payload: %w", err)
	}
	return string(plaintext), nil
}

func deriveKey(masterKey string) []byte {
	hash := sha256.Sum256([]byte(masterKey))
	return hash[:]
}
