package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/security"
	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	commercialCrawlerAutoBlockDuration              = 6 * time.Hour
	commercialCrawlerAutoBlockSourceReferencePrefix = "user-agent:"
	commercialCrawlerAuditResource                  = "global_ip_block_rule"
)

// CommercialCrawlerAutoBlockPolicy describes the durable action taken after a
// known commercial crawler signature is observed.
func CommercialCrawlerAutoBlockPolicy() gin.H {
	return gin.H{
		"enabled":          true,
		"source":           security.IPBlockRuleSourceCommercialBot,
		"duration_seconds": int(commercialCrawlerAutoBlockDuration.Seconds()),
		"duration":         commercialCrawlerAutoBlockDuration.String(),
		"scope":            "public client IP as an exact /32 or /128 rule",
		"audit":            "transactional",
	}
}

func persistCommercialCrawlerIPBlock(
	c *gin.Context,
	blockService *service.GlobalIPBlockService,
	detector CommercialCrawlerRule,
) {
	if c == nil || c.Request == nil || blockService == nil {
		return
	}

	ipAddress := service.NormalizeIP(c.ClientIP())
	if ipAddress == "" {
		recordCommercialCrawlerAutoBlockOutcome(
			c,
			detector,
			"skipped",
			"client_ip_unavailable",
			"persist_ip_block",
			0,
		)
		return
	}
	if isPrivateNetworkIP(ipAddress) {
		recordCommercialCrawlerAutoBlockOutcome(
			c,
			detector,
			"skipped_private_ip",
			"private_client_ip",
			"persist_ip_block",
			0,
		)
		return
	}

	cidr, err := service.NormalizeIPOrCIDR(ipAddress)
	if err != nil {
		recordCommercialCrawlerAutoBlockFailure(c, detector, ipAddress, err)
		return
	}

	sourceReference := commercialCrawlerAutoBlockSourceReference(detector.Provider)
	now := time.Now().UTC()
	if activeRules, lookupErr := blockService.FindActiveRulesBySourceReference(
		c.Request.Context(),
		security.IPBlockRuleSourceCommercialBot,
		sourceReference,
		now,
	); lookupErr == nil {
		for _, rule := range activeRules {
			if rule.CIDR == cidr {
				recordCommercialCrawlerAutoBlockOutcome(
					c,
					detector,
					"already_active",
					"existing_global_ip_block",
					"global_ip_block_rule",
					0,
				)
				return
			}
		}
	} else if !errors.Is(lookupErr, service.ErrIPBlockCacheUnavailable) {
		appLogger.Warn(
			"commercial crawler auto-block lookup failed; attempting durable write",
			zap.String("provider", detector.Provider),
			zap.String("cidr", cidr),
			zap.String("source_reference", sourceReference),
			zap.Error(lookupErr),
		)
	}

	expiresAt := now.Add(commercialCrawlerAutoBlockDuration)
	before, after, blockErr := blockService.BlockWithPreviousAndAudit(
		c.Request.Context(),
		service.IPBlockRuleInput{
			CIDR:            cidr,
			Source:          security.IPBlockRuleSourceCommercialBot,
			SourceReference: sourceReference,
			Reason:          "known commercial crawler user-agent signature matched",
			ExpiresAt:       &expiresAt,
		},
		newCommercialCrawlerIPBlockAuditFactory(c, detector, ipAddress),
	)
	if blockErr != nil {
		if errors.Is(blockErr, service.ErrIPBlockCacheRefresh) && after.ID > 0 {
			recordCommercialCrawlerAutoBlockOutcome(
				c,
				detector,
				"auto_blocked_cache_refresh_pending",
				"global_ip_block_rule",
				"durable_rule_saved_cache_refresh_pending",
				commercialCrawlerAutoBlockDuration,
			)
			return
		}
		recordCommercialCrawlerAutoBlockFailure(c, detector, cidr, blockErr)
		return
	}

	outcome := "auto_blocked"
	if before != nil {
		outcome = "auto_block_refreshed"
	}
	recordCommercialCrawlerAutoBlockOutcome(
		c,
		detector,
		outcome,
		"global_ip_block_rule",
		"known_user_agent_match",
		commercialCrawlerAutoBlockDuration,
	)
}

func commercialCrawlerAutoBlockSourceReference(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	provider = strings.NewReplacer(" ", "-", "_", "-").Replace(provider)
	if provider == "" {
		provider = "unknown"
	}
	return commercialCrawlerAutoBlockSourceReferencePrefix + provider
}

