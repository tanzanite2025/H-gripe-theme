package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDeriveStorefrontURLIssueDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		entry    seodomain.StorefrontRouteCatalogEntry
		expected string
	}{
		{
			name: "duplicate route becomes collision",
			entry: seodomain.StorefrontRouteCatalogEntry{
				EntryStatus: seodomain.RouteEntryStatusDuplicate,
			},
			expected: urlmanagementdomain.URLIssueTypePathCollision,
		},
		{
			name: "stale route is tracked",
			entry: seodomain.StorefrontRouteCatalogEntry{
				EntryStatus: seodomain.RouteEntryStatusStale,
			},
			expected: urlmanagementdomain.URLIssueTypeStaleRoute,
		},
		{
			name: "stale route does not reuse a historical not found check",
			entry: seodomain.StorefrontRouteCatalogEntry{
				EntryStatus:     seodomain.RouteEntryStatusStale,
				LastCheckStatus: seodomain.RouteCheckStatusNotFound,
			},
			expected: urlmanagementdomain.URLIssueTypeStaleRoute,
		},
		{
			name: "freshly recovered stale route has no active issue",
			entry: seodomain.StorefrontRouteCatalogEntry{
				EntryStatus:     seodomain.RouteEntryStatusStale,
				LastCheckStatus: seodomain.RouteCheckStatusOK,
				LastCheckedAt:   storefrontIssueCheckTimePointer(time.Now().UTC()),
			},
			expected: "",
		},
		{
			name: "stale route returning 404 is confirmed retired",
			entry: seodomain.StorefrontRouteCatalogEntry{
				EntryStatus:     seodomain.RouteEntryStatusStale,
				LastCheckStatus: seodomain.RouteCheckStatusNotFound,
				LastCheckedAt:   storefrontIssueCheckTimePointer(time.Now().UTC()),
			},
			expected: "",
		},
		{
			name: "stale route with an unresolved server error remains actionable",
			entry: seodomain.StorefrontRouteCatalogEntry{
				EntryStatus:     seodomain.RouteEntryStatusStale,
				LastCheckStatus: seodomain.RouteCheckStatusServerError,
				LastCheckedAt:   storefrontIssueCheckTimePointer(time.Now().UTC()),
			},
			expected: urlmanagementdomain.URLIssueTypeStaleRoute,
		},
		{
			name: "alias redirect chain remains actionable",
			entry: seodomain.StorefrontRouteCatalogEntry{
				IsAlias:         true,
				LastCheckStatus: seodomain.RouteCheckStatusRedirectChain,
			},
			expected: urlmanagementdomain.URLIssueTypeRedirectChain,
		},
		{
			name: "canonical route redirect is a status mismatch",
			entry: seodomain.StorefrontRouteCatalogEntry{
				LastCheckStatus: seodomain.RouteCheckStatusRedirect,
			},
			expected: urlmanagementdomain.URLIssueTypeRedirectStatusMismatch,
		},
		{
			name: "healthy alias redirect is not an issue",
			entry: seodomain.StorefrontRouteCatalogEntry{
				IsAlias:         true,
				LastCheckStatus: seodomain.RouteCheckStatusRedirect,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definitions := deriveStorefrontURLIssueDefinitions(tt.entry)
			if tt.expected == "" {
				if len(definitions) != 0 {
					t.Fatalf("expected no issue definitions, got %#v", definitions)
				}
				return
			}
			for _, definition := range definitions {
				if definition.issueType == tt.expected {
					return
				}
			}
			t.Fatalf("expected issue %q, got %#v", tt.expected, definitions)
		})
	}
}

func storefrontIssueCheckTimePointer(value time.Time) *time.Time {
	return &value
}

func TestStaleRouteOnlyProducesStaleIssue(t *testing.T) {
	definitions := deriveStorefrontURLIssueDefinitions(seodomain.StorefrontRouteCatalogEntry{
		EntryStatus:     seodomain.RouteEntryStatusStale,
		LastCheckStatus: seodomain.RouteCheckStatusNotFound,
	})
	if len(definitions) != 1 {
		t.Fatalf("expected one stale issue definition, got %#v", definitions)
	}
	if definitions[0].issueType != urlmanagementdomain.URLIssueTypeStaleRoute {
		t.Fatalf("issue type = %q, want %q", definitions[0].issueType, urlmanagementdomain.URLIssueTypeStaleRoute)
	}
}

