package service

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"
	"gorm.io/datatypes"
)

var (
	ErrOrderEvidenceAdminStoreUnavailable = errors.New("order evidence admin store is unavailable")
	ErrOrderEvidenceAdminPackageNotFound  = errors.New("order evidence package not found")
	ErrOrderEvidenceAdminItemNotFound     = errors.New("order evidence item not found")
)

// OrderEvidenceAdminService is the application boundary for the future
// evidence tab. It owns transaction orchestration and order-scope checks while
// the lifecycle service remains reusable by non-HTTP workflows.
type OrderEvidenceAdminService struct {
	txManager    *repository.TxManager
	orderRepo    *repository.OrderRepository
	evidenceRepo *repository.OrderEvidenceRepository
	assembler    *OrderEvidencePackageAssembler
	lifecycle    *OrderEvidenceService
}

type OrderEvidenceAdminPackageResult struct {
	Package         *orderevidence.OrderEvidencePackage `json:"package"`
	Completeness    orderevidence.EvidenceCompleteness  `json:"completeness"`
	TrackingContext *OrderEvidenceTrackingContext       `json:"tracking_context,omitempty"`
	Sources         []OrderEvidenceSourceReference      `json:"sources,omitempty"`
	Warnings        []string                            `json:"warnings,omitempty"`
}

type OrderEvidenceAdminListInput struct {
	Page           int
	PageSize       int
	Search         string
	PackageStatus  string
	HighValue      *bool
	SpokeTensionQC *bool
}

type OrderEvidenceAdminListItem struct {
	OrderID               uint      `json:"order_id"`
	OrderNumber           string    `json:"order_number"`
	CustomerFirstName     string    `json:"customer_first_name"`
	CustomerLastName      string    `json:"customer_last_name"`
	CustomerEmail         string    `json:"customer_email"`
	OrderStatus           string    `json:"order_status"`
	PaymentStatus         string    `json:"payment_status"`
	ShippingStatus        string    `json:"shipping_status"`
	TotalAmount           float64   `json:"total_amount"`
	Currency              string    `json:"currency"`
	CreatedAt             time.Time `json:"created_at"`
	PackageID             uint      `json:"package_id"`
	PackageVersion        int       `json:"package_version"`
	PackageStatus         string    `json:"package_status"`
	OrderTotalUSDSnapshot float64   `json:"order_total_usd_snapshot"`
	IsHighValue           bool      `json:"is_high_value"`
	HasSpokeTensionQC     bool      `json:"has_spoke_tension_qc"`
	TotalEvidenceItems    int       `json:"total_evidence_items"`
	CompleteEvidenceItems int       `json:"complete_evidence_items"`
	WaivedEvidenceItems   int       `json:"waived_evidence_items"`
	PendingEvidenceItems  int       `json:"pending_evidence_items"`
}

