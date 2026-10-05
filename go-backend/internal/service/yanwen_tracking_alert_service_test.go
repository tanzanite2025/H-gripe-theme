package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestBuildYanwenTrackingAlertsFromOfficialSnapshotsCreatesCustomsStagnationAlert(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	alerts := buildYanwenTrackingAlertsFromOfficialSnapshots([]shipping.YanwenTrackingSnapshot{{
		Environment:               "production",
		TrackingNumber:            "YW-CUSTOMS-1",
		WaybillNumber:             "YW-CUSTOMS-1",
		DestinationCountry:        "US",
		TrackingStatus:            "S303",
		LatestCheckpointStatus:    "S303",
		LatestCheckpointTimeStamp: "2026-10-02T11:00:00",
		CheckpointsData: datatypes.JSON([]byte(`[
			{"time_stamp":"2026-10-02T11:00:00","time_zone":"+08","tracking_status":"S303","is_last_mile_checkpoint":false}
		]`)),
		HasOfficialResult: true,
		LastSyncedAt:      now.Add(-time.Hour),
	}}, now)

	require.Len(t, alerts, 1)
	require.Equal(t, yanwenCustomsStagnationAlertType, alerts[0].AlertType)
	require.Equal(t, yanwenAlertSeverityAlert, alerts[0].Severity)
	require.Equal(t, int64(81), alerts[0].ElapsedHours)
	require.Equal(t, "+08", alerts[0].CheckpointTimeZone)
	require.Equal(t, "S303", alerts[0].TrackingStatus)
}

func TestBuildYanwenTrackingAlertsFromOfficialSnapshotsRequiresReliableOfficialTime(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	alerts := buildYanwenTrackingAlertsFromOfficialSnapshots([]shipping.YanwenTrackingSnapshot{{
		Environment:               "production",
		TrackingNumber:            "YW-BAD-TIME",
		WaybillNumber:             "YW-BAD-TIME",
		TrackingStatus:            "S303",
		LatestCheckpointStatus:    "S303",
		LatestCheckpointTimeStamp: "2026-10-02T11:00:00",
		CheckpointsData: datatypes.JSON([]byte(`[
			{"time_stamp":"2026-10-02T11:00:00","time_zone":"not-a-time-zone","tracking_status":"S303","is_last_mile_checkpoint":false}
		]`)),
		HasOfficialResult: true,
	}}, now)

	require.Empty(t, alerts)
}

func TestBuildYanwenTrackingAlertsFromOfficialSnapshotsCreatesInTransitTimeoutOnlyWithoutExplicitDestinationNode(t *testing.T) {
	now := time.Date(2026, 10, 2, 9, 0, 0, 0, time.FixedZone("+08", 8*60*60))
	alerts := buildYanwenTrackingAlertsFromOfficialSnapshots([]shipping.YanwenTrackingSnapshot{{
		Environment:               "production",
		TrackingNumber:            "YW-AIR-1",
		WaybillNumber:             "YW-AIR-1",
		DestinationCountry:        "DE",
		TrackingStatus:            "LH20",
		LatestCheckpointStatus:    "LH20",
		LatestCheckpointTimeStamp: "2026-09-21T09:00:00",
		CheckpointsData: datatypes.JSON([]byte(`[
			{"time_stamp":"2026-09-21T09:00:00","time_zone":"+08","tracking_status":"LH20","is_last_mile_checkpoint":false},
			{"time_stamp":"2026-09-22T11:00:00","time_zone":"+08","tracking_status":"UNKNOWN-LOCATION-ONLY","location":"DE","is_last_mile_checkpoint":false}
		]`)),
		HasOfficialResult: true,
	}}, now.UTC())

	require.Len(t, alerts, 1)
	require.Equal(t, yanwenInTransitTimeoutAlertType, alerts[0].AlertType)
	require.Equal(t, yanwenAlertSeverityCritical, alerts[0].Severity)
	require.Equal(t, 9, alerts[0].ElapsedBusinessDays)
}

func TestBuildYanwenTrackingAlertsFromOfficialSnapshotsSuppressesInTransitTimeoutAfterExplicitDestinationNode(t *testing.T) {
	now := time.Date(2026, 10, 2, 9, 0, 0, 0, time.FixedZone("+08", 8*60*60))
	alerts := buildYanwenTrackingAlertsFromOfficialSnapshots([]shipping.YanwenTrackingSnapshot{{
		Environment:               "production",
		TrackingNumber:            "YW-AIR-2",
		WaybillNumber:             "YW-AIR-2",
		TrackingStatus:            "LH20",
		LatestCheckpointStatus:    "LH20",
		LatestCheckpointTimeStamp: "2026-09-21T09:00:00",
		CheckpointsData: datatypes.JSON([]byte(`[
			{"time_stamp":"2026-09-21T09:00:00","time_zone":"+08","tracking_status":"LH20","is_last_mile_checkpoint":false},
			{"time_stamp":"2026-09-22T11:00:00","time_zone":"+08","tracking_status":"LM10","is_last_mile_checkpoint":true}
		]`)),
		HasOfficialResult: true,
	}}, now.UTC())

	require.Empty(t, alerts)
}

func TestCountYanwenBusinessDaysAfterExcludesDepartureDayAndWeekends(t *testing.T) {
	location := time.FixedZone("+08", 8*60*60)
	start := time.Date(2026, 9, 21, 9, 0, 0, 0, location)
	end := time.Date(2026, 10, 2, 9, 0, 0, 0, location)
	require.Equal(t, 9, countYanwenBusinessDaysAfter(start, end))
	require.Equal(t, 0, countYanwenBusinessDaysAfter(start, start))
}
