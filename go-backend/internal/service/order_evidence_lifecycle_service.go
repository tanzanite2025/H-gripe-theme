package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

var (
	ErrOrderEvidencePackageNotFound        = errors.New("order evidence package not found")
	ErrOrderEvidenceItemNotFound           = errors.New("order evidence item not found")
	ErrOrderEvidenceItemInvalid            = errors.New("order evidence item is invalid")
	ErrOrderEvidenceConfigurationImmutable = errors.New("configuration confirmation evidence item is immutable")
	ErrOrderEvidenceAttachmentRequired     = errors.New("order evidence item requires at least one attachment before completion")
)

type OrderEvidenceItemUpdateInput struct {
	OrderID    uint
	ItemID     uint
	Status     string
	DataJSON   datatypes.JSON
	CapturedAt *time.Time
	CapturedBy uint
}

type OrderEvidenceItemUpdateResult struct {
	Item         *orderevidence.OrderEvidenceItem
	Package      *orderevidence.OrderEvidencePackage
	Completeness orderevidence.EvidenceCompleteness
}

func (s *OrderEvidenceService) GetPackageCompleteness(
	repos repository.TxRepositories,
	packageID uint,
) (orderevidence.EvidenceCompleteness, error) {
	if s == nil || repos.OrderEvidence == nil {
		return orderevidence.EvidenceCompleteness{}, ErrOrderEvidenceStoreUnavailable
	}
	if packageID == 0 {
		return orderevidence.EvidenceCompleteness{}, errors.New("order evidence package id is required")
	}
	if _, err := repos.OrderEvidence.FindPackageByID(packageID); repository.IsRecordNotFound(err) {
		return orderevidence.EvidenceCompleteness{}, ErrOrderEvidencePackageNotFound
	} else if err != nil {
		return orderevidence.EvidenceCompleteness{}, err
	}
	items, err := repos.OrderEvidence.ListItemsByPackageID(packageID)
	if err != nil {
		return orderevidence.EvidenceCompleteness{}, err
	}
	return orderevidence.EvaluateEvidenceCompleteness(items), nil
}

func (s *OrderEvidenceService) UpdateItem(
	repos repository.TxRepositories,
	input OrderEvidenceItemUpdateInput,
) (*OrderEvidenceItemUpdateResult, error) {
	if s == nil || repos.OrderEvidence == nil {
		return nil, ErrOrderEvidenceStoreUnavailable
	}
	if input.ItemID == 0 {
		return nil, fmt.Errorf("%w: order evidence item id is required", ErrOrderEvidenceItemInvalid)
	}

	var (
		item *orderevidence.OrderEvidenceItem
		err  error
	)
	if input.OrderID != 0 {
		item, _, err = repos.OrderEvidence.FindItemAndPackageByIDForOrder(input.ItemID, input.OrderID)
	} else {
		item, err = repos.OrderEvidence.FindItemByID(input.ItemID)
	}
	if repository.IsRecordNotFound(err) {
		return nil, ErrOrderEvidenceItemNotFound
	}
	if err != nil {
		return nil, err
	}
	pkg, err := repos.OrderEvidence.FindPackageByID(item.PackageID)
	if repository.IsRecordNotFound(err) {
		return nil, ErrOrderEvidencePackageNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := pkg.EnsureMutable(); err != nil {
		return nil, err
	}
	if item.ItemType == orderevidence.EvidenceItemTypeConfigurationConfirmation {
		return nil, ErrOrderEvidenceConfigurationImmutable
	}

	updated := *item
	if input.Status != "" {
		updated.Status = input.Status
	}
	if len(input.DataJSON) > 0 {
		updated.DataJSON = append(datatypes.JSON(nil), input.DataJSON...)
	}
	if input.CapturedAt != nil {
		capturedAt := input.CapturedAt.UTC()
		updated.CapturedAt = &capturedAt
	}
	if input.CapturedBy != 0 {
		updated.CapturedBy = input.CapturedBy
	}
	updated.Status = normalizeEvidenceStatus(updated.Status)
	if len(updated.DataJSON) == 0 {
		updated.DataJSON = datatypes.JSON([]byte("{}"))
	}
	if !json.Valid(updated.DataJSON) {
		return nil, fmt.Errorf("%w: order evidence item data_json must be valid JSON", ErrOrderEvidenceItemInvalid)
	}

	switch updated.Status {
	case orderevidence.EvidenceItemStatusComplete:
		if updated.CapturedAt == nil || updated.CapturedAt.IsZero() {
			now := time.Now().UTC()
			updated.CapturedAt = &now
		}
		if updated.CapturedBy == 0 {
			return nil, fmt.Errorf("%w: complete order evidence item captured_by is required", ErrOrderEvidenceItemInvalid)
		}
		if evidenceItemRequiresAttachment(updated.ItemType) {
			attachments, err := repos.OrderEvidence.ListAttachmentsByItemID(updated.ID)
			if err != nil {
				return nil, err
			}
			if len(attachments) == 0 {
				return nil, fmt.Errorf(
					"%w: %s",
					ErrOrderEvidenceAttachmentRequired,
					updated.ItemType,
				)
			}
		}
		if err := validateCompletedEvidenceRecord(updated.ItemType, updated.DataJSON); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceItemInvalid, err)
		}
		updated.ContentSHA256 = hashEvidenceData(updated.DataJSON)
	case orderevidence.EvidenceItemStatusMissing:
		updated.CapturedAt = nil
		updated.CapturedBy = 0
		updated.ContentSHA256 = ""
		updated.DataJSON = datatypes.JSON([]byte("{}"))
	case orderevidence.EvidenceItemStatusDraft:
		if updated.CapturedAt == nil || updated.CapturedAt.IsZero() {
			now := time.Now().UTC()
			updated.CapturedAt = &now
		}
		if updated.CapturedBy == 0 {
			return nil, fmt.Errorf("%w: draft order evidence item captured_by is required", ErrOrderEvidenceItemInvalid)
		}
		updated.ContentSHA256 = hashEvidenceData(updated.DataJSON)
	case orderevidence.EvidenceItemStatusWaived:
		if updated.CapturedAt == nil || updated.CapturedAt.IsZero() {
			now := time.Now().UTC()
			updated.CapturedAt = &now
		}
		if updated.CapturedBy == 0 {
			return nil, fmt.Errorf("%w: waived order evidence item captured_by is required", ErrOrderEvidenceItemInvalid)
		}
		if err := validateEvidenceWaiverReason(updated.DataJSON); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceItemInvalid, err)
		}
		updated.ContentSHA256 = hashEvidenceData(updated.DataJSON)
	default:
		return nil, fmt.Errorf("%w: unsupported order evidence item status %q", ErrOrderEvidenceItemInvalid, updated.Status)
	}
	if err := updated.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceItemInvalid, err)
	}

	updatedPackage, completeness, err := repos.OrderEvidence.UpdateItemAndRefreshPackage(&updated)
	if err != nil {
		return nil, err
	}
	return &OrderEvidenceItemUpdateResult{
		Item:         &updated,
		Package:      updatedPackage,
		Completeness: completeness,
	}, nil
}

