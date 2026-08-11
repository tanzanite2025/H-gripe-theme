package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/requestsign"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	requestTimestampHeader = "X-Request-Timestamp"
	requestNonceHeader     = "X-Request-Nonce"
	requestSignatureHeader = "X-Request-Signature"
)

func RequestSignature(cfg config.RequestSigningConfig, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		timestamp := c.GetHeader(requestTimestampHeader)
		nonce := c.GetHeader(requestNonceHeader)
		signature := c.GetHeader(requestSignatureHeader)
		hasAnyHeader := timestamp != "" || nonce != "" || signature != ""
		required := requestSignatureRequired(c.Request.URL.Path, cfg.RequiredPaths)
		if !hasAnyHeader && !required {
			c.Next()
			return
		}
		if timestamp == "" || nonce == "" || signature == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request signature required"})
			c.Abort()
			return
		}

		const maxSignedBodyBytes = 8 << 20
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxSignedBodyBytes+1))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read request body"})
			c.Abort()
			return
		}
		if len(body) > maxSignedBodyBytes {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "signed request body is too large"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		target := c.Request.URL.EscapedPath()
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}
		if err := requestsign.Verify(
			c.Request.Method,
			target,
			timestamp,
			nonce,
			signature,
			body,
			cfg.Key,
			time.Now(),
			time.Duration(cfg.MaxSkewSeconds)*time.Second,
		); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid request signature"})
			c.Abort()
			return
		}

		if redisClient == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "request signature service unavailable"})
			c.Abort()
			return
		}
		replayKey := "commerce_platform:request-signature:nonce:" + requestsign.BodyDigest([]byte(nonce))
		accepted, err := redisClient.SetNX(c.Request.Context(), replayKey, "1", time.Duration(cfg.MaxSkewSeconds+15)*time.Second).Result()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "request signature service unavailable"})
			c.Abort()
			return
		}
		if !accepted {
			c.JSON(http.StatusConflict, gin.H{"error": "request signature replay detected"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func requestSignatureRequired(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix != "" && strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