func TestResolveRetiredStaleRouteClosesWithoutVerificationProbe(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
	))

	now := time.Now().UTC()
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey: "manifest:stale:en", Path: "/legacy", Locale: "en", SourceType: seodomain.RouteSourceStatic,
		EntryStatus: seodomain.RouteEntryStatusStale, IsCheckable: true, LastSeenAt: now,
	}
	require.NoError(t, db.Create(&entry).Error)
	issue := urlmanagementdomain.StorefrontURLIssue{
		RouteEntryID:    entry.ID,
		IssueType:       urlmanagementdomain.URLIssueTypeStaleRoute,
		Severity:        urlmanagementdomain.URLIssueSeverityMedium,
		State:           urlmanagementdomain.URLIssueStateOpen,
		FirstDetectedAt: now,
		LastDetectedAt:  now,
	}
	require.NoError(t, db.Create(&issue).Error)

	service := NewStorefrontURLIssueService(
		repository.NewStorefrontURLIssueRepository(db),
		repository.NewStorefrontRouteCatalogRepository(db),
		nil,
	)
	resolved, err := service.Resolve(issue.ID, 7, urlmanagementdomain.StorefrontURLIssueResolutionInput{
		ResolutionType: urlmanagementdomain.URLIssueResolutionRetired,
		ResolutionNote: "旧路径已正式退役",
	})
	require.NoError(t, err)
	require.Equal(t, urlmanagementdomain.URLIssueStateVerified, resolved.State)
	require.Equal(t, urlmanagementdomain.URLIssueResolutionRetired, resolved.ResolutionType)

	var event urlmanagementdomain.StorefrontURLIssueEvent
	require.NoError(t, db.Where("issue_id = ?", issue.ID).Order("id DESC").First(&event).Error)
	require.Equal(t, urlmanagementdomain.URLIssueEventStaleRouteRetired, event.EventType)
}

func TestLinkRedirectRejectsUnrelatedSourcePath(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
		&urlmanagementdomain.StorefrontRedirectRule{},
	))

	now := time.Now().UTC()
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey: "manifest:legacy:en", Path: "/legacy", Locale: "en", SourceType: seodomain.RouteSourceStatic,
		EntryStatus: seodomain.RouteEntryStatusStale, IsCheckable: true, LastSeenAt: now,
	}
	require.NoError(t, db.Create(&entry).Error)
	issue := urlmanagementdomain.StorefrontURLIssue{
		RouteEntryID: entry.ID, IssueType: urlmanagementdomain.URLIssueTypeStaleRoute,
		Severity: urlmanagementdomain.URLIssueSeverityMedium, State: urlmanagementdomain.URLIssueStateOpen,
		FirstDetectedAt: now, LastDetectedAt: now,
	}
	require.NoError(t, db.Create(&issue).Error)
	rule := urlmanagementdomain.StorefrontRedirectRule{
		SourcePath: "/other-legacy", TargetPath: "/support", StatusCode: 301,
		State: urlmanagementdomain.RedirectRuleStateDraft, Reason: "迁移旧路径",
	}
	require.NoError(t, db.Create(&rule).Error)

	issueService := NewStorefrontURLIssueService(
		repository.NewStorefrontURLIssueRepository(db),
		repository.NewStorefrontRouteCatalogRepository(db),
		repository.NewStorefrontRedirectRuleRepository(db),
	)
	_, err = issueService.LinkRedirect(issue.ID, 7, urlmanagementdomain.StorefrontURLIssueLinkRedirectInput{
		RedirectRuleID: rule.ID,
	})
	require.ErrorContains(t, err, "does not match URL issue path")
}