func (s *OrderEvidenceService) LockPackage(
	repos repository.TxRepositories,
	packageID uint,
	lockedAt time.Time,
) (*orderevidence.OrderEvidencePackage, orderevidence.EvidenceCompleteness, error) {
	if s == nil || repos.OrderEvidence == nil {
		return nil, orderevidence.EvidenceCompleteness{}, ErrOrderEvidenceStoreUnavailable
	}
	if packageID == 0 {
		return nil, orderevidence.EvidenceCompleteness{}, errors.New("order evidence package id is required")
	}
	if lockedAt.IsZero() {
		lockedAt = time.Now().UTC()
	}
	pkg, completeness, err := repos.OrderEvidence.LockPackage(packageID, lockedAt.UTC())
	if repository.IsRecordNotFound(err) {
		return nil, orderevidence.EvidenceCompleteness{}, ErrOrderEvidencePackageNotFound
	}
	if err != nil {
		return nil, orderevidence.EvidenceCompleteness{}, err
	}
	return pkg, completeness, nil
}

func (s *OrderEvidenceService) CreateRevision(
	repos repository.TxRepositories,
	packageID uint,
	createdBy uint,
) (*orderevidence.OrderEvidencePackage, error) {
	if s == nil || repos.OrderEvidence == nil {
		return nil, ErrOrderEvidenceStoreUnavailable
	}
	if packageID == 0 {
		return nil, errors.New("order evidence package id is required")
	}
	if _, err := repos.OrderEvidence.FindPackageByID(packageID); repository.IsRecordNotFound(err) {
		return nil, ErrOrderEvidencePackageNotFound
	} else if err != nil {
		return nil, err
	}
	return repos.OrderEvidence.CreateRevision(packageID, createdBy)
}

func normalizeEvidenceStatus(value string) string {
	switch value {
	case "":
		return orderevidence.EvidenceItemStatusMissing
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func hashEvidenceData(data datatypes.JSON) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func validateEvidenceWaiverReason(data datatypes.JSON) error {
	var payload struct {
		WaiverReason string `json:"waiver_reason"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return errors.New("waived order evidence item data_json must be a JSON object")
	}
	if strings.TrimSpace(payload.WaiverReason) == "" {
		return errors.New("waived order evidence item waiver_reason is required")
	}
	return nil
}

func evidenceItemRequiresAttachment(itemType string) bool {
	switch itemType {
	case orderevidence.EvidenceItemTypeProductIdentity,
		orderevidence.EvidenceItemTypeOutboundWeightPackaging,
		orderevidence.EvidenceItemTypeSignedPOD,
		orderevidence.EvidenceItemTypeSpokeQCTension:
		return true
	default:
		return false
	}
}

func validateCompletedEvidenceRecord(itemType string, data datatypes.JSON) error {
	switch itemType {
	case orderevidence.EvidenceItemTypeOutboundWeightPackaging:
		_, err := orderevidence.ParseOutboundWeightPackagingRecord(data)
		return err
	case orderevidence.EvidenceItemTypeSignedPOD:
		_, err := orderevidence.ParseSignedPODRecord(data)
		return err
	default:
		return nil
	}
}
