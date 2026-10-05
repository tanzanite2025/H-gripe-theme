package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
	"gorm.io/datatypes"
)

const defaultYanwenTrackingPollingInterval = 5 * time.Minute

// YanwenTrackingOperationsService owns official production tracking queries
// and persistence of the Yanwen-only latest-result snapshots. It never writes
// generic tracking events or reads another carrier's data.
type YanwenTrackingOperationsService struct {
	configuration *YanwenGatewayConfigurationService
	snapshots     *repository.YanwenTrackingSnapshotRepository
	gateway       *YanwenGatewayClient
}

func NewYanwenTrackingOperationsService(
	configuration *YanwenGatewayConfigurationService,
	snapshots *repository.YanwenTrackingSnapshotRepository,
	gateway *YanwenGatewayClient,
) *YanwenTrackingOperationsService {
	if gateway == nil {
		gateway = NewYanwenGatewayClient()
	}
	return &YanwenTrackingOperationsService{
		configuration: configuration,
		snapshots:     snapshots,
		gateway:       gateway,
	}
}

func (s *YanwenTrackingOperationsService) QueryYanwenTracking(ctx context.Context, trackingNumbers []string) ([]YanwenTrackingResult, error) {
	if s == nil || s.configuration == nil {
		return nil, errors.New("Yanwen API configuration service is not configured")
	}
	authorization, err := s.configuration.resolveYanwenTrackingAuthorization()
	if err != nil {
		return nil, err
	}
	if s.gateway == nil {
		return nil, errors.New("Yanwen gateway client is not configured")
	}
	results, err := s.gateway.QueryYanwenTracking(ctx, authorization, trackingNumbers)
	if err != nil {
		return nil, err
	}
	if err := s.persistYanwenTrackingSnapshots(results); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *YanwenTrackingOperationsService) yanwenTrackingSnapshotRepository() *repository.YanwenTrackingSnapshotRepository {
	if s == nil {
		return nil
	}
	return s.snapshots
}

func (s *YanwenTrackingOperationsService) persistYanwenTrackingSnapshots(results []YanwenTrackingResult) error {
	if s == nil || s.snapshots == nil || len(results) == 0 {
		return nil
	}
	syncedAt := time.Now().UTC()
	snapshots := make([]*shipping.YanwenTrackingSnapshot, 0, len(results))
	for _, result := range results {
		checkpointsData, err := json.Marshal(result.Checkpoints)
		if err != nil {
			return fmt.Errorf("encode Yanwen tracking checkpoints for %s: %w", result.TrackingNumber, err)
		}
		responseData := result.RawResponse
		if len(responseData) == 0 {
			responseData, err = json.Marshal(result)
			if err != nil {
				return fmt.Errorf("encode Yanwen tracking result for %s: %w", result.TrackingNumber, err)
			}
		}
		latestCheckpointStatus := ""
		latestCheckpointTimeStamp := ""
		if len(result.Checkpoints) > 0 {
			latestCheckpoint := result.Checkpoints[len(result.Checkpoints)-1]
			latestCheckpointStatus = latestCheckpoint.TrackingStatus
			latestCheckpointTimeStamp = latestCheckpoint.TimeStamp
		}
		pollingState := shipping.YanwenTrackingPollingStateActive
		var nextSyncAt *time.Time
		if shipping.IsYanwenTrackingTerminalStatus(result.TrackingStatus) {
			pollingState = shipping.YanwenTrackingPollingStateTerminal
		} else {
			next := syncedAt.Add(defaultYanwenTrackingPollingInterval)
			nextSyncAt = &next
		}
		snapshots = append(snapshots, &shipping.YanwenTrackingSnapshot{
			Environment:               "production",
			TrackingNumber:            result.TrackingNumber,
			WaybillNumber:             result.WaybillNumber,
			ExchangeNumber:            result.ExchangeNumber,
			LastMileCarrier:           result.LastMileCarrier,
			LastMileCarrierWebsite:    result.LastMileCarrierWebsite,
			LastMileCarrierContact:    result.LastMileCarrierContact,
			TrackingStatus:            result.TrackingStatus,
			TrackingStatusLevel1:      result.TrackingStatusWaybill.Level1,
			TrackingStatusLevel2:      result.TrackingStatusWaybill.Level2,
			TrackingStatusLevel3:      result.TrackingStatusWaybill.Level3,
			LastMileTrackingExpected:  result.LastMileTrackingExpected,
			OriginCountry:             result.OriginCountry,
			DestinationCountry:        result.DestinationCountry,
			LatestCheckpointStatus:    latestCheckpointStatus,
			LatestCheckpointTimeStamp: latestCheckpointTimeStamp,
			CheckpointsData:           datatypes.JSON(checkpointsData),
			ResponseData:              datatypes.JSON(responseData),
			HasOfficialResult:         true,
			PollingState:              pollingState,
			NextSyncAt:                nextSyncAt,
			LastSyncedAt:              syncedAt,
			UpdatedAt:                 syncedAt,
		})
	}
	if err := s.snapshots.UpsertYanwenTrackingSnapshots(snapshots); err != nil {
		return fmt.Errorf("save Yanwen tracking snapshots: %w", err)
	}
	return nil
}
