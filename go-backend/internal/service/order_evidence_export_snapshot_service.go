package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

var (
	ErrOrderEvidenceExportSnapshotUnavailable = errors.New("order evidence export snapshot service is unavailable")
	ErrOrderEvidenceExportPackageMissing      = errors.New("order evidence export package is missing")
	ErrOrderEvidenceExportRequiresLocked      = errors.New("order evidence export requires a locked evidence package")
	ErrOrderEvidenceExportSnapshotInvalid     = errors.New("order evidence export snapshot is invalid")
)

// OrderEvidenceExportSnapshotPayload is the safe, immutable manifest returned
// by the order evidence export endpoint. Storage keys are intentionally not
// included; attachment IDs, metadata, and hashes remain available for the
// authenticated attachment endpoint.
type OrderEvidenceExportSnapshotPayload struct {
	SchemaVersion   int                                `json:"schema_version"`
	ExportType      string                             `json:"export_type"`
	ExportedAt      time.Time                          `json:"exported_at"`
	Order           *order.Order                       `json:"order"`
	Package         *OrderEvidenceExportPackage        `json:"package"`
	Completeness    orderevidence.EvidenceCompleteness `json:"completeness"`
	TrackingContext *OrderEvidenceTrackingContext      `json:"tracking_context,omitempty"`
	TrackingEvents  []shipping.TrackingEvent           `json:"tracking_events"`
	Sources         []OrderEvidenceSourceReference     `json:"sources"`
	Warnings        []string                           `json:"warnings,omitempty"`
}

type OrderEvidenceExportPackage struct {
	ID                    uint                      `json:"id"`
	OrderID               uint                      `json:"order_id"`
	SnapshotID            uint                      `json:"snapshot_id"`
	PackageVersion        int                       `json:"package_version"`
	Status                string                    `json:"status"`
	OrderTotalUSDSnapshot float64                   `json:"order_total_usd_snapshot"`
	IsHighValue           bool                      `json:"is_high_value"`
	HasSpokeTensionQC     bool                      `json:"has_spoke_tension_qc"`
	SchemaVersion         int                       `json:"schema_version"`
	CreatedBy             uint                      `json:"created_by"`
	LockedAt              *time.Time                `json:"locked_at,omitempty"`
	CreatedAt             time.Time                 `json:"created_at"`
	UpdatedAt             time.Time                 `json:"updated_at"`
	Items                 []OrderEvidenceExportItem `json:"items"`
}

type OrderEvidenceExportItem struct {
	ID             uint                            `json:"id"`
	PackageID      uint                            `json:"package_id"`
	OrderID        uint                            `json:"order_id"`
	OrderItemID    *uint                           `json:"order_item_id,omitempty"`
	SnapshotID     *uint                           `json:"snapshot_id,omitempty"`
	ItemType       string                          `json:"item_type"`
	Status         string                          `json:"status"`
	RequiredReason string                          `json:"required_reason"`
	DataJSON       json.RawMessage                 `json:"data_json"`
	CapturedAt     *time.Time                      `json:"captured_at,omitempty"`
	CapturedBy     uint                            `json:"captured_by"`
	ContentSHA256  string                          `json:"content_sha256"`
	Attachments    []OrderEvidenceExportAttachment `json:"attachments"`
	CreatedAt      time.Time                       `json:"created_at"`
	UpdatedAt      time.Time                       `json:"updated_at"`
}

type OrderEvidenceExportAttachment struct {
	ID               uint      `json:"id"`
	EvidenceItemID   uint      `json:"evidence_item_id"`
	OriginalFilename string    `json:"original_filename"`
	MimeType         string    `json:"mime_type"`
	SizeBytes        int64     `json:"size_bytes"`
	SHA256           string    `json:"sha256"`
	UploadedBy       uint      `json:"uploaded_by"`
	CreatedAt        time.Time `json:"created_at"`
}

