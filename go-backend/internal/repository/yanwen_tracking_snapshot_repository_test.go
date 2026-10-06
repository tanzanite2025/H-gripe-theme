package repository

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenTrackingSnapshotRepositoryUpsertsLatestOfficialResultByEnvironmentAndNumber(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenTrackingSnapshot{}))
	repository := NewYanwenTrackingSnapshotRepository(db)

	syncedAt := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	snapshot := &shipping.YanwenTrackingSnapshot{
		Environment:               "production",
		TrackingNumber:            "UH123",
		WaybillNumber:             "UH123",
		TrackingStatus:            "OR10",
		LatestCheckpointStatus:    "OR10",
		LatestCheckpointTimeStamp: "2026-10-05T12:00:00",
		CheckpointsData:           datatypes.JSON([]byte(`[{"tracking_status":"OR10"}]`)),
		ResponseData:              datatypes.JSON([]byte(`{"code":0}`)),
		LastSyncedAt:              syncedAt,
		UpdatedAt:                 syncedAt,
	}
	require.NoError(t, repository.UpsertYanwenTrackingSnapshots([]*shipping.YanwenTrackingSnapshot{snapshot}))

	snapshot.TrackingStatus = "LH20"
	snapshot.LatestCheckpointStatus = "LH20"
	snapshot.LatestCheckpointTimeStamp = "2026-10-06T12:00:00"
	snapshot.LastSyncedAt = syncedAt.Add(time.Hour)
	snapshot.UpdatedAt = snapshot.LastSyncedAt
	require.NoError(t, repository.UpsertYanwenTrackingSnapshots([]*shipping.YanwenTrackingSnapshot{snapshot}))

	stored, err := repository.FindYanwenTrackingSnapshotByEnvironmentAndTrackingNumber("production", "UH123")
	require.NoError(t, err)
	require.Equal(t, "LH20", stored.TrackingStatus)
	require.Equal(t, "2026-10-06T12:00:00", stored.LatestCheckpointTimeStamp)
	require.Equal(t, syncedAt.Add(time.Hour), stored.LastSyncedAt)

	all, err := repository.FindYanwenTrackingSnapshotsByEnvironment("production")
	require.NoError(t, err)
	require.Len(t, all, 1)
}

func TestYanwenTrackingSnapshotRepositoryRejectsFATSnapshot(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenTrackingSnapshot{}))
	repository := NewYanwenTrackingSnapshotRepository(db)

	err = repository.UpsertYanwenTrackingSnapshots([]*shipping.YanwenTrackingSnapshot{{
		Environment:    "fat",
		TrackingNumber: "UH123",
		WaybillNumber:  "UH123",
		TrackingStatus: "OR10",
		LastSyncedAt:   time.Now().UTC(),
	}})
	require.ErrorContains(t, err, "environment must be production")
}

func TestYanwenTrackingSnapshotRepositoryRegistersClaimsAndFinalizesPollingTargets(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenTrackingSnapshot{}))
	repository := NewYanwenTrackingSnapshotRepository(db)

	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	require.NoError(t, repository.RegisterProductionTrackingNumbers(context.Background(), []string{"UH-POLL", "UH-POLL"}, now))
	claimed, err := repository.ClaimDueProductionTrackingSnapshots(context.Background(), now, now.Add(10*time.Minute), 20, "lease-1")
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	require.Equal(t, shipping.YanwenTrackingPollingStatePending, claimed[0].PollingState)
	require.False(t, claimed[0].HasOfficialResult)

	require.NoError(t, repository.MarkProductionTrackingPollingFailed(
		context.Background(), "UH-POLL", "lease-1", now, now.Add(time.Hour), "official endpoint unavailable",
	))
	failed, err := repository.FindYanwenTrackingSnapshotByEnvironmentAndTrackingNumber("production", "UH-POLL")
	require.NoError(t, err)
	require.Equal(t, shipping.YanwenTrackingPollingStatePending, failed.PollingState)
	require.Equal(t, 1, failed.PollingFailureCount)
	require.Equal(t, "official endpoint unavailable", failed.LastPollingError)
	require.Empty(t, failed.PollingLeaseToken)

	claimed, err = repository.ClaimDueProductionTrackingSnapshots(context.Background(), now.Add(2*time.Hour), now.Add(2*time.Hour+10*time.Minute), 20, "lease-2")
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	next := now.Add(3 * time.Hour)
	require.NoError(t, repository.MarkProductionTrackingPollingSucceeded(
		context.Background(), "UH-POLL", "lease-2", now.Add(2*time.Hour), &next, shipping.YanwenTrackingPollingStateActive,
	))
	active, err := repository.FindYanwenTrackingSnapshotByEnvironmentAndTrackingNumber("production", "UH-POLL")
	require.NoError(t, err)
	require.Equal(t, shipping.YanwenTrackingPollingStateActive, active.PollingState)
	require.Equal(t, 0, active.PollingFailureCount)
	require.Empty(t, active.LastPollingError)
	require.Equal(t, next, active.NextSyncAt.UTC())
}
