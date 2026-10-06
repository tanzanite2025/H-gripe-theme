package shipping

import (
	"errors"
	"strings"
	"time"

	"gorm.io/datatypes"
)

const (
	YanwenTrackingPollingStatePending  = "pending"
	YanwenTrackingPollingStateActive   = "active"
	YanwenTrackingPollingStateTerminal = "terminal"
)

const (
	YanwenTrackingTerminalStatusDelivered = "LM40"
	YanwenTrackingTerminalStatusReturned  = "LM90"
)

// YanwenTrackingSnapshot stores the latest official response for one
// production tracking number inside the Yanwen domain. It is deliberately
// separate from the generic tracking_events table used by the storefront and
// order evidence flows.
type YanwenTrackingSnapshot struct {
	ID                        uint           `gorm:"primarykey" json:"id"`
	Environment               string         `gorm:"size:16;not null;uniqueIndex:idx_yanwen_tracking_snapshot_environment_number,priority:1;index" json:"environment"`
	TrackingNumber            string         `gorm:"size:120;not null;uniqueIndex:idx_yanwen_tracking_snapshot_environment_number,priority:2;index" json:"tracking_number"`
	WaybillNumber             string         `gorm:"size:120;not null;index" json:"waybill_number"`
	ExchangeNumber            string         `gorm:"size:120;not null;default:''" json:"exchange_number"`
	LastMileCarrier           string         `gorm:"size:160;not null;default:''" json:"last_mile_carrier"`
	LastMileCarrierWebsite    string         `gorm:"size:500;not null;default:''" json:"last_mile_carrier_website"`
	LastMileCarrierContact    string         `gorm:"size:160;not null;default:''" json:"last_mile_carrier_contact_number"`
	TrackingStatus            string         `gorm:"size:80;not null;default:''" json:"tracking_status"`
	TrackingStatusLevel1      string         `gorm:"size:40;not null;default:''" json:"tracking_status_level1"`
	TrackingStatusLevel2      string         `gorm:"size:40;not null;default:''" json:"tracking_status_level2"`
	TrackingStatusLevel3      string         `gorm:"size:80;not null;default:''" json:"tracking_status_level3"`
	LastMileTrackingExpected  bool           `gorm:"not null;default:false" json:"last_mile_tracking_expected"`
	OriginCountry             string         `gorm:"size:16;not null;default:''" json:"origin_country"`
	DestinationCountry        string         `gorm:"size:16;not null;default:''" json:"destination_country"`
	LatestCheckpointStatus    string         `gorm:"size:80;not null;default:'';index" json:"latest_checkpoint_status"`
	LatestCheckpointTimeStamp string         `gorm:"size:80;not null;default:''" json:"latest_checkpoint_time_stamp"`
	CheckpointsData           datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"checkpoints"`
	ResponseData              datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"-"`
	HasOfficialResult         bool           `gorm:"not null;default:false;index" json:"has_official_result"`
	PollingState              string         `gorm:"size:16;not null;default:'pending';index" json:"polling_state"`
	NextSyncAt                *time.Time     `gorm:"index" json:"next_sync_at,omitempty"`
	LastPollingAttemptAt      *time.Time     `gorm:"index" json:"last_polling_attempt_at,omitempty"`
	PollingLeaseToken         string         `gorm:"size:64;not null;default:''" json:"-"`
	PollingLeaseUntil         *time.Time     `gorm:"index" json:"-"`
	PollingFailureCount       int            `gorm:"not null;default:0" json:"polling_failure_count"`
	LastPollingError          string         `gorm:"size:1000;not null;default:''" json:"last_polling_error,omitempty"`
	LastSyncedAt              time.Time      `gorm:"not null;index" json:"last_synced_at"`
	CreatedAt                 time.Time      `json:"created_at"`
	UpdatedAt                 time.Time      `json:"updated_at"`
}

func (YanwenTrackingSnapshot) TableName() string { return "shipping_yanwen_tracking_snapshots" }