type OrderEvidenceAdminListResult struct {
	Items    []OrderEvidenceAdminListItem `json:"items"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
	Total    int64                        `json:"total"`
}

type OrderEvidenceAdminItemUpdateInput struct {
	ItemID     uint
	Status     string
	DataJSON   datatypes.JSON
	CapturedAt *time.Time
	CapturedBy uint
}

func NewOrderEvidenceAdminService(
	txManager *repository.TxManager,
	orderRepo *repository.OrderRepository,
	evidenceRepo *repository.OrderEvidenceRepository,
	lifecycle *OrderEvidenceService,
) *OrderEvidenceAdminService {
	if lifecycle == nil {
		lifecycle = NewOrderEvidenceService()
	}
	return &OrderEvidenceAdminService{
		txManager:    txManager,
		orderRepo:    orderRepo,
		evidenceRepo: evidenceRepo,
		assembler:    NewOrderEvidencePackageAssembler(orderRepo, evidenceRepo, nil),
		lifecycle:    lifecycle,
	}
}

func (s *OrderEvidenceAdminService) ConfigureOrderEvidencePackageAssembler(
	assembler *OrderEvidencePackageAssembler,
) {
	if s == nil {
		return
	}
	s.assembler = assembler
}

func (s *OrderEvidenceAdminService) ListOrders(
	input OrderEvidenceAdminListInput,
) (*OrderEvidenceAdminListResult, error) {
	if s == nil || s.evidenceRepo == nil {
		return nil, ErrOrderEvidenceAdminStoreUnavailable
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	rows, total, err := s.evidenceRepo.ListAdminOrders(repository.OrderEvidenceAdminListQuery{
		Page:           input.Page,
		PageSize:       input.PageSize,
		Search:         input.Search,
		PackageStatus:  input.PackageStatus,
		HighValue:      input.HighValue,
		SpokeTensionQC: input.SpokeTensionQC,
	})
	if err != nil {
		return nil, err
	}

	items := make([]OrderEvidenceAdminListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, OrderEvidenceAdminListItem{
			OrderID:               row.OrderID,
			OrderNumber:           row.OrderNumber,
			CustomerFirstName:     row.CustomerFirstName,
			CustomerLastName:      row.CustomerLastName,
			CustomerEmail:         row.CustomerEmail,
			OrderStatus:           row.OrderStatus,
			PaymentStatus:         row.PaymentStatus,
			ShippingStatus:        row.ShippingStatus,
			TotalAmount:           row.TotalAmount,
			Currency:              row.Currency,
			CreatedAt:             row.CreatedAt,
			PackageID:             row.PackageID,
			PackageVersion:        row.PackageVersion,
			PackageStatus:         row.PackageStatus,
			OrderTotalUSDSnapshot: row.OrderTotalUSDSnapshot,
			IsHighValue:           row.IsHighValue,
			HasSpokeTensionQC:     row.HasSpokeTensionQC,
			TotalEvidenceItems:    row.TotalEvidenceItems,
			CompleteEvidenceItems: row.CompleteEvidenceItems,
			WaivedEvidenceItems:   row.WaivedEvidenceItems,
			PendingEvidenceItems:  row.PendingEvidenceItems,
		})
	}
	return &OrderEvidenceAdminListResult{
		Items:    items,
		Page:     input.Page,
		PageSize: input.PageSize,
		Total:    total,
	}, nil
}

func (s *OrderEvidenceAdminService) GetPackage(
	orderID uint,
) (*OrderEvidenceAdminPackageResult, error) {
	if s == nil || s.orderRepo == nil || s.evidenceRepo == nil || s.assembler == nil {
		return nil, ErrOrderEvidenceAdminStoreUnavailable
	}
	if orderID == 0 {
		return nil, errors.New("order id is required")
	}

	assembly, err := s.assembler.Assemble(orderID)
	if repository.IsRecordNotFound(err) {
		return nil, ErrOrderEvidenceAdminPackageNotFound
	}
	if err != nil {
		return nil, err
	}
	if assembly.Package == nil {
		return nil, ErrOrderEvidenceAdminPackageNotFound
	}
	return &OrderEvidenceAdminPackageResult{
		Package:         assembly.Package,
		Completeness:    assembly.Completeness,
		TrackingContext: assembly.TrackingContext,
		Sources:         assembly.Sources,
		Warnings:        assembly.Warnings,
	}, nil
}

func (s *OrderEvidenceAdminService) UpdateItem(
	orderID uint,
	input OrderEvidenceAdminItemUpdateInput,
) (*OrderEvidenceAdminPackageResult, error) {
	if s == nil || s.txManager == nil || s.evidenceRepo == nil || s.lifecycle == nil {
		return nil, ErrOrderEvidenceAdminStoreUnavailable
	}
	if orderID == 0 || input.ItemID == 0 {
		return nil, errors.New("order id and evidence item id are required")
	}

	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.OrderEvidence == nil {
			return ErrOrderEvidenceAdminStoreUnavailable
		}
		if _, _, err := repos.OrderEvidence.FindItemAndPackageByIDForOrder(input.ItemID, orderID); repository.IsRecordNotFound(err) {
			return ErrOrderEvidenceAdminItemNotFound
		} else if err != nil {
			return err
		}

		_, err := s.lifecycle.UpdateItem(repos, OrderEvidenceItemUpdateInput{
			OrderID:    orderID,
			ItemID:     input.ItemID,
			Status:     input.Status,
			DataJSON:   input.DataJSON,
			CapturedAt: input.CapturedAt,
			CapturedBy: input.CapturedBy,
		})
		return err
	})
	if err != nil {
		return nil, normalizeOrderEvidenceAdminError(err)
	}
	return s.GetPackage(orderID)
}

func (s *OrderEvidenceAdminService) LockPackage(
	orderID uint,
) (*OrderEvidenceAdminPackageResult, error) {
	if s == nil || s.txManager == nil || s.evidenceRepo == nil || s.lifecycle == nil {
		return nil, ErrOrderEvidenceAdminStoreUnavailable
	}
	if orderID == 0 {
		return nil, errors.New("order id is required")
	}

	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.OrderEvidence == nil {
			return ErrOrderEvidenceAdminStoreUnavailable
		}
		pkg, err := repos.OrderEvidence.FindLatestPackageByOrderID(orderID)
		if repository.IsRecordNotFound(err) {
			return ErrOrderEvidenceAdminPackageNotFound
		}
		if err != nil {
			return err
		}
		_, _, err = s.lifecycle.LockPackage(repos, pkg.ID, time.Now().UTC())
		return err
	})
	if err != nil {
		return nil, normalizeOrderEvidenceAdminError(err)
	}
	return s.GetPackage(orderID)
}

func (s *OrderEvidenceAdminService) CreateRevision(
	orderID uint,
	createdBy uint,
) (*OrderEvidenceAdminPackageResult, error) {
	if s == nil || s.txManager == nil || s.evidenceRepo == nil || s.lifecycle == nil {
		return nil, ErrOrderEvidenceAdminStoreUnavailable
	}
	if orderID == 0 || createdBy == 0 {
		return nil, errors.New("order id and created_by are required")
	}

	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.OrderEvidence == nil {
			return ErrOrderEvidenceAdminStoreUnavailable
		}
		pkg, err := repos.OrderEvidence.FindLatestPackageByOrderID(orderID)
		if repository.IsRecordNotFound(err) {
			return ErrOrderEvidenceAdminPackageNotFound
		}
		if err != nil {
			return err
		}
		_, err = s.lifecycle.CreateRevision(repos, pkg.ID, createdBy)
		return err
	})
	if err != nil {
		return nil, normalizeOrderEvidenceAdminError(err)
	}
	return s.GetPackage(orderID)
}

func normalizeOrderEvidenceAdminError(err error) error {
	switch {
	case repository.IsRecordNotFound(err):
		return ErrOrderEvidenceAdminPackageNotFound
	default:
		return err
	}
}
