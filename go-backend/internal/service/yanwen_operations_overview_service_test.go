package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/require"
)

func TestBuildYanwenOperationsOverviewFromOfficialRecordsUsesOnlyYanwenFacts(t *testing.T) {
	now := time.Date(2026, time.October, 5, 12, 0, 0, 0, time.UTC)
	lastOfficialSync := now.Add(-time.Hour)
	waybills := []shipping.YanwenWaybill{
		{
			Environment:          "production",
			DestinationCountry:   "US",
			ChannelName:          "燕文普货",
			ProductCode:          "481",
			CreatedAt:            now.Add(-2 * time.Hour),
			LastOfficialSyncedAt: &lastOfficialSync,
			IsPrinted:            false,
		},
		{
			Environment:        "production",
			DestinationCountry: "US",
			ChannelName:        "燕文普货",
			ProductCode:        "481",
			CreatedAt:          now.Add(-3 * time.Hour),
			IsPrinted:          true,
		},
		{
			Environment:        "production",
			DestinationCountry: "KR",
			ChannelName:        "燕文韩国专线",
			ProductCode:        "482",
			CreatedAt:          now.Add(-25 * time.Hour),
		},
	}
	trackingSnapshots := []shipping.YanwenTrackingSnapshot{
		{
			TrackingStatus:         "IC50",
			LatestCheckpointStatus: "PU20",
			PollingState:           shipping.YanwenTrackingPollingStateActive,
		},
		{
			TrackingStatus: "LM40",
			PollingState:   shipping.YanwenTrackingPollingStateTerminal,
		},
		{
			TrackingStatus:         "LM10",
			LatestCheckpointStatus: "IC70",
			PollingState:           shipping.YanwenTrackingPollingStateActive,
		},
	}

	overview := buildYanwenOperationsOverviewFromOfficialRecords(
		"production",
		now,
		waybills,
		trackingSnapshots,
		YanwenOperationsOverviewGatewayStatus{Environment: "production", CredentialsConfigured: true, Enabled: true},
	)

	require.Equal(t, 2, overview.TodayCreatedWaybillCount)
	require.Equal(t, 1, overview.OfficialPendingPrintWaybillCount)
	require.Equal(t, 2, overview.ActiveTrackingSnapshotCount)
	require.Equal(t, 2, overview.CustomsExceptionTrackingSnapshotCount)
	require.Equal(t, []YanwenOperationsOverviewDestinationFlow{
		{CountryCode: "US", WaybillCount: 2, ChannelName: "燕文普货"},
		{CountryCode: "KR", WaybillCount: 1, ChannelName: "燕文韩国专线"},
	}, overview.DestinationFlows)
	require.Equal(t, "production", overview.Gateway.Environment)
	require.True(t, overview.Gateway.CredentialsConfigured)
}

func TestBuildYanwenOperationsOverviewFromOfficialRecordsDoesNotInventFreightOrTrackingData(t *testing.T) {
	now := time.Date(2026, time.October, 5, 12, 0, 0, 0, time.UTC)
	overview := buildYanwenOperationsOverviewFromOfficialRecords("fat", now, nil, nil, YanwenOperationsOverviewGatewayStatus{
		Environment: "fat",
	})

	require.Equal(t, 0, overview.TodayCreatedWaybillCount)
	require.Equal(t, 0, overview.OfficialPendingPrintWaybillCount)
	require.Equal(t, 0, overview.ActiveTrackingSnapshotCount)
	require.Equal(t, 0, overview.CustomsExceptionTrackingSnapshotCount)
	require.Empty(t, overview.DestinationFlows)
	require.Equal(t, "fat", overview.Environment)
}