func newCommercialCrawlerIPBlockAuditFactory(
	c *gin.Context,
	detector CommercialCrawlerRule,
	ipAddress string,
) service.IPBlockAuditLogFactory {
	return func(before *service.IPBlockRuleSnapshot, after service.IPBlockRuleSnapshot) (*audit.AuditLog, error) {
		createdAt := time.Now().UTC()
		action := "create"
		var oldValue interface{}
		if before != nil {
			action = "update"
			oldValue = commercialCrawlerIPBlockAuditDetails(detector, before)
		}
		newValue := commercialCrawlerIPBlockAuditDetails(detector, &after)

		changes, err := marshalCommercialCrawlerAuditValue(newValue)
		if err != nil {
			return nil, fmt.Errorf("marshal commercial crawler audit changes: %w", err)
		}
		oldValueJSON, err := marshalCommercialCrawlerAuditValue(oldValue)
		if err != nil {
			return nil, fmt.Errorf("marshal commercial crawler audit old value: %w", err)
		}
		newValueJSON, err := marshalCommercialCrawlerAuditValue(newValue)
		if err != nil {
			return nil, fmt.Errorf("marshal commercial crawler audit new value: %w", err)
		}

		log := &audit.AuditLog{
			Username:   "automatic-commercial-crawler",
			Action:     action,
			Resource:   commercialCrawlerAuditResource,
			ResourceID: after.ID,
			IPAddress:  ipAddress,
			Changes:    changes,
			OldValue:   oldValueJSON,
			NewValue:   newValueJSON,
			Status:     "success",
			Duration:   int(time.Since(createdAt).Milliseconds()),
			CreatedAt:  createdAt,
		}
		if c != nil && c.Request != nil {
			log.Method = c.Request.Method
			log.UserAgent = "commercial-crawler/" + strings.TrimSpace(detector.Provider)
			if c.Request.URL != nil {
				log.Path = c.Request.URL.Path
			}
		}
		return log, nil
	}
}

func commercialCrawlerIPBlockAuditDetails(
	detector CommercialCrawlerRule,
	rule *service.IPBlockRuleSnapshot,
) map[string]interface{} {
	details := map[string]interface{}{
		"detector":         "known_user_agent",
		"provider":         strings.TrimSpace(detector.Provider),
		"signature":        strings.TrimSpace(detector.UserAgent),
		"source":           security.IPBlockRuleSourceCommercialBot,
		"source_reference": commercialCrawlerAutoBlockSourceReference(detector.Provider),
	}
	if rule == nil {
		return details
	}
	details["rule_id"] = rule.ID
	details["cidr"] = rule.CIDR
	details["status"] = rule.Status
	details["expires"] = rule.ExpiresAt != nil
	if rule.ExpiresAt != nil {
		details["expires_at"] = rule.ExpiresAt.UTC()
	}
	return details
}

func marshalCommercialCrawlerAuditValue(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func recordCommercialCrawlerAutoBlockOutcome(
	c *gin.Context,
	detector CommercialCrawlerRule,
	outcome string,
	action string,
	reason string,
	window time.Duration,
) {
	metrics.CommercialIntelligenceActions.WithLabelValues(
		"known-commercial-crawlers",
		outcome,
	).Inc()
	logCommercialProtectionAction(c, commercialProtectionAuditInput{
		SeedID:      "known-commercial-crawlers",
		Provider:    detector.Provider,
		Outcome:     outcome,
		Action:      action,
		Reason:      reason,
		Window:      window,
		ReleaseMode: "global_ip_block_rule",
	})
}

func recordCommercialCrawlerAutoBlockFailure(
	c *gin.Context,
	detector CommercialCrawlerRule,
	target string,
	err error,
) {
	metrics.CommercialIntelligenceActions.WithLabelValues(
		"known-commercial-crawlers",
		"auto_block_failed",
	).Inc()
	appLogger.Error(
		"commercial crawler automatic IP block failed",
		zap.String("provider", detector.Provider),
		zap.String("target", target),
		zap.Error(err),
	)
	logCommercialProtectionAction(c, commercialProtectionAuditInput{
		SeedID:      "known-commercial-crawlers",
		Provider:    detector.Provider,
		Outcome:     "auto_block_failed",
		Action:      "global_ip_block_rule",
		Reason:      "durable_rule_or_audit_write_failed",
		ReleaseMode: "request_scope",
	})
}
