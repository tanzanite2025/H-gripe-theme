package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

const (
	yanwenCustomsStagnationAlertType = "CUSTOMS_STAGNATION"
	yanwenInTransitTimeoutAlertType  = "IN_TRANSIT_TIMEOUT"
	yanwenAlertSeverityAlert         = "ALERT"
	yanwenAlertSeverityCritical      = "CRITICAL"
	yanwenCustomsStagnationThreshold = 48 * time.Hour
	yanwenInTransitTimeoutWorkdays   = 7
)

var yanwenExplicitLastMileTrackingStatuses = map[string]struct{}{
	"LM10": {},
	"LM20": {},
	"LM25": {},
	"LM40": {},
}

// YanwenTrackingAlert is an operational warning derived only from an
// official production Yanwen tracking snapshot. It is not an official Yanwen
// SLA or a customer-service case.
type YanwenTrackingAlert struct {
	TrackingNumber         string    `json:"tracking_number"`
	WaybillNumber          string    `json:"waybill_number"`
	DestinationCountry     string    `json:"destination_country"`
	AlertType              string    `json:"alert_type"`
	Severity               string    `json:"severity"`
	TrackingStatus         string    `json:"tracking_status"`
	LatestCheckpointStatus string    `json:"latest_checkpoint_status"`
	CheckpointTimeStamp    string    `json:"checkpoint_time_stamp"`
	CheckpointTimeZone     string    `json:"checkpoint_time_zone"`
	CheckpointAt           time.Time `json:"checkpoint_at"`
	ElapsedHours           int64     `json:"elapsed_hours"`
	ElapsedBusinessDays    int       `json:"elapsed_business_days"`
	SnapshotLastSyncedAt   time.Time `json:"snapshot_last_synced_at"`
	Message                string    `json:"message"`
}

// YanwenTrackingAlertService owns the read-only alert projection for the
// Yanwen tracking domain. It never reads generic tracking, order, or 4PX
// records.
type YanwenTrackingAlertService struct {
	trackingSnapshots *repository.YanwenTrackingSnapshotRepository
}

func NewYanwenTrackingAlertService(
	trackingSnapshots *repository.YanwenTrackingSnapshotRepository,
) *YanwenTrackingAlertService {
	return &YanwenTrackingAlertService{trackingSnapshots: trackingSnapshots}
}

