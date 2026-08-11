package service

import (
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/recommendation"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBehaviorEventServiceIngestIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))

	eventService := NewBehaviorEventService(repository.NewRecommendationEventRepository(db))
	occurredAt := time.Now().UTC()
	input := BehaviorEventInput{
		EventID:     "event_test_001",
		EventType:   "product_view",
		AnonymousID: "anon_test",
		SessionID:   "session_test",
		ProductID:   behaviorTestUintPointer(42),
		Locale:      "en-US",
		Path:        "/shop/carbon-wheelset",
		Metadata: map[string]any{
			"surface":  "product_page",
			"position": float64(1),
		},
		OccurredAt: occurredAt,
	}

	first, err := eventService.Ingest(nil, []BehaviorEventInput{input})
	require.NoError(t, err)
	require.Equal(t, int64(1), first.Accepted)
	require.Equal(t, 0, first.Duplicates)

	second, err := eventService.Ingest(nil, []BehaviorEventInput{input})
	require.NoError(t, err)
	require.Equal(t, int64(0), second.Accepted)
	require.Equal(t, 1, second.Duplicates)

	var stored recommendation.Event
	require.NoError(t, db.First(&stored, "event_id = ?", input.EventID).Error)
	require.Equal(t, "product_view", stored.EventType)
	require.Equal(t, "en-us", stored.Locale)
	require.Equal(t, uint(42), *stored.ProductID)
	require.JSONEq(t, `{"position":1,"surface":"product_page"}`, string(stored.MetadataJSON))
}

func TestBehaviorEventServiceRejectsUnscopedEvents(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))

	eventService := NewBehaviorEventService(repository.NewRecommendationEventRepository(db))
	_, err = eventService.Ingest(nil, []BehaviorEventInput{{
		EventID:    "event_test_002",
		EventType:  "product_view",
		OccurredAt: time.Now().UTC(),
	}})
	require.ErrorIs(t, err, ErrBehaviorEventIdentityRequired)
}

func TestBehaviorEventServiceRejectsInvalidIdentityTokens(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))

	eventService := NewBehaviorEventService(repository.NewRecommendationEventRepository(db))
	_, err = eventService.Ingest(nil, []BehaviorEventInput{{
		EventID:     "event_test_invalid_identity",
		EventType:   "product_view",
		AnonymousID: strings.Repeat("a", 129),
		OccurredAt:  time.Now().UTC(),
	}})
	require.ErrorIs(t, err, ErrBehaviorEventIdentityInvalid)
}

func TestBehaviorEventServiceAcceptsCategoryNavigationEvents(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))

	eventService := NewBehaviorEventService(repository.NewRecommendationEventRepository(db))
	_, err = eventService.Ingest(nil, []BehaviorEventInput{{
		EventID:     "event_test_category_navigation",
		EventType:   "category_navigation_click",
		AnonymousID: "anon_test",
		CategoryID:  behaviorTestUintPointer(7),
		OccurredAt:  time.Now().UTC(),
	}})
	require.NoError(t, err)

	var stored recommendation.Event
	require.NoError(t, db.First(&stored, "event_id = ?", "event_test_category_navigation").Error)
	require.Equal(t, "category_navigation_click", stored.EventType)
	require.NotNil(t, stored.CategoryID)
	require.Equal(t, uint(7), *stored.CategoryID)
}

func TestBehaviorEventServiceAcceptsAdLandingButRejectsClientPurchaseEvents(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))

	eventService := NewBehaviorEventService(repository.NewRecommendationEventRepository(db))
	_, err = eventService.Ingest(nil, []BehaviorEventInput{{
		EventID:     "event_test_ad_landing",
		EventType:   "ad_landing",
		AnonymousID: "anon_test",
		Metadata: map[string]any{
			"gclid":     "click_123",
			"untrusted": "discarded",
		},
		OccurredAt: time.Now().UTC(),
	}})
	require.NoError(t, err)

	var stored recommendation.Event
	require.NoError(t, db.First(&stored, "event_id = ?", "event_test_ad_landing").Error)
	require.JSONEq(t, `{"gclid":"click_123","utm_source":"google"}`, string(stored.MetadataJSON))

	_, err = eventService.Ingest(nil, []BehaviorEventInput{{
		EventID:     "event_test_client_purchase",
		EventType:   "purchase",
		AnonymousID: "anon_test",
		OccurredAt:  time.Now().UTC(),
	}})
	require.ErrorIs(t, err, ErrBehaviorEventTypeInvalid)
}

func TestBehaviorEventServiceCleanupUsesIntentRetentionTiers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))

	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&[]recommendation.Event{
		behaviorTestEvent("old_page_view", "page_view", now.AddDate(0, 0, -31)),
		behaviorTestEvent("fresh_page_view", "page_view", now.AddDate(0, 0, -29)),
		behaviorTestEvent("old_product_view", "product_view", now.AddDate(0, 0, -61)),
		behaviorTestEvent("fresh_product_view", "product_view", now.AddDate(0, 0, -59)),
		behaviorTestEvent("old_add_to_cart", "add_to_cart", now.AddDate(0, 0, -181)),
		behaviorTestEvent("fresh_add_to_cart", "add_to_cart", now.AddDate(0, 0, -179)),
	}).Error)

	eventService := NewBehaviorEventService(
		repository.NewRecommendationEventRepository(db),
		config.BehaviorEventsConfig{
			LowIntentRetentionDays:      30,
			StandardIntentRetentionDays: 60,
			HighIntentRetentionDays:     180,
			CleanupBatchLimit:           100,
		},
	)

	result, err := eventService.CleanupExpiredEvents(now)
	require.NoError(t, err)
	require.EqualValues(t, 1, result.DeletedLowIntent)
	require.EqualValues(t, 1, result.DeletedStandardIntent)
	require.EqualValues(t, 1, result.DeletedHighIntent)
	require.EqualValues(t, 3, result.TotalDeleted)

	var remaining []recommendation.Event
	require.NoError(t, db.Order("event_id ASC").Find(&remaining).Error)
	require.Len(t, remaining, 3)
	require.Equal(t, []string{"fresh_add_to_cart", "fresh_page_view", "fresh_product_view"}, []string{
		remaining[0].EventID,
		remaining[1].EventID,
		remaining[2].EventID,
	})
}

func behaviorTestUintPointer(value uint) *uint {
	return &value
}

func behaviorTestEvent(eventID string, eventType string, occurredAt time.Time) recommendation.Event {
	return recommendation.Event{
		EventID:      eventID,
		EventType:    eventType,
		AnonymousID:  "anon_test",
		MetadataJSON: []byte(`{}`),
		OccurredAt:   occurredAt,
		ReceivedAt:   occurredAt,
	}
}