// OrderEvidenceExportSnapshotService creates or reuses an immutable manifest
// for the current locked package version. It never mutates evidence facts.
type OrderEvidenceExportSnapshotService struct {
	repo         *repository.OrderEvidenceExportSnapshotRepository
	evidenceRepo *repository.OrderEvidenceRepository
	assembler    *OrderEvidencePackageAssembler
}

func NewOrderEvidenceExportSnapshotService(
	repo *repository.OrderEvidenceExportSnapshotRepository,
	evidenceRepo *repository.OrderEvidenceRepository,
	assembler *OrderEvidencePackageAssembler,
) *OrderEvidenceExportSnapshotService {
	return &OrderEvidenceExportSnapshotService{
		repo:         repo,
		evidenceRepo: evidenceRepo,
		assembler:    assembler,
	}
}

func (s *OrderEvidenceExportSnapshotService) Export(
	orderID uint,
	createdBy uint,
) (*orderevidence.OrderEvidenceExportSnapshot, error) {
	if s == nil || s.repo == nil || s.evidenceRepo == nil || s.assembler == nil {
		return nil, ErrOrderEvidenceExportSnapshotUnavailable
	}
	if orderID == 0 {
		return nil, errors.New("order id is required")
	}

	currentPackage, err := s.evidenceRepo.FindLatestPackageByOrderID(orderID)
	if repository.IsRecordNotFound(err) {
		return nil, ErrOrderEvidenceExportPackageMissing
	}
	if err != nil {
		return nil, err
	}
	if !currentPackage.IsLocked() {
		return nil, ErrOrderEvidenceExportRequiresLocked
	}

	if existing, err := s.repo.FindByPackageID(currentPackage.ID); err == nil {
		if existing.OrderID != orderID ||
			existing.EvidencePackageVersion != currentPackage.PackageVersion {
			return nil, fmt.Errorf("%w: stored package scope does not match current package", ErrOrderEvidenceExportSnapshotInvalid)
		}
		if err := existing.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceExportSnapshotInvalid, err)
		}
		return existing, nil
	} else if !repository.IsRecordNotFound(err) {
		return nil, err
	}

	assembly, err := s.assembler.Assemble(orderID)
	if err != nil {
		return nil, err
	}
	if assembly == nil || assembly.Package == nil {
		return nil, ErrOrderEvidenceExportPackageMissing
	}
	if !assembly.Package.IsLocked() {
		return nil, ErrOrderEvidenceExportRequiresLocked
	}

	if existing, err := s.repo.FindByPackageID(assembly.Package.ID); err == nil {
		if existing.OrderID != orderID ||
			existing.EvidencePackageVersion != assembly.Package.PackageVersion {
			return nil, fmt.Errorf("%w: stored package scope does not match assembled package", ErrOrderEvidenceExportSnapshotInvalid)
		}
		if err := existing.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceExportSnapshotInvalid, err)
		}
		return existing, nil
	} else if !repository.IsRecordNotFound(err) {
		return nil, err
	}

	now := time.Now().UTC()
	payload := buildOrderEvidenceExportSnapshotPayload(assembly, now)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: marshal payload: %v", ErrOrderEvidenceExportSnapshotInvalid, err)
	}
	hash := sha256.Sum256(payloadBytes)
	snapshot := &orderevidence.OrderEvidenceExportSnapshot{
		OrderID:                orderID,
		EvidencePackageID:      assembly.Package.ID,
		EvidencePackageVersion: assembly.Package.PackageVersion,
		Version:                1,
		Status:                 orderevidence.ExportSnapshotStatusLocked,
		SchemaVersion:          orderevidence.OrderEvidenceExportSnapshotSchemaVersion,
		LockedAt:               now,
		SnapshotData:           payloadBytes,
		SnapshotSHA256:         hex.EncodeToString(hash[:]),
		CreatedBy:              createdBy,
	}
	if err := snapshot.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceExportSnapshotInvalid, err)
	}

	canonical, err := s.repo.CreateOrGetForPackage(snapshot)
	if err != nil {
		return nil, err
	}
	if canonical == nil {
		return nil, ErrOrderEvidenceExportSnapshotInvalid
	}
	if err := canonical.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceExportSnapshotInvalid, err)
	}
	return canonical, nil
}