func TestResolveRedirectPublishedRequiresPublishedMatchingRedirect(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
		&urlmanagementdomain.StorefrontRedirectRule{},
	))

	now := time.Now().UTC()
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey: "manifest:not-found:en", Path: "/legacy", Locale: "en", SourceType: seodomain.RouteSourceStatic,
		EntryStatus: seodomain.RouteEntryStatusActive, IsCheckable: true, LastSeenAt: now,
	}
	require.NoError(t, db.Create(&entry).Error)
	issue := urlmanagementdomain.StorefrontURLIssue{
		RouteEntryID: entry.ID, IssueType: urlmanagementdomain.URLIssueTypeNotFound,
		Severity: urlmanagementdomain.URLIssueSeverityHigh, State: urlmanagementdomain.URLIssueStateOpen,
		FirstDetectedAt: now, LastDetectedAt: now,
	}
	require.NoError(t, db.Create(&issue).Error)
	rule := urlmanagementdomain.StorefrontRedirectRule{
		SourcePath: "/legacy", TargetPath: "/support", StatusCode: 301,
		State: urlmanagementdomain.RedirectRuleStateDraft, Reason: "迁移旧路径",
	}
	require.NoError(t, db.Create(&rule).Error)

	issueService := NewStorefrontURLIssueService(
		repository.NewStorefrontURLIssueRepository(db),
		repository.NewStorefrontRouteCatalogRepository(db),
		repository.NewStorefrontRedirectRuleRepository(db),
	)
	_, err = issueService.Resolve(issue.ID, 7, urlmanagementdomain.StorefrontURLIssueResolutionInput{
		ResolutionType:       urlmanagementdomain.URLIssueResolutionRedirectPublished,
		ResolutionNote:       "已发布永久重定向",
		LinkedRedirectRuleID: nil,
	})
	require.ErrorContains(t, err, "requires a linked redirect rule")

	_, err = issueService.Resolve(issue.ID, 7, urlmanagementdomain.StorefrontURLIssueResolutionInput{
		ResolutionType:       urlmanagementdomain.URLIssueResolutionRedirectPublished,
		ResolutionNote:       "已发布永久重定向",
		LinkedRedirectRuleID: &rule.ID,
	})
	require.ErrorContains(t, err, "requires a published redirect rule")

	require.NoError(t, db.Model(&rule).Update("state", urlmanagementdomain.RedirectRuleStatePublished).Error)
	otherRule := urlmanagementdomain.StorefrontRedirectRule{
		SourcePath: "/different-legacy", TargetPath: "/support", StatusCode: 301,
		State: urlmanagementdomain.RedirectRuleStatePublished, Reason: "迁移另一条旧路径",
	}
	require.NoError(t, db.Create(&otherRule).Error)
	_, err = issueService.Resolve(issue.ID, 7, urlmanagementdomain.StorefrontURLIssueResolutionInput{
		ResolutionType:       urlmanagementdomain.URLIssueResolutionRedirectPublished,
		ResolutionNote:       "已发布永久重定向",
		LinkedRedirectRuleID: &otherRule.ID,
	})
	require.ErrorContains(t, err, "does not match URL issue path")
}

func TestResolveRetiredRejectsNonStaleRouteIssue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
	))

	now := time.Now().UTC()
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey: "manifest:not-found:en", Path: "/missing", Locale: "en", SourceType: seodomain.RouteSourceStatic,
		EntryStatus: seodomain.RouteEntryStatusActive, IsCheckable: true, LastSeenAt: now,
	}
	require.NoError(t, db.Create(&entry).Error)
	issue := urlmanagementdomain.StorefrontURLIssue{
		RouteEntryID: entry.ID, IssueType: urlmanagementdomain.URLIssueTypeNotFound,
		Severity: urlmanagementdomain.URLIssueSeverityHigh, State: urlmanagementdomain.URLIssueStateOpen,
		FirstDetectedAt: now, LastDetectedAt: now,
	}
	require.NoError(t, db.Create(&issue).Error)

	issueService := NewStorefrontURLIssueService(
		repository.NewStorefrontURLIssueRepository(db),
		repository.NewStorefrontRouteCatalogRepository(db),
		nil,
	)
	_, err = issueService.Resolve(issue.ID, 7, urlmanagementdomain.StorefrontURLIssueResolutionInput{
		ResolutionType: urlmanagementdomain.URLIssueResolutionRetired,
		ResolutionNote: "该问题已退役",
	})
	require.ErrorContains(t, err, "only stale-route URL issues can be marked retired")
}