// ListProductionYanwenTrackingAlerts evaluates only production snapshots. A
// malformed checkpoint timestamp is ignored rather than turned into a guessed
// alert, because the alert must be based on an official timestamp and offset.
func (s *YanwenTrackingAlertService) ListProductionYanwenTrackingAlerts(now time.Time) ([]YanwenTrackingAlert, error) {
	if s == nil || s.trackingSnapshots == nil {
		return nil, errors.New("Yanwen tracking snapshot repository is not configured")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	snapshots, err := s.trackingSnapshots.FindYanwenTrackingSnapshotsByEnvironment("production")
	if err != nil {
		return nil, fmt.Errorf("load Yanwen tracking alert snapshots: %w", err)
	}
	return buildYanwenTrackingAlertsFromOfficialSnapshots(snapshots, now), nil
}

type yanwenTrackingAlertCheckpoint struct {
	TimeStamp            string `json:"time_stamp"`
	TimeZone             string `json:"time_zone"`
	TrackingStatus       string `json:"tracking_status"`
	IsLastMileCheckpoint bool   `json:"is_last_mile_checkpoint"`
	At                   time.Time
	HasReliableTime      bool
}

func buildYanwenTrackingAlertsFromOfficialSnapshots(
	snapshots []shipping.YanwenTrackingSnapshot,
	now time.Time,
) []YanwenTrackingAlert {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	alerts := make([]YanwenTrackingAlert, 0)
	for _, snapshot := range snapshots {
		if strings.ToLower(strings.TrimSpace(snapshot.Environment)) != "production" || !snapshot.HasOfficialResult {
			continue
		}
		checkpoints, err := decodeYanwenTrackingAlertCheckpoints(snapshot.CheckpointsData)
		if err != nil {
			continue
		}
		currentStatus := strings.ToUpper(strings.TrimSpace(snapshot.TrackingStatus))
		latestStatus := strings.ToUpper(strings.TrimSpace(snapshot.LatestCheckpointStatus))
		if currentStatus == "" || latestStatus == "" || currentStatus != latestStatus {
			continue
		}
		currentCheckpoint, found := findYanwenCurrentAlertCheckpoint(
			checkpoints,
			latestStatus,
			snapshot.LatestCheckpointTimeStamp,
		)
		if !found || !currentCheckpoint.HasReliableTime {
			continue
		}
		if currentCheckpoint.At.After(now) {
			continue
		}

		switch currentStatus {
		case "S303", "IC50":
			elapsed := now.Sub(currentCheckpoint.At)
			if elapsed <= yanwenCustomsStagnationThreshold {
				continue
			}
			alerts = append(alerts, YanwenTrackingAlert{
				TrackingNumber:         snapshot.TrackingNumber,
				WaybillNumber:          snapshot.WaybillNumber,
				DestinationCountry:     snapshot.DestinationCountry,
				AlertType:              yanwenCustomsStagnationAlertType,
				Severity:               yanwenAlertSeverityAlert,
				TrackingStatus:         currentStatus,
				LatestCheckpointStatus: latestStatus,
				CheckpointTimeStamp:    currentCheckpoint.TimeStamp,
				CheckpointTimeZone:     currentCheckpoint.TimeZone,
				CheckpointAt:           currentCheckpoint.At.UTC(),
				ElapsedHours:           int64(elapsed / time.Hour),
				SnapshotLastSyncedAt:   snapshot.LastSyncedAt.UTC(),
				Message:                fmt.Sprintf("当前官方状态 %s 已超过 48 小时没有后续节点（已持续 %d 小时）。", currentStatus, int64(elapsed/time.Hour)),
			})
		case "LH20":
			departure, found := findLatestYanwenDepartureCheckpoint(checkpoints)
			if !found || !departure.HasReliableTime {
				continue
			}
			hasDestinationNode, reliable := yanwenHasDestinationNodeAfterDeparture(checkpoints, departure.At)
			if !reliable || hasDestinationNode {
				continue
			}
			elapsedBusinessDays := countYanwenBusinessDaysAfter(departure.At, now)
			if elapsedBusinessDays <= yanwenInTransitTimeoutWorkdays {
				continue
			}
			alerts = append(alerts, YanwenTrackingAlert{
				TrackingNumber:         snapshot.TrackingNumber,
				WaybillNumber:          snapshot.WaybillNumber,
				DestinationCountry:     snapshot.DestinationCountry,
				AlertType:              yanwenInTransitTimeoutAlertType,
				Severity:               yanwenAlertSeverityCritical,
				TrackingStatus:         currentStatus,
				LatestCheckpointStatus: latestStatus,
				CheckpointTimeStamp:    departure.TimeStamp,
				CheckpointTimeZone:     departure.TimeZone,
				CheckpointAt:           departure.At.UTC(),
				ElapsedHours:           int64(now.Sub(departure.At) / time.Hour),
				ElapsedBusinessDays:    elapsedBusinessDays,
				SnapshotLastSyncedAt:   snapshot.LastSyncedAt.UTC(),
				Message:                fmt.Sprintf("官方状态 LH20 起飞后已超过 %d 个工作日仍没有明确的目的国或尾程节点。", yanwenInTransitTimeoutWorkdays),
			})
		}
	}
	sort.Slice(alerts, func(left, right int) bool {
		leftSeverity := yanwenTrackingAlertSeverityRank(alerts[left].Severity)
		rightSeverity := yanwenTrackingAlertSeverityRank(alerts[right].Severity)
		if leftSeverity != rightSeverity {
			return leftSeverity > rightSeverity
		}
		if !alerts[left].CheckpointAt.Equal(alerts[right].CheckpointAt) {
			return alerts[left].CheckpointAt.Before(alerts[right].CheckpointAt)
		}
		return alerts[left].TrackingNumber < alerts[right].TrackingNumber
	})
	return alerts
}

func decodeYanwenTrackingAlertCheckpoints(raw []byte) ([]yanwenTrackingAlertCheckpoint, error) {
	if len(raw) == 0 || strings.EqualFold(strings.TrimSpace(string(raw)), "null") {
		return nil, errors.New("Yanwen tracking checkpoints are required")
	}
	var records []yanwenTrackingAlertCheckpoint
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, err
	}
	for index := range records {
		records[index].TimeStamp = strings.TrimSpace(records[index].TimeStamp)
		records[index].TimeZone = strings.TrimSpace(records[index].TimeZone)
		records[index].TrackingStatus = strings.ToUpper(strings.TrimSpace(records[index].TrackingStatus))
		parsed, ok := parseYanwenOfficialCheckpointTime(records[index].TimeStamp, records[index].TimeZone)
		if ok {
			records[index].At = parsed
			records[index].HasReliableTime = true
		}
	}
	return records, nil
}

