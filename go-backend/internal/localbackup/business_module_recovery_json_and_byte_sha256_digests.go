package localbackup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Keep the existing json.Marshal representation and raw-byte SHA-256 algorithms. Do not change serialization, field ordering or nil handling during extraction.
func businessModuleRecoveryCanonicalJSONSHA256(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func businessModuleRecoveryBytesSHA256(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
