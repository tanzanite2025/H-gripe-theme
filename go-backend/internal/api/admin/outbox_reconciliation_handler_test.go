package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOutboxReconciliationHandlerListsUnknownEventsWithoutPayload(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outbox.Event{
		EventKey:       "webhook:unknown:1",
		EventType:      outbox.EventTypeOrderPaid,
		AggregateType:  outbox.AggregateTypeOrder,
		AggregateID:    "1",
		Payload:        []byte(`{"email":"buyer@example.test"}`),
		Status:         outbox.EventStatusUnknown,
		AvailableAt:    now,
		UncertainAt:    &now,
		ReconcileAfter: &now,
		LastError:      "timeout",
	}).Error)

	handler := NewOutboxReconciliationHandler(service.NewOutboxService(repository.NewOutboxRepository(db)))
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/outbox/unknown", nil)

	handler.ListUnknown(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Data struct {
			Count  int `json:"count"`
			Events []struct {
				EventKey string `json:"event_key"`
				Payload  string `json:"payload"`
			} `json:"events"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 1, body.Data.Count)
	require.Len(t, body.Data.Events, 1)
	require.Equal(t, "webhook:unknown:1", body.Data.Events[0].EventKey)
	require.Empty(t, body.Data.Events[0].Payload)
}

func TestOutboxReconciliationHandlerRequiresNoteAndResumesEvent(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	now := time.Now().UTC()
	event := outbox.Event{
		EventKey:       "webhook:unknown:2",
		EventType:      outbox.EventTypeOrderPaid,
		AggregateType:  outbox.AggregateTypeOrder,
		AggregateID:    "2",
		Payload:        []byte(`{}`),
		Status:         outbox.EventStatusUnknown,
		AvailableAt:    now,
		UncertainAt:    &now,
		ReconcileAfter: &now,
	}
	require.NoError(t, db.Create(&event).Error)
	handler := NewOutboxReconciliationHandler(service.NewOutboxService(repository.NewOutboxRepository(db)))

	invalidRecorder := httptest.NewRecorder()
	invalidContext, _ := gin.CreateTestContext(invalidRecorder)
	invalidContext.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/ops/outbox/unknown/"+strconv.FormatUint(uint64(event.ID), 10)+"/resume",
		bytes.NewBufferString(`{"note":"x"}`),
	)
	invalidContext.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(event.ID), 10)}}
	handler.Resume(invalidContext)
	require.Equal(t, http.StatusBadRequest, invalidRecorder.Code)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/ops/outbox/unknown/"+strconv.FormatUint(uint64(event.ID), 10)+"/resume",
		bytes.NewBufferString(`{"note":"provider query confirmed no side effect"}`),
	)
	context.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(event.ID), 10)}}
	handler.Resume(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var saved outbox.Event
	require.NoError(t, db.First(&saved, event.ID).Error)
	require.Equal(t, outbox.EventStatusFailed, saved.Status)
	require.Nil(t, saved.UncertainAt)
	require.Nil(t, saved.ReconcileAfter)
}

func TestOutboxReconciliationHandlerMarksEventProcessed(t *testing.T) {
	db := newOutboxReconciliationTestDB(t)
	now := time.Now().UTC()
	event := outbox.Event{
		EventKey:       "webhook:unknown:3",
		EventType:      outbox.EventTypeOrderPaid,
		AggregateType:  outbox.AggregateTypeOrder,
		AggregateID:    "3",
		Payload:        []byte(`{}`),
		Status:         outbox.EventStatusUnknown,
		AvailableAt:    now,
		UncertainAt:    &now,
		ReconcileAfter: &now,
	}
	require.NoError(t, db.Create(&event).Error)
	handler := NewOutboxReconciliationHandler(service.NewOutboxService(repository.NewOutboxRepository(db)))

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/ops/outbox/unknown/"+strconv.FormatUint(uint64(event.ID), 10)+"/mark-processed",
		bytes.NewBufferString(`{"note":"provider query confirmed delivery"}`),
	)
	context.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(event.ID), 10)}}
	handler.MarkProcessed(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var saved outbox.Event
	require.NoError(t, db.First(&saved, event.ID).Error)
	require.Equal(t, outbox.EventStatusProcessed, saved.Status)
	require.NotNil(t, saved.ProcessedAt)
}

func newOutboxReconciliationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	return db
}