func findYanwenCurrentAlertCheckpoint(
	checkpoints []yanwenTrackingAlertCheckpoint,
	status string,
	timeStamp string,
) (yanwenTrackingAlertCheckpoint, bool) {
	status = strings.ToUpper(strings.TrimSpace(status))
	timeStamp = strings.TrimSpace(timeStamp)
	for index := len(checkpoints) - 1; index >= 0; index-- {
		checkpoint := checkpoints[index]
		if checkpoint.TrackingStatus == status && checkpoint.TimeStamp == timeStamp {
			return checkpoint, true
		}
	}
	return yanwenTrackingAlertCheckpoint{}, false
}

func findLatestYanwenDepartureCheckpoint(checkpoints []yanwenTrackingAlertCheckpoint) (yanwenTrackingAlertCheckpoint, bool) {
	var departure yanwenTrackingAlertCheckpoint
	found := false
	for _, checkpoint := range checkpoints {
		if checkpoint.TrackingStatus != "LH20" {
			continue
		}
		if !checkpoint.HasReliableTime {
			return yanwenTrackingAlertCheckpoint{}, false
		}
		if !found || checkpoint.At.After(departure.At) {
			departure = checkpoint
			found = true
		}
	}
	return departure, found
}

func yanwenHasDestinationNodeAfterDeparture(
	checkpoints []yanwenTrackingAlertCheckpoint,
	departureAt time.Time,
) (hasDestinationNode bool, reliable bool) {
	for _, checkpoint := range checkpoints {
		if !isYanwenExplicitDestinationNode(checkpoint) {
			continue
		}
		if !checkpoint.HasReliableTime {
			return false, false
		}
		if !checkpoint.At.Before(departureAt) {
			return true, true
		}
	}
	return false, true
}

func isYanwenExplicitDestinationNode(checkpoint yanwenTrackingAlertCheckpoint) bool {
	if checkpoint.IsLastMileCheckpoint {
		return true
	}
	_, exists := yanwenExplicitLastMileTrackingStatuses[checkpoint.TrackingStatus]
	return exists
}

func parseYanwenOfficialCheckpointTime(timeStamp, timeZone string) (time.Time, bool) {
	timeStamp = strings.TrimSpace(timeStamp)
	if timeStamp == "" {
		return time.Time{}, false
	}
	offsetSeconds, ok := parseYanwenOfficialTimeZoneOffset(timeZone)
	if !ok {
		return time.Time{}, false
	}
	if parsed, err := time.Parse(time.RFC3339Nano, timeStamp); err == nil {
		_, parsedOffsetSeconds := parsed.Zone()
		if parsedOffsetSeconds != offsetSeconds {
			return time.Time{}, false
		}
		return parsed, true
	}
	location := time.FixedZone(timeZone, offsetSeconds)
	for _, layout := range []string{
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	} {
		if parsed, err := time.ParseInLocation(layout, timeStamp, location); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func parseYanwenOfficialTimeZoneOffset(timeZone string) (int, bool) {
	timeZone = strings.TrimSpace(strings.ToUpper(timeZone))
	if timeZone == "Z" || timeZone == "UTC" || timeZone == "GMT" {
		return 0, true
	}
	if len(timeZone) < 3 || (timeZone[0] != '+' && timeZone[0] != '-') {
		return 0, false
	}
	digits := strings.ReplaceAll(timeZone[1:], ":", "")
	if len(digits) != 2 && len(digits) != 4 {
		return 0, false
	}
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return 0, false
		}
	}
	hours := int(digits[0]-'0')*10 + int(digits[1]-'0')
	minutes := 0
	if len(digits) == 4 {
		minutes = int(digits[2]-'0')*10 + int(digits[3]-'0')
	}
	if hours > 14 || minutes > 59 || (hours == 14 && minutes != 0) {
		return 0, false
	}
	offset := hours*60*60 + minutes*60
	if timeZone[0] == '-' {
		offset = -offset
	}
	return offset, true
}

func countYanwenBusinessDaysAfter(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}
	location := start.Location()
	startLocal := start.In(location)
	endLocal := end.In(location)
	day := time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day(), 0, 0, 0, 0, location).AddDate(0, 0, 1)
	lastDay := time.Date(endLocal.Year(), endLocal.Month(), endLocal.Day(), 0, 0, 0, 0, location)
	count := 0
	for !day.After(lastDay) {
		if day.Weekday() != time.Saturday && day.Weekday() != time.Sunday {
			count++
		}
		day = day.AddDate(0, 0, 1)
	}
	return count
}

func yanwenTrackingAlertSeverityRank(severity string) int {
	if severity == yanwenAlertSeverityCritical {
		return 2
	}
	return 1
}
