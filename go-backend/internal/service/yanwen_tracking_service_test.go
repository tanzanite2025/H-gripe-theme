package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenTrackingServiceUsesStoredProductionAuthorizationWithoutOrderToken(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "tracking-test-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, "TRACKING-USER", request.Header.Get("Authorization"))
		_, _ = writer.Write([]byte(`{"code":0,"message":"success","result":[]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepository := repository.NewYanwenAPIConfigRepository(db)
	apiService := NewYanwenAPIService(configRepository, nil)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "production",
		Endpoint:    yanwenDefaultEndpoint,
		UserID:      "TRACKING-USER",
		Enabled:     true,
	}))

	results, err := apiService.QueryYanwenTracking(context.Background(), []string{"UH123"})
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestYanwenTrackingServicePersistsOfficialResultsInYanwenOnlySnapshotRepository(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "tracking-snapshot-test-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenTrackingSnapshot{}))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"code":0,"message":"success","result":[{"tracking_number":"UH123","waybill_number":"UH123","checkpoints":[{"time_stamp":"2026-10-05T12:00:00","time_zone":"+08","tracking_status":"OR10","message":"accepted","is_last_mile_checkpoint":0}],"tracking_status":"OR10","tracking_status_waybill":{"level1":"1","level2":"1","level3":"OR10"},"last_mile_tracking_expected":false,"origin_country":"CN","destination_country":"DE"}]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepository := repository.NewYanwenAPIConfigRepository(db)
	snapshotRepository := repository.NewYanwenTrackingSnapshotRepository(db)
	apiService := NewYanwenAPIService(configRepository, nil)
	apiService.ConfigureYanwenTrackingSnapshotRepository(snapshotRepository)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "production",
		Endpoint:    yanwenDefaultEndpoint,
		UserID:      "TRACKING-USER",
		Enabled:     true,
	}))

	results, err := apiService.QueryYanwenTracking(context.Background(), []string{"UH123"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	stored, err := snapshotRepository.FindYanwenTrackingSnapshotByEnvironmentAndTrackingNumber("production", "UH123")
	require.NoError(t, err)
	require.Equal(t, "OR10", stored.TrackingStatus)
	require.Equal(t, "OR10", stored.LatestCheckpointStatus)
	require.JSONEq(t, `[{
      "time_stamp":"2026-10-05T12:00:00",
      "time_zone":"+08",
      "tracking_status":"OR10",
      "message":"accepted",
      "location":"",
      "is_last_mile_checkpoint":false
    }]`, string(stored.CheckpointsData))
}

func TestYanwenTrackingPollingServiceRefreshesOnlyItsSnapshotAndStopsAtDocumentedTerminalStatus(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "tracking-polling-test-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenTrackingSnapshot{}))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, "TRACKING-USER", request.Header.Get("Authorization"))
		_, _ = writer.Write([]byte(`{"code":0,"message":"success","result":[{"tracking_number":"UH-POLL","waybill_number":"UH-POLL","checkpoints":[{"time_stamp":"2026-10-05T12:00:00","time_zone":"+08","tracking_status":"LM40","message":"delivered","is_last_mile_checkpoint":1}],"tracking_status":"LM40","tracking_status_waybill":{"level1":"4","level2":"4","level3":"LM40"},"last_mile_tracking_expected":true,"origin_country":"CN","destination_country":"DE"}]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepository := repository.NewYanwenAPIConfigRepository(db)
	snapshotRepository := repository.NewYanwenTrackingSnapshotRepository(db)
	apiService := NewYanwenAPIService(configRepository, nil)
	apiService.ConfigureYanwenTrackingSnapshotRepository(snapshotRepository)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "production",
		Endpoint:    yanwenDefaultEndpoint,
		UserID:      "TRACKING-USER",
		Enabled:     true,
	}))
	require.NoError(t, snapshotRepository.RegisterProductionTrackingNumbers(context.Background(), []string{"UH-POLL"}, time.Now().UTC()))

	pollingService := NewYanwenTrackingPollingService(apiService, snapshotRepository, time.Second, 20)
	result, err := pollingService.PollDueYanwenTrackingTargets(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, result.Claimed)
	require.Equal(t, 1, result.Synced)
	require.Equal(t, 1, result.Terminal)
	require.Zero(t, result.Failed)

	stored, err := snapshotRepository.FindYanwenTrackingSnapshotByEnvironmentAndTrackingNumber("production", "UH-POLL")
	require.NoError(t, err)
	require.True(t, stored.HasOfficialResult)
	require.Equal(t, shipping.YanwenTrackingPollingStateTerminal, stored.PollingState)
	require.Nil(t, stored.NextSyncAt)
	require.Equal(t, "LM40", stored.TrackingStatus)
	require.Empty(t, stored.PollingLeaseToken)
}

func TestYanwenTrackingPollingServiceRecordsFailureWithoutClearingPreviousOfficialSnapshot(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "tracking-polling-failure-test-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenTrackingSnapshot{}))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte(`gateway unavailable`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepository := repository.NewYanwenAPIConfigRepository(db)
	snapshotRepository := repository.NewYanwenTrackingSnapshotRepository(db)
	apiService := NewYanwenAPIService(configRepository, nil)
	apiService.ConfigureYanwenTrackingSnapshotRepository(snapshotRepository)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{
		Environment: "production",
		Endpoint:    yanwenDefaultEndpoint,
		UserID:      "TRACKING-USER",
		Enabled:     true,
	}))
	now := time.Now().UTC().Add(-time.Hour)
	require.NoError(t, snapshotRepository.UpsertYanwenTrackingSnapshots([]*shipping.YanwenTrackingSnapshot{{
		Environment:       "production",
		TrackingNumber:    "UH-FAIL",
		WaybillNumber:     "UH-FAIL",
		TrackingStatus:    "OR10",
		HasOfficialResult: true,
		PollingState:      shipping.YanwenTrackingPollingStateActive,
		NextSyncAt:        &now,
		LastSyncedAt:      now,
		CheckpointsData:   datatypes.JSON([]byte(`[{"tracking_status":"OR10"}]`)),
		ResponseData:      datatypes.JSON([]byte(`{"code":0}`)),
	}}))

	pollingService := NewYanwenTrackingPollingService(apiService, snapshotRepository, time.Second, 20)
	result, err := pollingService.PollDueYanwenTrackingTargets(context.Background())
	require.Error(t, err)
	require.Equal(t, 1, result.Claimed)
	require.Equal(t, 1, result.Failed)

	stored, err := snapshotRepository.FindYanwenTrackingSnapshotByEnvironmentAndTrackingNumber("production", "UH-FAIL")
	require.NoError(t, err)
	require.True(t, stored.HasOfficialResult)
	require.Equal(t, "OR10", stored.TrackingStatus)
	require.Equal(t, 1, stored.PollingFailureCount)
	require.Contains(t, stored.LastPollingError, "Yanwen tracking endpoint returned HTTP 502")
	require.Empty(t, stored.PollingLeaseToken)
}
