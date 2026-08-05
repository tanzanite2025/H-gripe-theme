package visitor

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"gorm.io/datatypes"
)

const (
	RiskLevelNormal     = "normal"
	RiskLevelWatch      = "watch"
	RiskLevelSuspicious = "suspicious"
	RiskLevelBlock      = "block"

	RiskDecisionScopeIPHash   = "ip_hash"
	RiskDecisionScopeIPUAHash = "ip_ua_hash"

	RiskDecisionActionIgnore         = "ignore"
	RiskDecisionActionWatch          = "watch"
	RiskDecisionActionTemporaryBlock = "temporary_block"
	RiskDecisionActionBlockCandidate = "block_candidate"
)

type RiskDailyFact struct {
	ID                    uint           `gorm:"primarykey" json:"id"`
	Day                   time.Time      `gorm:"type:date;not null;uniqueIndex:uk_visitor_risk_daily_fact" json:"day"`
	IPHash                string         `gorm:"size:64;not null;uniqueIndex:uk_visitor_risk_daily_fact;index" json:"ip_hash"`
	UserAgentHash         string         `gorm:"size:64;not null;default:'';uniqueIndex:uk_visitor_risk_daily_fact" json:"user_agent_hash"`
	CountryCode           string         `gorm:"size:8;index" json:"country_code,omitempty"`
	FirstSeenAt           time.Time      `gorm:"not null" json:"first_seen_at"`
	LastSeenAt            time.Time      `gorm:"not null" json:"last_seen_at"`
	RequestCount          int            `gorm:"not null;default:0" json:"request_count"`
	UniquePathCount       int            `gorm:"not null;default:0" json:"unique_path_count"`
	UniqueAnonymousCount  int            `gorm:"not null;default:0" json:"unique_anonymous_count"`
	UniqueSessionCount    int            `gorm:"not null;default:0" json:"unique_session_count"`
	InvalidRequestCount   int            `gorm:"not null;default:0" json:"invalid_request_count"`
	AuthFailureCount      int            `gorm:"not null;default:0" json:"auth_failure_count"`
	CheckoutFailureCount  int            `gorm:"not null;default:0" json:"checkout_failure_count"`
	BotLikeUserAgentCount int            `gorm:"not null;default:0" json:"bot_like_user_agent_count"`
	NoCookieRequestCount  int            `gorm:"not null;default:0" json:"no_cookie_request_count"`
	MeaningfulActionCount int            `gorm:"not null;default:0" json:"meaningful_action_count"`
	RiskScore             int            `gorm:"not null;default:0;index" json:"risk_score"`
	RiskLevel             string         `gorm:"size:16;not null;default:'normal';index" json:"risk_level"`
	SamplePaths           datatypes.JSON `gorm:"not null" json:"sample_paths"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func (RiskDailyFact) TableName() string {
	return "visitor_risk_daily_facts"
}

type RiskDecision struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	Scope     string     `gorm:"size:24;not null;index" json:"scope"`
	ValueHash string     `gorm:"size:64;not null;index" json:"value_hash"`
	Action    string     `gorm:"size:32;not null;index" json:"action"`
	Reason    string     `gorm:"size:500;not null" json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedBy *uint      `gorm:"index" json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (RiskDecision) TableName() string {
	return "visitor_risk_decisions"
}

func RiskDecisionIPUAValueHash(ipHash, userAgentHash string) string {
	ipHash = strings.TrimSpace(ipHash)
	userAgentHash = strings.TrimSpace(userAgentHash)
	if ipHash == "" || userAgentHash == "" {
		return ""
	}
	sum := sha256.Sum256([]byte("ip_ua:" + ipHash + ":" + userAgentHash))
	return hex.EncodeToString(sum[:])
}
