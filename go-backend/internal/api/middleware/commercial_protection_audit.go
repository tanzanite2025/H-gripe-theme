package middleware

import (
	"strconv"
	"time"

	appLogger "tanzanite/internal/pkg/logger"
	"tanzanite/internal/pkg/visitorcookie"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const commercialProtectionRuleVersion = "commercial-behavior-v1"

type commercialProtectionAuditInput struct {
	SeedID      string
	Outcome     string
	Action      string
	Reason      string
	Window      time.Duration
	ReleaseMode string
}

type commercialProtectionAuditContext struct {
	RuleVersion   string
	SeedID        string
	Outcome       string
	Action        string
	Reason        string
	ReleaseMode   string
	ExpiresAt     *time.Time
	Method        string
	Path          string
	IdentityKeys  []string
	UserAgentHash string
}

func logCommercialProtectionAction(c *gin.Context, input commercialProtectionAuditInput) {
	context := buildCommercialProtectionAuditContext(c, input, time.Now().UTC())
	fields := []zap.Field{
		zap.String("rule_version", context.RuleVersion),
		zap.String("seed_id", context.SeedID),
		zap.String("outcome", context.Outcome),
		zap.String("action", context.Action),
		zap.String("reason", context.Reason),
		zap.String("release_mode", context.ReleaseMode),
		zap.String("method", context.Method),
		zap.String("path", context.Path),
		zap.Strings("identity_keys", context.IdentityKeys),
		zap.String("user_agent_hash", context.UserAgentHash),
	}
	if context.ExpiresAt != nil {
		fields = append(fields, zap.Time("expires_at", *context.ExpiresAt))
	}
	appLogger.Warn("commercial intelligence protection action", fields...)
}

func buildCommercialProtectionAuditContext(c *gin.Context, input commercialProtectionAuditInput, now time.Time) commercialProtectionAuditContext {
	context := commercialProtectionAuditContext{
		RuleVersion: commercialProtectionRuleVersion,
		SeedID:      input.SeedID,
		Outcome:     input.Outcome,
		Action:      input.Action,
		Reason:      input.Reason,
		ReleaseMode: input.ReleaseMode,
	}
	if input.Window > 0 {
		expiresAt := now.Add(input.Window)
		context.ExpiresAt = &expiresAt
	}
	if c == nil || c.Request == nil {
		return context
	}

	context.Method = c.Request.Method
	context.Path = c.Request.URL.Path
	context.IdentityKeys = commercialProtectionIdentityKeys(c)
	if userAgent := c.Request.UserAgent(); userAgent != "" {
		context.UserAgentHash = commercialIdentityKey("user_agent", userAgent)
	}
	return context
}

func commercialProtectionIdentityKeys(c *gin.Context) []string {
	if c == nil {
		return nil
	}

	identities := make([]string, 0, 4)
	if ipIdentity := commercialRequestIdentity(c); ipIdentity != "" {
		identities = append(identities, commercialIdentityKey("ip", ipIdentity))
	}
	if userID, exists := c.Get("user_id"); exists {
		switch typed := userID.(type) {
		case uint:
			identities = append(identities, commercialIdentityKey("user", strconv.FormatUint(uint64(typed), 10)))
		case int:
			if typed > 0 {
				identities = append(identities, commercialIdentityKey("user", strconv.Itoa(typed)))
			}
		}
	}
	if sessionID, err := c.Cookie("session_id"); err == nil {
		identities = append(identities, commercialIdentityKey("session", sessionID))
	}
	if visitorCookie, err := c.Cookie(visitorcookie.CustomerServiceVisitorCookie); err == nil {
		identities = append(identities, commercialIdentityKey("visitor", visitorCookie))
	}
	return appendUniqueString(nil, identities...)
}
