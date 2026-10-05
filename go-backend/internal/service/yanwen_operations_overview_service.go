package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

const maximumYanwenOperationsOverviewDestinationFlows = 5

var yanwenOperationsOverviewCustomsExceptionStatuses = map[string]struct{}{
	"S303": {},
	"IC50": {},
	"IC51": {},
	"IC70": {},
}

// YanwenOperationsOverview is the Yanwen-only read model used by the
// operations dashboard. It is assembled from the Yanwen waybill ledger and
// Yanwen tracking snapshots; it never reads generic shipping or tracking data.
type YanwenOperationsOverview struct {
	GeneratedAt                           time.Time                                 `json:"generated_at"`
	Environment                           string                                    `json:"environment"`
	TodayCreatedWaybillCount              int                                       `json:"today_created_waybill_count"`
	OfficialPendingPrintWaybillCount      int                                       `json:"official_pending_print_waybill_count"`
	ActiveTrackingSnapshotCount           int                                       `json:"active_tracking_snapshot_count"`
	CustomsExceptionTrackingSnapshotCount int                                       `json:"customs_exception_tracking_snapshot_count"`
	DestinationFlows                      []YanwenOperationsOverviewDestinationFlow `json:"destination_flows"`
	Gateway                               YanwenOperationsOverviewGatewayStatus     `json:"gateway"`
}

// YanwenOperationsOverviewDestinationFlow summarizes only persisted Yanwen
// waybills. ChannelName is the most frequent channel for the destination in
// the selected Yanwen environment.
type YanwenOperationsOverviewDestinationFlow struct {
	CountryCode  string `json:"country_code"`
	WaybillCount int    `json:"waybill_count"`
	ChannelName  string `json:"channel_name"`
}

// YanwenOperationsOverviewGatewayStatus exposes non-secret gateway state.
type YanwenOperationsOverviewGatewayStatus struct {
	Environment           string `json:"environment"`
	Endpoint              string `json:"endpoint"`
	CredentialsConfigured bool   `json:"credentials_configured"`
	Enabled               bool   `json:"enabled"`
}

// YanwenOperationsOverviewService owns the read-only dashboard projection.
// It reads only Yanwen's local waybill, tracking snapshot, and configuration
// repositories and never initiates a hidden external request.
type YanwenOperationsOverviewService struct {
	configuration     *YanwenGatewayConfigurationService
	waybills          *repository.YanwenWaybillRepository
	trackingSnapshots *repository.YanwenTrackingSnapshotRepository
}

func NewYanwenOperationsOverviewService(
	configuration *YanwenGatewayConfigurationService,
	waybills *repository.YanwenWaybillRepository,
	trackingSnapshots *repository.YanwenTrackingSnapshotRepository,
) *YanwenOperationsOverviewService {
	return &YanwenOperationsOverviewService{
		configuration:     configuration,
		waybills:          waybills,
		trackingSnapshots: trackingSnapshots,
	}
}

