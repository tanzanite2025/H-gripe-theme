package ordernumber

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// The custom epoch keeps the snowflake timestamp compact without exposing it.
	snowflakeEpochMillis int64 = 1704067200000 // 2024-01-01T00:00:00Z
	nodeBits                   = 10
	sequenceBits               = 12
	maxNodeID                  = (1 << nodeBits) - 1
	maxSequence                = (1 << sequenceBits) - 1
	payloadBytes               = 10
	checksumBytes              = 3
)

var (
	ErrMissingSecret    = errors.New("order number secret is not configured")
	ErrInvalidNodeID    = errors.New("order number node id is out of range")
	ErrInvalidOrderNo   = errors.New("invalid order number")
	publicNumberPattern = regexp.MustCompile(`^TZ-[0-9]{4}-[A-Z2-7]{20}$`)
	sequentialPattern   = regexp.MustCompile(`^#?[0-9]+$`)
	base32Encoding      = base32.StdEncoding.WithPadding(base32.NoPadding)
)

// Generator creates opaque public order numbers from an internal snowflake ID.
// The snowflake value is never returned directly; HMAC-SHA-256 hides its
// timestamp, node, and sequence bits before the public token is formatted.
type Generator struct {
	key            []byte
	validationKeys [][]byte
	processNonce   [16]byte
	nodeID         uint16
	mu             sync.Mutex
	lastMS         int64
	sequence       uint16
}

func NewGenerator(secret string, nodeID uint16) (*Generator, error) {
	return NewGeneratorWithPreviousSecret(secret, "", nodeID)
}

// NewGeneratorWithPreviousSecret creates public numbers with secret and accepts
// checksums signed by either secret or previousSecret. previousSecret exists
// solely for a controlled one-key rotation window; it is never used to generate
// new order numbers.
func NewGeneratorWithPreviousSecret(secret, previousSecret string, nodeID uint16) (*Generator, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, ErrMissingSecret
	}
	if int(nodeID) > maxNodeID {
		return nil, ErrInvalidNodeID
	}

	key := deriveKey(secret)

	generator := &Generator{
		key:            key,
		validationKeys: [][]byte{key},
		nodeID:         nodeID,
	}
	previousSecret = strings.TrimSpace(previousSecret)
	if previousSecret != "" && previousSecret != secret {
		generator.validationKeys = append(generator.validationKeys, deriveKey(previousSecret))
	}
	if _, err := rand.Read(generator.processNonce[:]); err != nil {
		return nil, fmt.Errorf("generate order number process nonce: %w", err)
	}
	return generator, nil
}

func deriveKey(secret string) []byte {
	keyMAC := hmac.New(sha256.New, []byte(secret))
	_, _ = keyMAC.Write([]byte("commerce-platform/order-number/v1"))
	return keyMAC.Sum(nil)
}

func (g *Generator) Generate() (string, error) {
	return g.GenerateAt(time.Now().UTC())
}

func (g *Generator) GenerateAt(now time.Time) (string, error) {
	if g == nil || len(g.key) == 0 {
		return "", ErrMissingSecret
	}

	g.mu.Lock()
	snowflake, timestampMillis := g.nextSnowflake(now.UTC().UnixMilli())
	g.mu.Unlock()

	var raw [8]byte
	binary.BigEndian.PutUint64(raw[:], snowflake)

	payloadMAC := hmac.New(sha256.New, g.key)
	_, _ = payloadMAC.Write([]byte("payload|"))
	_, _ = payloadMAC.Write(g.processNonce[:])
	_, _ = payloadMAC.Write(raw[:])
	payload := base32Encoding.EncodeToString(payloadMAC.Sum(nil)[:payloadBytes])

	year := time.UnixMilli(timestampMillis).UTC().Year()
	checksum := checksumFor(g.key, year, payload)

	return fmt.Sprintf("TZ-%04d-%s%s", year, payload, checksum), nil
}

func (g *Generator) nextSnowflake(nowMillis int64) (uint64, int64) {
	if nowMillis < snowflakeEpochMillis {
		nowMillis = snowflakeEpochMillis
	}
	if nowMillis < g.lastMS {
		nowMillis = g.lastMS
	}
	if nowMillis == g.lastMS {
		if g.sequence >= maxSequence {
			nowMillis++
			g.sequence = 0
		} else {
			g.sequence++
		}
	} else {
		g.sequence = 0
	}
	g.lastMS = nowMillis

	timestamp := uint64(nowMillis-snowflakeEpochMillis) << (nodeBits + sequenceBits)
	node := uint64(g.nodeID) << sequenceBits
	return timestamp | node | uint64(g.sequence), nowMillis
}

func (g *Generator) Validate(value string) bool {
	value = strings.ToUpper(strings.TrimSpace(value))
	if !publicNumberPattern.MatchString(value) {
		return false
	}
	if g == nil || len(g.key) == 0 {
		return false
	}

	parts := strings.SplitN(value, "-", 3)
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	payloadAndChecksum := parts[2]
	payload := payloadAndChecksum[:len(payloadAndChecksum)-4]
	actualChecksum := payloadAndChecksum[len(payloadAndChecksum)-4:]

	for _, key := range g.validationKeys {
		if hmac.Equal([]byte(checksumFor(key, year, payload)), []byte(actualChecksum)) {
			return true
		}
	}
	return false
}

func checksumFor(key []byte, year int, payload string) string {
	checksumMAC := hmac.New(sha256.New, key)
	_, _ = checksumMAC.Write([]byte("checksum|"))
	_, _ = checksumMAC.Write([]byte(strconv.Itoa(year)))
	_, _ = checksumMAC.Write([]byte("|"))
	_, _ = checksumMAC.Write([]byte(payload))
	return base32Encoding.EncodeToString(checksumMAC.Sum(nil)[:checksumBytes])[:4]
}

func IsProtectedFormat(value string) bool {
	return publicNumberPattern.MatchString(strings.ToUpper(strings.TrimSpace(value)))
}

func IsSequentialCandidate(value string) bool {
	return sequentialPattern.MatchString(strings.TrimSpace(value))
}