func buildOrderEvidenceExportSnapshotPayload(
	assembly *OrderEvidencePackageAssembly,
	exportedAt time.Time,
) OrderEvidenceExportSnapshotPayload {
	payload := OrderEvidenceExportSnapshotPayload{
		SchemaVersion:   orderevidence.OrderEvidenceExportSnapshotSchemaVersion,
		ExportType:      "order_evidence_manifest",
		ExportedAt:      exportedAt.UTC(),
		Completeness:    assembly.Completeness,
		TrackingContext: assembly.TrackingContext,
		TrackingEvents:  append([]shipping.TrackingEvent{}, assembly.TrackingEvents...),
		Sources:         append([]OrderEvidenceSourceReference{}, assembly.Sources...),
		Warnings:        append([]string{}, assembly.Warnings...),
	}
	if assembly.Order != nil {
		orderCopy := *assembly.Order
		orderCopy.Items = append([]order.OrderItem{}, assembly.Order.Items...)
		payload.Order = &orderCopy
	}
	if assembly.Package != nil {
		payload.Package = orderEvidenceExportPackage(assembly.Package)
	}
	return payload
}

func orderEvidenceExportPackage(pkg *orderevidence.OrderEvidencePackage) *OrderEvidenceExportPackage {
	if pkg == nil {
		return nil
	}
	result := &OrderEvidenceExportPackage{
		ID:                    pkg.ID,
		OrderID:               pkg.OrderID,
		SnapshotID:            pkg.SnapshotID,
		PackageVersion:        pkg.PackageVersion,
		Status:                pkg.Status,
		OrderTotalUSDSnapshot: pkg.OrderTotalUSDSnapshot,
		IsHighValue:           pkg.IsHighValue,
		HasSpokeTensionQC:     pkg.HasSpokeTensionQC,
		SchemaVersion:         pkg.SchemaVersion,
		CreatedBy:             pkg.CreatedBy,
		LockedAt:              copyTimePointer(pkg.LockedAt),
		CreatedAt:             pkg.CreatedAt,
		UpdatedAt:             pkg.UpdatedAt,
		Items:                 make([]OrderEvidenceExportItem, 0, len(pkg.Items)),
	}
	for _, item := range pkg.Items {
		exportItem := OrderEvidenceExportItem{
			ID:             item.ID,
			PackageID:      item.PackageID,
			OrderID:        item.OrderID,
			OrderItemID:    copyExportUintPointer(item.OrderItemID),
			SnapshotID:     copyExportUintPointer(item.SnapshotID),
			ItemType:       item.ItemType,
			Status:         item.Status,
			RequiredReason: item.RequiredReason,
			DataJSON:       append(json.RawMessage{}, item.DataJSON...),
			CapturedAt:     copyTimePointer(item.CapturedAt),
			CapturedBy:     item.CapturedBy,
			ContentSHA256:  item.ContentSHA256,
			Attachments:    make([]OrderEvidenceExportAttachment, 0, len(item.Attachments)),
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		}
		for _, attachment := range item.Attachments {
			exportItem.Attachments = append(exportItem.Attachments, OrderEvidenceExportAttachment{
				ID:               attachment.ID,
				EvidenceItemID:   attachment.EvidenceItemID,
				OriginalFilename: attachment.OriginalFilename,
				MimeType:         attachment.MimeType,
				SizeBytes:        attachment.SizeBytes,
				SHA256:           attachment.SHA256,
				UploadedBy:       attachment.UploadedBy,
				CreatedAt:        attachment.CreatedAt,
			})
		}
		result.Items = append(result.Items, exportItem)
	}
	return result
}

func copyTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copyValue := value.UTC()
	return &copyValue
}

func copyExportUintPointer(value *uint) *uint {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}