func TestRecheckAutomaticallyClosesRecoveredStaleRouteIssue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&seodomain.StorefrontRouteCheckResult{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
	))

	now := time.Now().UTC()
	entry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey: "manifest:stale:en", Path: "/legacy", Locale: "en", SourceType: seodomain.RouteSourceStatic,
		CanonicalPath: "/legacy", EntryStatus: seodomain.RouteEntryStatusStale, IsCheckable: true, LastSeenAt: now,
	}
	require.NoError(t, db.Create(&entry).Error)
	stillFailingEntry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey: "manifest:stale-failing:en", Path: "/still-failing", Locale: "en", SourceType: seodomain.RouteSourceStatic,
		EntryStatus: seodomain.RouteEntryStatusStale, IsCheckable: true, LastSeenAt: now,
	}
	require.NoError(t, db.Create(&stillFailingEntry).Error)
	retiredEntry := seodomain.StorefrontRouteCatalogEntry{
		RouteKey: "manifest:stale-retired:en", Path: "/retired", Locale: "en", SourceType: seodomain.RouteSourceStatic,
		EntryStatus: seodomain.RouteEntryStatusStale, IsCheckable: true, LastSeenAt: now,
	}
	require.NoError(t, db.Create(&retiredEntry).Error)
	issues := []urlmanagementdomain.StorefrontURLIssue{
		{
			RouteEntryID:    entry.ID,
			IssueType:       urlmanagementdomain.URLIssueTypeStaleRoute,
			Severity:        urlmanagementdomain.URLIssueSeverityMedium,
			State:           urlmanagementdomain.URLIssueStateOpen,
			FirstDetectedAt: now,
			LastDetectedAt:  now,
		},
		{
			RouteEntryID:    stillFailingEntry.ID,
			IssueType:       urlmanagementdomain.URLIssueTypeStaleRoute,
			Severity:        urlmanagementdomain.URLIssueSeverityMedium,
			State:           urlmanagementdomain.URLIssueStateOpen,
			FirstDetectedAt: now,
			LastDetectedAt:  now,
		},
		{
			RouteEntryID:    retiredEntry.ID,
			IssueType:       urlmanagementdomain.URLIssueTypeStaleRoute,
			Severity:        urlmanagementdomain.URLIssueSeverityMedium,
			State:           urlmanagementdomain.URLIssueStateOpen,
			FirstDetectedAt: now,
			LastDetectedAt:  now,
		},
	}
	require.NoError(t, db.Create(&issues).Error)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/still-failing" {
			http.Error(writer, "still failing", http.StatusInternalServerError)
			return
		}
		if request.URL.Path == "/retired" {
			http.NotFound(writer, request)
			return
		}
		require.Equal(t, "/legacy", request.URL.Path)
		writer.Header().Set("Content-Type", "text/html")
		_, writeErr := writer.Write([]byte(`<html><head><link rel="canonical" href="/legacy"></head><body>restored</body></html>`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	routeRepository := repository.NewStorefrontRouteCatalogRepository(db)
	issueRepository := repository.NewStorefrontURLIssueRepository(db)
	issueService := NewStorefrontURLIssueService(issueRepository, routeRepository, nil)
	catalog := NewStorefrontRouteCatalogService(routeRepository, nil, nil, server.URL, server.URL)
	catalog.ConfigureIssueReconciler(issueService)

	result, err := catalog.CheckStaleRouteEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, seodomain.RouteCheckStatusOK, result.Status)
	failingResult, err := catalog.CheckStaleRouteEntry(context.Background(), stillFailingEntry.ID)
	require.NoError(t, err)
	require.Equal(t, seodomain.RouteCheckStatusServerError, failingResult.Status)
	retiredResult, err := catalog.CheckStaleRouteEntry(context.Background(), retiredEntry.ID)
	require.NoError(t, err)
	require.Equal(t, seodomain.RouteCheckStatusNotFound, retiredResult.Status)

	updatedIssue, err := issueRepository.FindByID(issues[0].ID)
	require.NoError(t, err)
	require.Equal(t, urlmanagementdomain.URLIssueStateVerified, updatedIssue.State)
	require.Equal(t, urlmanagementdomain.URLIssueResolutionRuntimeFixed, updatedIssue.ResolutionType)
	stillActiveIssue, err := issueRepository.FindByID(issues[1].ID)
	require.NoError(t, err)
	require.Equal(t, urlmanagementdomain.URLIssueStateOpen, stillActiveIssue.State)
	retiredIssue, err := issueRepository.FindByID(issues[2].ID)
	require.NoError(t, err)
	require.Equal(t, urlmanagementdomain.URLIssueStateVerified, retiredIssue.State)
	require.Equal(t, urlmanagementdomain.URLIssueResolutionRetired, retiredIssue.ResolutionType)

	activeIssues, total, err := issueRepository.List(repository.StorefrontURLIssueListFilter{State: "active"})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, activeIssues, 1)
	require.Equal(t, urlmanagementdomain.URLIssueTypeStaleRoute, activeIssues[0].IssueType)

	var verificationEvent urlmanagementdomain.StorefrontURLIssueEvent
	require.NoError(t, db.Where("issue_id = ? AND event_type = ?", issues[0].ID, urlmanagementdomain.URLIssueEventVerificationPassed).
		First(&verificationEvent).Error)
	var retirementEvent urlmanagementdomain.StorefrontURLIssueEvent
	require.NoError(t, db.Where("issue_id = ? AND event_type = ?", issues[2].ID, urlmanagementdomain.URLIssueEventStaleRouteRetired).
		First(&retirementEvent).Error)
}