// GetYanwenOperationsOverview reads the selected Yanwen environment's local
// ledger and official tracking snapshots for the operations dashboard.
func (s *YanwenOperationsOverviewService) GetYanwenOperationsOverview(environment string) (*YanwenOperationsOverview, error) {
	environment = normalizeYanwenEnvironment(environment)
	if !isSupportedYanwenEnvironment(environment) {
		return nil, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.waybills == nil {
		return nil, errors.New("Yanwen waybill repository is not configured")
	}
	waybills, err := s.waybills.FindYanwenWaybills(environment, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("load Yanwen waybill overview records: %w", err)
	}
	trackingSnapshots := make([]shipping.YanwenTrackingSnapshot, 0)
	if s.trackingSnapshots != nil {
		trackingSnapshots, err = s.trackingSnapshots.FindYanwenTrackingSnapshotsByEnvironment(environment)
		if err != nil {
			return nil, fmt.Errorf("load Yanwen tracking overview records: %w", err)
		}
	}

	gatewayStatus := YanwenOperationsOverviewGatewayStatus{
		Environment: environment,
		Endpoint:    yanwenEndpointForEnvironment(environment),
	}
	if s.configuration != nil {
		view, viewErr := s.configuration.GetYanwenAPIConfigurationView(environment)
		if viewErr != nil {
			return nil, fmt.Errorf("load Yanwen gateway overview state: %w", viewErr)
		}
		gatewayStatus = YanwenOperationsOverviewGatewayStatus{
			Environment:           view.Environment,
			Endpoint:              view.Endpoint,
			CredentialsConfigured: view.UserIDConfigured && view.APITokenConfigured,
			Enabled:               view.Enabled,
		}
	}

	return buildYanwenOperationsOverviewFromOfficialRecords(
		environment,
		time.Now().UTC(),
		waybills,
		trackingSnapshots,
		gatewayStatus,
	), nil
}

func buildYanwenOperationsOverviewFromOfficialRecords(
	environment string,
	generatedAt time.Time,
	waybills []shipping.YanwenWaybill,
	trackingSnapshots []shipping.YanwenTrackingSnapshot,
	gatewayStatus YanwenOperationsOverviewGatewayStatus,
) *YanwenOperationsOverview {
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}
	generatedAt = generatedAt.UTC()
	startOfGeneratedDay := time.Date(
		generatedAt.Year(),
		generatedAt.Month(),
		generatedAt.Day(),
		0,
		0,
		0,
		0,
		generatedAt.Location(),
	)

	overview := &YanwenOperationsOverview{
		GeneratedAt:      generatedAt,
		Environment:      environment,
		DestinationFlows: make([]YanwenOperationsOverviewDestinationFlow, 0),
		Gateway:          gatewayStatus,
	}
	destinationCounts := make(map[string]*yanwenOverviewDestinationAccumulator)
	for _, waybill := range waybills {
		if !waybill.CreatedAt.IsZero() && !waybill.CreatedAt.Before(startOfGeneratedDay) && waybill.CreatedAt.Before(generatedAt) {
			overview.TodayCreatedWaybillCount++
		}
		if waybill.LastOfficialSyncedAt != nil && !waybill.IsPrinted {
			overview.OfficialPendingPrintWaybillCount++
		}
		countryCode := strings.ToUpper(strings.TrimSpace(waybill.DestinationCountry))
		if countryCode == "" {
			continue
		}
		accumulator := destinationCounts[countryCode]
		if accumulator == nil {
			accumulator = &yanwenOverviewDestinationAccumulator{ChannelCounts: make(map[string]int)}
			destinationCounts[countryCode] = accumulator
		}
		accumulator.WaybillCount++
		channelName := strings.TrimSpace(waybill.ChannelName)
		if channelName == "" {
			channelName = strings.TrimSpace(waybill.ProductCode)
		}
		if channelName != "" {
			accumulator.ChannelCounts[channelName]++
		}
	}
	for _, snapshot := range trackingSnapshots {
		if snapshot.PollingState != shipping.YanwenTrackingPollingStateTerminal && !shipping.IsYanwenTrackingTerminalStatus(snapshot.TrackingStatus) {
			overview.ActiveTrackingSnapshotCount++
		}
		if yanwenTrackingSnapshotContainsCustomsExceptionStatus(snapshot) {
			overview.CustomsExceptionTrackingSnapshotCount++
		}
	}

	type destinationFlowWithSortKey struct {
		flow YanwenOperationsOverviewDestinationFlow
	}
	flows := make([]destinationFlowWithSortKey, 0, len(destinationCounts))
	for countryCode, accumulator := range destinationCounts {
		flows = append(flows, destinationFlowWithSortKey{flow: YanwenOperationsOverviewDestinationFlow{
			CountryCode:  countryCode,
			WaybillCount: accumulator.WaybillCount,
			ChannelName:  mostFrequentYanwenOverviewChannel(accumulator.ChannelCounts),
		}})
	}
	sort.Slice(flows, func(left, right int) bool {
		if flows[left].flow.WaybillCount != flows[right].flow.WaybillCount {
			return flows[left].flow.WaybillCount > flows[right].flow.WaybillCount
		}
		return flows[left].flow.CountryCode < flows[right].flow.CountryCode
	})
	if len(flows) > maximumYanwenOperationsOverviewDestinationFlows {
		flows = flows[:maximumYanwenOperationsOverviewDestinationFlows]
	}
	for _, flow := range flows {
		overview.DestinationFlows = append(overview.DestinationFlows, flow.flow)
	}
	return overview
}

type yanwenOverviewDestinationAccumulator struct {
	WaybillCount  int
	ChannelCounts map[string]int
}

func mostFrequentYanwenOverviewChannel(channelCounts map[string]int) string {
	bestChannel := ""
	bestCount := 0
	for channelName, count := range channelCounts {
		if count > bestCount || (count == bestCount && (bestChannel == "" || channelName < bestChannel)) {
			bestChannel = channelName
			bestCount = count
		}
	}
	return bestChannel
}

func yanwenTrackingSnapshotContainsCustomsExceptionStatus(snapshot shipping.YanwenTrackingSnapshot) bool {
	for _, candidate := range []string{snapshot.TrackingStatus, snapshot.LatestCheckpointStatus} {
		if _, exists := yanwenOperationsOverviewCustomsExceptionStatuses[strings.ToUpper(strings.TrimSpace(candidate))]; exists {
			return true
		}
	}
	return false
}
