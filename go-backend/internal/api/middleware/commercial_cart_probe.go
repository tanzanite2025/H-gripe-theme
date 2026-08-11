package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/visitorcookie"

	"github.com/gin-gonic/gin"
)

const (
	commercialCartWriteRequestThreshold   = 24
	commercialAnonymousCartWriteThreshold = 8
	commercialCartMaximumQuantity         = 25
	commercialCartTargetRequestThreshold  = 12
)

type commercialCartRequestInspection struct {
	quantityExceeded bool
	targets          []string
}

func inspectCommercialCartRequest(c *gin.Context) commercialCartRequestInspection {
	inspection := commercialCartRequestInspection{}
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return inspection
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return inspection
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	path := strings.Trim(strings.TrimSpace(c.Request.URL.Path), "/")
	if path == "api/v1/cart/sync" {
		var items []struct {
			ProductID uint  `json:"product_id"`
			VariantID *uint `json:"variant_id"`
			Quantity  int   `json:"quantity"`
		}
		if err := json.Unmarshal(body, &items); err != nil {
			return inspection
		}
		for _, item := range items {
			if item.Quantity > commercialCartMaximumQuantity {
				inspection.quantityExceeded = true
			}
			if target := commercialCartItemTarget(item.ProductID, item.VariantID); target != "" {
				inspection.targets = appendUniqueString(inspection.targets, target)
			}
		}
		return inspection
	}

	var item struct {
		ProductID uint  `json:"product_id"`
		VariantID *uint `json:"variant_id"`
		Quantity  int   `json:"quantity"`
	}
	if err := json.Unmarshal(body, &item); err != nil {
		return inspection
	}
	if item.Quantity > commercialCartMaximumQuantity {
		inspection.quantityExceeded = true
	}

	if pathPrefix := "api/v1/cart/items/"; strings.HasPrefix(path, pathPrefix) && item.ProductID == 0 {
		if productID, err := strconv.ParseUint(strings.TrimPrefix(path, pathPrefix), 10, 32); err == nil {
			item.ProductID = uint(productID)
		}
	}
	if target := commercialCartItemTarget(item.ProductID, item.VariantID); target != "" {
		inspection.targets = append(inspection.targets, target)
	}
	return inspection
}

func commercialCartItemTarget(productID uint, variantID *uint) string {
	if productID == 0 && variantID == nil {
		return ""
	}

	target := "product:" + strconv.FormatUint(uint64(productID), 10)
	if variantID != nil {
		target += "|variant:" + strconv.FormatUint(uint64(*variantID), 10)
	}
	return target
}

func commercialCartIdentities(c *gin.Context) []string {
	if c == nil {
		return nil
	}

	identities := make([]string, 0, 4)
	if ipIdentity := commercialRequestIdentity(c); ipIdentity != "" {
		identities = append(identities, commercialIdentityKey("ip", ipIdentity))
	}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok && id > 0 {
			identities = append(identities, commercialIdentityKey("user", strconv.FormatUint(uint64(id), 10)))
		}
	}
	if sessionID, err := c.Cookie("session_id"); err == nil && strings.TrimSpace(sessionID) != "" {
		identities = append(identities, commercialIdentityKey("session", sessionID))
	}
	if visitorCookie, err := c.Cookie(visitorcookie.CustomerServiceVisitorCookie); err == nil && strings.TrimSpace(visitorCookie) != "" {
		// This is a throttling hint, not an authentication factor. The IP
		// identity remains mandatory when the request comes from the public edge.
		identities = append(identities, commercialIdentityKey("visitor", visitorCookie))
	}

	return appendUniqueString(nil, identities...)
}

func commercialIdentityKey(kind, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(kind + "|" + value))
	return kind + ":" + hex.EncodeToString(sum[:])
}

func appendUniqueString(values []string, additions ...string) []string {
	seen := make(map[string]struct{}, len(values)+len(additions))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			seen[value] = struct{}{}
		}
	}
	for _, value := range additions {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}