func TestRouteCatalogSyncKeepsActiveRuntimeIssueAndVerifiesResolvedPathCollision(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&urlmanagementdomain.StorefrontURLIssue{},
		&urlmanagementdomain.StorefrontURLIssueEvent{},
	))

	checkedAt := time.Now().UTC().Add(-time.Hour)
	entries := []seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey: "manifest:runtime-error:en", Path: "/runtime-error", Locale: "en",
			SourceType: seodomain.RouteSourceStatic, CanonicalPath: "/runtime-error",
			EntryStatus: seodomain.RouteEntryStatusActive, IsCheckable: true,
			LastCheckStatus: seodomain.RouteCheckStatusServerError, LastHTTPStatus: http.StatusInternalServerError,
			LastCheckedAt: storefrontIssueCheckTimePointer(checkedAt), LastSeenAt: checkedAt,
		},
		{
			RouteKey: "manifest:resolved-collision:en", Path: "/resolved-collision", Locale: "en",
			SourceType: seodomain.RouteSourceStatic, CanonicalPath: "/resolved-collision",
			EntryStatus: seodomain.RouteEntryStatusDuplicate, IsCheckable: true, LastSeenAt: checkedAt,
		},
	}
	require.NoError(t, db.Create(&entries).Error)
	issues := []urlmanagementdomain.StorefrontURLIssue{
		{
			RouteEntryID: entries[0].ID, IssueType: urlmanagementdomain.URLIssueTypeServerError,
			Severity: urlmanagementdomain.URLIssueSeverityHigh, State: urlmanagementdomain.URLIssueStateOpen,
			FirstDetectedAt: checkedAt, LastDetectedAt: checkedAt,
		},
		{
			RouteEntryID: entries[1].ID, IssueType: urlmanagementdomain.URLIssueTypePathCollision,
			Severity: urlmanagementdomain.URLIssueSeverityCritical, State: urlmanagementdomain.URLIssueStateOpen,
			FirstDetectedAt: checkedAt, LastDetectedAt: checkedAt,
		},
	}
	require.NoError(t, db.Create(&issues).Error)

	seenAt := time.Now().UTC()
	routeRepository := repository.NewStorefrontRouteCatalogRepository(db)
	require.NoError(t, routeRepository.UpsertSnapshot([]seodomain.StorefrontRouteCatalogEntry{
		{
			RouteKey: entries[0].RouteKey, Path: entries[0].Path, Locale: entries[0].Locale,
			SourceType: entries[0].SourceType, CanonicalPath: entries[0].CanonicalPath,
			EntryStatus: seodomain.RouteEntryStatusActive, IsCheckable: true,
		},
		{
			RouteKey: entries[1].RouteKey, Path: entries[1].Path, Locale: entries[1].Locale,
			SourceType: entries[1].SourceType, CanonicalPath: entries[1].CanonicalPath,
			EntryStatus: seodomain.RouteEntryStatusActive, IsCheckable: true,
		},
	}, seenAt))

	issueService := NewStorefrontURLIssueService(
		repository.NewStorefrontURLIssueRepository(db),
		routeRepository,
		nil,
	)
	require.NoError(t, issueService.ReconcileCatalog(context.Background()))

	activeRuntimeIssue, err := issueService.Get(issues[0].ID)
	require.NoError(t, err)
	require.Equal(t, urlmanagementdomain.URLIssueStateOpen, activeRuntimeIssue.State)
	verifiedCollisionIssue, err := issueService.Get(issues[1].ID)
	require.NoError(t, err)
	require.Equal(t, urlmanagementdomain.URLIssueStateVerified, verifiedCollisionIssue.State)
}
