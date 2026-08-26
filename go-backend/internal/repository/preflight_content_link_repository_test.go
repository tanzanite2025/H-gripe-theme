package repository

import (
	"encoding/json"
	"testing"
	"time"

	preflightdomain "commerce-platform/internal/domain/preflight"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPreflightContentLinkVerificationKeepsRuleIdentityEvidence(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&preflightdomain.ContentLinkRun{},
		&preflightdomain.ContentLinkIssue{},
		&preflightdomain.ContentLinkIssueEvent{},
	))

	now := time.Date(2026, time.August, 23, 14, 0, 0, 0, time.UTC)
	run := preflightdomain.ContentLinkRun{
		TargetURL: "https://example.com/blog/post",
		Status:    preflightdomain.ContentLinkRunStatusSuccess,
		CheckedAt: now,
	}
	require.NoError(t, db.Create(&run).Error)

	issue := preflightdomain.ContentLinkIssue{
		RunID:           run.ID,
		TargetURL:       run.TargetURL,
		RuleID:          preflightdomain.ContentLinkRuleID,
		ProviderAuditID: preflightdomain.ContentLinkProviderAuditID,
		LinkURL:         "https://example.com/guides",
		LinkText:        "Learn more",
		Selector:        "main article a",
		SourceType:      "blog_post",
		SourceKey:       "post-1",
		SourceField:     "content",
		IssueKey:        "issue-1",
		Severity:        "medium",
		State:           preflightdomain.ContentLinkIssueStateOpen,
		FixStatus:       preflightdomain.ContentLinkFixStatusPending,
		FirstDetectedAt: now.Add(-time.Minute),
		LastDetectedAt:  now.Add(-time.Minute),
	}
	require.NoError(t, db.Create(&issue).Error)

	require.NoError(t, NewPreflightContentLinkRepository(db).RecordDetections(&run, nil))

	var verified preflightdomain.ContentLinkIssue
	require.NoError(t, db.First(&verified, issue.ID).Error)
	require.Equal(t, preflightdomain.ContentLinkIssueStateVerified, verified.State)

	var evidence map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(verified.LatestEvidence), &evidence))
	require.Equal(t, preflightdomain.ContentLinkRuleID, evidence["rule_id"])
	require.Equal(t, preflightdomain.ContentLinkProviderAuditID, evidence["provider_audit_id"])
	require.Equal(t, issue.LinkURL, evidence["link_url"])
	require.Equal(t, issue.LinkText, evidence["link_text"])
	require.Equal(t, issue.Selector, evidence["selector"])
}

func TestPreflightContentLinkReadNormalizesHistoricalRuleIdentity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&preflightdomain.ContentLinkRun{},
		&preflightdomain.ContentLinkIssue{},
		&preflightdomain.ContentLinkIssueEvent{},
	))

	now := time.Date(2026, time.August, 23, 14, 0, 0, 0, time.UTC)
	issue := preflightdomain.ContentLinkIssue{
		RunID:           1,
		TargetURL:       "https://example.com/blog/post",
		RuleID:          "link-text",
		ProviderAuditID: "link_descriptive_text",
		LinkURL:         "https://example.com/guides",
		LinkText:        "Learn more",
		IssueKey:        "legacy-content-link",
		Severity:        "medium",
		State:           preflightdomain.ContentLinkIssueStateOpen,
		FixStatus:       preflightdomain.ContentLinkFixStatusPending,
		LatestEvidence:  `{"rule_id":"link-text","provider_audit_id":"link_descriptive_text","link_text":"Learn more"}`,
		FirstDetectedAt: now,
		LastDetectedAt:  now,
	}
	require.NoError(t, db.Create(&issue).Error)

	normalized, err := NewPreflightContentLinkRepository(db).FindIssueByID(issue.ID)
	require.NoError(t, err)
	require.Equal(t, preflightdomain.ContentLinkRuleID, normalized.RuleID)
	require.Equal(t, preflightdomain.ContentLinkProviderAuditID, normalized.ProviderAuditID)

	var evidence map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(normalized.LatestEvidence), &evidence))
	require.Equal(t, preflightdomain.ContentLinkRuleID, evidence["rule_id"])
	require.Equal(t, preflightdomain.ContentLinkProviderAuditID, evidence["provider_audit_id"])
	require.Equal(t, issue.LinkText, evidence["link_text"])
}

func TestPreflightContentLinkDetectionEventCarriesRuleIdentity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&preflightdomain.ContentLinkRun{},
		&preflightdomain.ContentLinkIssue{},
		&preflightdomain.ContentLinkIssueEvent{},
	))

	now := time.Date(2026, time.August, 23, 14, 0, 0, 0, time.UTC)
	run := preflightdomain.ContentLinkRun{
		TargetURL: "https://example.com/",
		Status:    preflightdomain.ContentLinkRunStatusSuccess,
		CheckedAt: now,
	}
	require.NoError(t, db.Create(&run).Error)

	detection := preflightdomain.ContentLinkDetection{
		RunID:           run.ID,
		TargetURL:       run.TargetURL,
		RuleID:          preflightdomain.ContentLinkRuleID,
		ProviderAuditID: preflightdomain.ContentLinkProviderAuditID,
		LinkURL:         "https://example.com/guides",
		LinkText:        "Learn more",
		Selector:        "main a",
		IssueKey:        "content-link:event-1",
		Severity:        "medium",
		LatestEvidence:  `{"link_url":"https://example.com/guides","link_text":"Learn more"}`,
		FirstDetectedAt: now,
		LastDetectedAt:  now,
	}
	require.NoError(t, NewPreflightContentLinkRepository(db).RecordDetections(&run, []preflightdomain.ContentLinkDetection{detection}))

	var event preflightdomain.ContentLinkIssueEvent
	require.NoError(t, db.Order("id ASC").First(&event).Error)
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(event.Metadata), &metadata))
	require.Equal(t, preflightdomain.ContentLinkRuleID, metadata["rule_id"])
	require.Equal(t, preflightdomain.ContentLinkProviderAuditID, metadata["provider_audit_id"])
}