func (s *YanwenTrackingSnapshot) Validate() error {
	if s == nil {
		return errors.New("Yanwen tracking snapshot is required")
	}
	s.Environment = strings.ToLower(strings.TrimSpace(s.Environment))
	s.TrackingNumber = strings.TrimSpace(s.TrackingNumber)
	s.WaybillNumber = strings.TrimSpace(s.WaybillNumber)
	s.ExchangeNumber = strings.TrimSpace(s.ExchangeNumber)
	s.LastMileCarrier = strings.TrimSpace(s.LastMileCarrier)
	s.LastMileCarrierWebsite = strings.TrimSpace(s.LastMileCarrierWebsite)
	s.LastMileCarrierContact = strings.TrimSpace(s.LastMileCarrierContact)
	s.TrackingStatus = strings.TrimSpace(s.TrackingStatus)
	s.TrackingStatusLevel1 = strings.TrimSpace(s.TrackingStatusLevel1)
	s.TrackingStatusLevel2 = strings.TrimSpace(s.TrackingStatusLevel2)
	s.TrackingStatusLevel3 = strings.TrimSpace(s.TrackingStatusLevel3)
	s.PollingState = strings.ToLower(strings.TrimSpace(s.PollingState))
	s.PollingLeaseToken = strings.TrimSpace(s.PollingLeaseToken)
	s.LastPollingError = strings.TrimSpace(s.LastPollingError)
	s.OriginCountry = strings.ToUpper(strings.TrimSpace(s.OriginCountry))
	s.DestinationCountry = strings.ToUpper(strings.TrimSpace(s.DestinationCountry))
	s.LatestCheckpointStatus = strings.TrimSpace(s.LatestCheckpointStatus)
	s.LatestCheckpointTimeStamp = strings.TrimSpace(s.LatestCheckpointTimeStamp)
	if s.Environment != "production" {
		return errors.New("Yanwen tracking snapshot environment must be production")
	}
	if s.TrackingNumber == "" || s.WaybillNumber == "" {
		return errors.New("Yanwen tracking snapshot numbers are required")
	}
	if s.PollingState == "" {
		if s.HasOfficialResult || s.TrackingStatus != "" || !s.LastSyncedAt.IsZero() {
			s.PollingState = YanwenTrackingPollingStateActive
		} else {
			s.PollingState = YanwenTrackingPollingStatePending
		}
	}
	if s.PollingState != YanwenTrackingPollingStatePending &&
		s.PollingState != YanwenTrackingPollingStateActive &&
		s.PollingState != YanwenTrackingPollingStateTerminal {
		return errors.New("Yanwen tracking snapshot polling state is invalid")
	}
	if !s.HasOfficialResult && (s.TrackingStatus != "" || !s.LastSyncedAt.IsZero()) {
		s.HasOfficialResult = true
	}
	if s.HasOfficialResult && s.TrackingStatus == "" {
		return errors.New("Yanwen tracking snapshot status is required for an official result")
	}
	if s.HasOfficialResult && s.LastSyncedAt.IsZero() {
		return errors.New("Yanwen tracking snapshot sync time is required for an official result")
	}
	if s.PollingFailureCount < 0 {
		return errors.New("Yanwen tracking snapshot polling failure count cannot be negative")
	}
	if len(s.CheckpointsData) == 0 {
		s.CheckpointsData = datatypes.JSON([]byte("[]"))
	}
	if len(s.ResponseData) == 0 {
		s.ResponseData = datatypes.JSON([]byte("{}"))
	}
	return nil
}

// IsYanwenTrackingTerminalStatus reports only statuses that the documented
// tracking UI treats as permanently complete. Transient exception states stay
// active so a later official response can move the parcel forward.
func IsYanwenTrackingTerminalStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case YanwenTrackingTerminalStatusDelivered, YanwenTrackingTerminalStatusReturned:
		return true
	default:
		return false
	}
}
