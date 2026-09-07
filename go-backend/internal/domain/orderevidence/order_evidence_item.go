package orderevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	EvidenceItemTypeConfigurationConfirmation = "configuration_confirmation"
	EvidenceItemTypeProductIdentity           = "product_identity"
	EvidenceItemTypeOutboundWeightPackaging   = "outbound_weight_packaging"
	EvidenceItemTypeSignedPOD                 = "signed_pod"
	EvidenceItemTypeSpokeQCTension            = "spoke_qc_tension"

	EvidenceItemStatusMissing  = "missing"
	EvidenceItemStatusDraft    = "draft"
	EvidenceItemStatusComplete = "complete"
	EvidenceItemStatusWaived   = "waived"

	EvidenceRequiredReasonBase           = "base"
	EvidenceRequiredReasonHighValue      = "high_value"
	EvidenceRequiredReasonSpokeTensionQC = "spoke_tension_qc"
)

// OrderEvidenceItem is a mutable structured evidence record inside a package.
// Its attachment references are stored separately.
type OrderEvidenceItem struct {
	ID             uint                      `gorm:"primaryKey" json:"id"`
	PackageID      uint                      `gorm:"not null;index" json:"package_id"`
	OrderID        uint                      `gorm:"not null;index" json:"order_id"`
	OrderItemID    *uint                     `gorm:"index" json:"order_item_id,omitempty"`
	SnapshotID     *uint                     `gorm:"index" json:"snapshot_id,omitempty"`
	ItemType       string                    `gorm:"size:64;not null" json:"item_type"`
	Status         string                    `gorm:"size:16;not null;index" json:"status"`
	RequiredReason string                    `gorm:"size:32;not null" json:"required_reason"`
	DataJSON       datatypes.JSON            `gorm:"column:data_json;type:jsonb;not null;default:'{}'" json:"data_json"`
	CapturedAt     *time.Time                `json:"captured_at,omitempty"`
	CapturedBy     uint                      `gorm:"not null;default:0" json:"captured_by"`
	ContentSHA256  string                    `gorm:"column:content_sha256;type:char(64);not null;default:''" json:"content_sha256"`
	Attachments    []OrderEvidenceAttachment `gorm:"foreignKey:EvidenceItemID" json:"attachments,omitempty"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

func (OrderEvidenceItem) TableName() string {
	return "order_evidence_items"
}

func (i OrderEvidenceItem) Validate() error {
	if i.PackageID == 0 {
		return errors.New("order evidence item package_id is required")
	}
	if i.OrderID == 0 {
		return errors.New("order evidence item order_id is required")
	}
	itemType := strings.ToLower(strings.TrimSpace(i.ItemType))
	if !isValidEvidenceItemType(itemType) {
		return fmt.Errorf("unsupported order evidence item type %q", i.ItemType)
	}
	status := strings.ToLower(strings.TrimSpace(i.Status))
	switch status {
	case EvidenceItemStatusMissing, EvidenceItemStatusDraft, EvidenceItemStatusComplete, EvidenceItemStatusWaived:
	default:
		return fmt.Errorf("unsupported order evidence item status %q", i.Status)
	}
	requiredReason := strings.ToLower(strings.TrimSpace(i.RequiredReason))
	switch requiredReason {
	case EvidenceRequiredReasonBase, EvidenceRequiredReasonHighValue, EvidenceRequiredReasonSpokeTensionQC:
	default:
		return fmt.Errorf("unsupported order evidence item required reason %q", i.RequiredReason)
	}
	if evidenceItemRequiresOrderItem(itemType) && (i.OrderItemID == nil || *i.OrderItemID == 0) {
		return fmt.Errorf("order evidence item type %q requires order_item_id", i.ItemType)
	}
	if !evidenceItemRequiresOrderItem(itemType) && i.OrderItemID != nil {
		return fmt.Errorf("order evidence item type %q cannot have order_item_id", i.ItemType)
	}
	if itemType == EvidenceItemTypeConfigurationConfirmation && (i.SnapshotID == nil || *i.SnapshotID == 0) {
		return errors.New("configuration confirmation evidence item requires snapshot_id")
	}
	if i.SnapshotID != nil && (itemType != EvidenceItemTypeConfigurationConfirmation || status != EvidenceItemStatusComplete) {
		return errors.New("snapshot-backed order evidence items must be complete configuration confirmations")
	}
	if len(i.DataJSON) == 0 {
		i.DataJSON = datatypes.JSON([]byte("{}"))
	}
	if !json.Valid(i.DataJSON) {
		return errors.New("order evidence item data_json must be valid JSON")
	}
	if status == EvidenceItemStatusComplete {
		if len(strings.TrimSpace(i.ContentSHA256)) != 64 {
			return errors.New("complete order evidence item content_sha256 is required")
		}
		if i.CapturedAt == nil || i.CapturedAt.IsZero() {
			return errors.New("complete order evidence item captured_at is required")
		}
		hash := sha256.Sum256(i.DataJSON)
		if !strings.EqualFold(strings.TrimSpace(i.ContentSHA256), hex.EncodeToString(hash[:])) {
			return errors.New("complete order evidence item content_sha256 does not match data_json")
		}
	}
	return nil
}

func (i *OrderEvidenceItem) BeforeCreate(tx *gorm.DB) error {
	if i == nil {
		return errors.New("order evidence item is required")
	}
	i.normalize()
	return i.Validate()
}

func (i *OrderEvidenceItem) BeforeSave(tx *gorm.DB) error {
	if i == nil {
		return errors.New("order evidence item is required")
	}
	i.normalize()
	return i.Validate()
}

func (i *OrderEvidenceItem) normalize() {
	if i == nil {
		return
	}
	i.ItemType = strings.ToLower(strings.TrimSpace(i.ItemType))
	i.Status = strings.ToLower(strings.TrimSpace(i.Status))
	i.RequiredReason = strings.ToLower(strings.TrimSpace(i.RequiredReason))
	if len(i.DataJSON) == 0 {
		i.DataJSON = datatypes.JSON([]byte("{}"))
	}
}

func isValidEvidenceItemType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case EvidenceItemTypeConfigurationConfirmation,
		EvidenceItemTypeProductIdentity,
		EvidenceItemTypeOutboundWeightPackaging,
		EvidenceItemTypeSignedPOD,
		EvidenceItemTypeSpokeQCTension:
		return true
	default:
		return false
	}
}

func evidenceItemRequiresOrderItem(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case EvidenceItemTypeConfigurationConfirmation,
		EvidenceItemTypeProductIdentity,
		EvidenceItemTypeSpokeQCTension:
		return true
	default:
		return false
	}
}
