package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

var (
	ErrAfterSalesCaseNotFound                 = errors.New("after-sales case not found")
	ErrAfterSalesOrderNotFound                = errors.New("order not found")
	ErrAfterSalesOrderNotEligible             = errors.New("order is not eligible for after-sales processing")
	ErrAfterSalesTypeInvalid                  = errors.New("invalid after-sales case type")
	ErrAfterSalesStatusInvalid                = errors.New("invalid after-sales case status")
	ErrAfterSalesTransitionInvalid            = errors.New("invalid after-sales case status transition")
	ErrAfterSalesItemsRequired                = errors.New("at least one after-sales item is required")
	ErrAfterSalesItemNotFound                 = errors.New("after-sales order item not found")
	ErrAfterSalesItemOrderMismatch            = errors.New("after-sales item does not belong to the order")
	ErrAfterSalesQuantityInvalid              = errors.New("after-sales item quantity must be positive")
	ErrAfterSalesQuantityExceeded             = errors.New("after-sales item quantity exceeds the remaining eligible quantity")
	ErrAfterSalesDescriptionRequired          = errors.New("after-sales case description is required")
	ErrAfterSalesRequestAlreadyExists         = errors.New("an active after-sales request already exists for this order")
	ErrAfterSalesReturnCarrierRequired        = errors.New("return carrier is required for in-transit returns")
	ErrAfterSalesReturnTrackingRequired       = errors.New("return tracking number is required for in-transit returns")
	ErrAfterSalesReturnWarehouseRequired      = errors.New("return warehouse and address are required before return transit")
	ErrAfterSalesReturnReceiverRequired       = errors.New("return receiver is required when marking a return received")
	ErrAfterSalesAttachmentKindInvalid        = errors.New("invalid after-sales attachment kind")
	ErrAfterSalesAttachmentNotFound           = errors.New("after-sales attachment not found")
	ErrAfterSalesAttachmentStorageUnavailable = errors.New("after-sales attachment storage is unavailable")
)

type AfterSalesCaseItemInput struct {
	OrderItemID uint `json:"order_item_id" binding:"required"`
	Quantity    int  `json:"quantity" binding:"required"`
}

type CreateAfterSalesCaseInput struct {
	OrderID     uint
	Type        string
	Reason      string
	Description string
	Items       []AfterSalesCaseItemInput
	CreatedBy   uint
}

// CreateCustomerAfterSalesRequestInput is intentionally narrower than the
// admin case input. Customer submissions never choose a resolution type,
// refund amount, or item IDs.
type CreateCustomerAfterSalesRequestInput struct {
	OrderID     uint
	Reason      string
	Description string
	Attachments []aftersales.AfterSalesCaseAttachment
	CreatedBy   uint
}

type ListAfterSalesCasesInput struct {
	Page     int
	PageSize int
	Status   string
	Type     string
	Search   string
}

type AfterSalesService struct {
	caseRepo          *repository.AfterSalesCaseRepository
	orderRepo         *repository.OrderRepository
	refundReviewRepo  *repository.AfterSalesRefundReviewRepository
	txManager         *repository.TxManager
	userRepo          *repository.UserRepository
	attachmentStorage storage.StorageService
}

func NewAfterSalesService(
	caseRepo *repository.AfterSalesCaseRepository,
	orderRepo *repository.OrderRepository,
	refundReviewRepos ...*repository.AfterSalesRefundReviewRepository,
) *AfterSalesService {
	var refundReviewRepo *repository.AfterSalesRefundReviewRepository
	if len(refundReviewRepos) > 0 {
		refundReviewRepo = refundReviewRepos[0]
	}
	return &AfterSalesService{
		caseRepo:         caseRepo,
		orderRepo:        orderRepo,
		refundReviewRepo: refundReviewRepo,
	}
}

func (s *AfterSalesService) ConfigureTxManager(txManager *repository.TxManager) {
	if s != nil {
		s.txManager = txManager
	}
}

func (s *AfterSalesService) ConfigureUserRepository(userRepo *repository.UserRepository) {
	if s != nil {
		s.userRepo = userRepo
	}
}

func (s *AfterSalesService) CreateCase(input CreateAfterSalesCaseInput) (*aftersales.AfterSalesCase, error) {
	if s == nil || s.caseRepo == nil || s.orderRepo == nil {
		return nil, errors.New("after-sales service is not configured")
	}
	if input.OrderID == 0 {
		return nil, ErrAfterSalesOrderNotFound
	}

	input.Type = strings.TrimSpace(input.Type)
	input.Reason = strings.TrimSpace(input.Reason)
	input.Description = strings.TrimSpace(input.Description)
	if !aftersales.IsValidType(input.Type) {
		return nil, ErrAfterSalesTypeInvalid
	}
	if input.Reason == "" {
		return nil, errors.New("after-sales case reason is required")
	}
	if len(input.Items) == 0 {
		return nil, ErrAfterSalesItemsRequired
	}

	orderRecord, err := s.orderRepo.FindByID(input.OrderID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrAfterSalesOrderNotFound
		}
		return nil, err
	}
	if !eligibleForAfterSales(orderRecord) {
		return nil, ErrAfterSalesOrderNotEligible
	}

	orderItems := make(map[uint]order.OrderItem, len(orderRecord.Items))
	for _, item := range orderRecord.Items {
		orderItems[item.ID] = item
	}

	caseItems := make([]aftersales.AfterSalesCaseItem, 0, len(input.Items))
	seenItems := make(map[uint]struct{}, len(input.Items))
	for _, itemInput := range input.Items {
		if itemInput.OrderItemID == 0 {
			return nil, ErrAfterSalesItemNotFound
		}
		if itemInput.Quantity <= 0 {
			return nil, ErrAfterSalesQuantityInvalid
		}
		if _, exists := seenItems[itemInput.OrderItemID]; exists {
			return nil, fmt.Errorf("%w: duplicate order item %d", ErrAfterSalesItemOrderMismatch, itemInput.OrderItemID)
		}
		seenItems[itemInput.OrderItemID] = struct{}{}

		orderItem, exists := orderItems[itemInput.OrderItemID]
		if !exists {
			return nil, ErrAfterSalesItemOrderMismatch
		}
		activeQuantity, err := s.caseRepo.SumActiveQuantity(orderItem.ID)
		if err != nil {
			return nil, err
		}
		if activeQuantity+itemInput.Quantity > orderItem.Quantity {
			return nil, fmt.Errorf(
				"%w: item %d has %d remaining",
				ErrAfterSalesQuantityExceeded,
				orderItem.ID,
				orderItem.Quantity-activeQuantity,
			)
		}

		caseItems = append(caseItems, aftersales.AfterSalesCaseItem{
			OrderID:     orderRecord.ID,
			OrderItemID: orderItem.ID,
			ProductID:   orderItem.ProductID,
			VariantID:   orderItem.VariantID,
			ProductName: orderItem.ProductName,
			SKU:         orderItem.SKU,
			Quantity:    itemInput.Quantity,
		})
	}

	caseRecord := &aftersales.AfterSalesCase{
		OrderID:     orderRecord.ID,
		Type:        input.Type,
		Status:      aftersales.StatusRequested,
		Reason:      input.Reason,
		Description: input.Description,
		CreatedBy:   input.CreatedBy,
		UpdatedBy:   input.CreatedBy,
		Items:       caseItems,
	}
	if err := caseRecord.Validate(); err != nil {
		return nil, err
	}
	if err := s.caseRepo.CreateWithItems(caseRecord, caseItems); err != nil {
		return nil, err
	}
	return s.GetCase(caseRecord.ID)
}

// CreateCustomerRequest creates a neutral customer-originated after-sales
// case. It only records the request for review; it never creates a refund or
// calls a payment provider.
func (s *AfterSalesService) CreateCustomerRequest(input CreateCustomerAfterSalesRequestInput) (*aftersales.AfterSalesCase, error) {
	if s == nil || s.caseRepo == nil || s.orderRepo == nil {
		return nil, errors.New("after-sales service is not configured")
	}
	if s.txManager == nil {
		return nil, errors.New("after-sales transaction manager is not configured")
	}
	if input.OrderID == 0 {
		return nil, ErrAfterSalesOrderNotFound
	}

	input.Reason = strings.TrimSpace(input.Reason)
	input.Description = strings.TrimSpace(input.Description)
	if input.Reason == "" {
		return nil, errors.New("after-sales case reason is required")
	}
	if input.Description == "" {
		return nil, ErrAfterSalesDescriptionRequired
	}
	if len(input.Reason) > 500 {
		return nil, errors.New("after-sales case reason is too long")
	}
	if len(input.Description) > 5000 {
		return nil, errors.New("after-sales case description is too long")
	}
	for _, attachment := range input.Attachments {
		if attachment.Kind != aftersales.AttachmentKindImage && attachment.Kind != aftersales.AttachmentKindVideo {
			return nil, ErrAfterSalesAttachmentKindInvalid
		}
	}

	var caseID uint
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.AfterSalesCase == nil {
			return errors.New("after-sales case repository is not configured for transactions")
		}

		orderRecord, err := repos.Order.FindByIDForUpdateWithItems(input.OrderID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrAfterSalesOrderNotFound
			}
			return err
		}
		if !eligibleForAfterSales(orderRecord) {
			return ErrAfterSalesOrderNotEligible
		}

		existingCases, err := repos.AfterSalesCase.FindByOrderID(orderRecord.ID, "")
		if err != nil {
			return err
		}
		for _, existingCase := range existingCases {
			// Customer-originated requests are informational snapshots and are
			// intentionally limited to one active request per order. They must not
			// block a customer from opening a new request while an operator-managed
			// case is still in transit or awaiting resolution.
			if existingCase.Type == aftersales.TypeCustomerRequest && !aftersales.IsTerminalStatus(existingCase.Status) {
				return ErrAfterSalesRequestAlreadyExists
			}
		}

		caseItems := make([]aftersales.AfterSalesCaseItem, 0, len(orderRecord.Items))
		for _, orderItem := range orderRecord.Items {
			if orderItem.Quantity <= 0 {
				continue
			}
			// This is a support snapshot, not an item reservation. Include the
			// full order even when another case has the item in transit so the
			// customer can still contact support about that same case or order.
			caseItems = append(caseItems, aftersales.AfterSalesCaseItem{
				OrderID:     orderRecord.ID,
				OrderItemID: orderItem.ID,
				ProductID:   orderItem.ProductID,
				VariantID:   orderItem.VariantID,
				ProductName: orderItem.ProductName,
				SKU:         orderItem.SKU,
				Quantity:    orderItem.Quantity,
			})
		}
		if len(caseItems) == 0 {
			return ErrAfterSalesItemsRequired
		}

		caseRecord := &aftersales.AfterSalesCase{
			OrderID:     orderRecord.ID,
			Type:        aftersales.TypeCustomerRequest,
			Status:      aftersales.StatusRequested,
			Reason:      input.Reason,
			Description: input.Description,
			CreatedBy:   input.CreatedBy,
			UpdatedBy:   input.CreatedBy,
		}
		if err := caseRecord.Validate(); err != nil {
			return err
		}
		if err := repos.AfterSalesCase.CreateWithItemsAndAttachments(caseRecord, caseItems, input.Attachments); err != nil {
			return err
		}
		caseID = caseRecord.ID
		return nil
	})
	if err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, ErrAfterSalesRequestAlreadyExists
		}
		return nil, err
	}
	return s.GetCase(caseID)
}

func (s *AfterSalesService) ListCasesByOrder(orderID uint, status string) ([]aftersales.AfterSalesCase, error) {
	if s == nil || s.caseRepo == nil || s.orderRepo == nil {
		return nil, errors.New("after-sales service is not configured")
	}
	if orderID == 0 {
		return nil, ErrAfterSalesOrderNotFound
	}
	if status != "" && !aftersales.IsValidStatus(status) {
		return nil, ErrAfterSalesStatusInvalid
	}
	if _, err := s.orderRepo.FindByID(orderID); err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrAfterSalesOrderNotFound
		}
		return nil, err
	}
	return s.caseRepo.FindByOrderID(orderID, strings.TrimSpace(status))
}

// ListCustomerCasesByOrder returns the customer's cases with their status
// history and refund-review details populated. Ownership is enforced by the
// public order handler before this method is called.
func (s *AfterSalesService) ListCustomerCasesByOrder(orderID uint, status string) ([]aftersales.AfterSalesCase, error) {
	records, err := s.ListCasesByOrder(orderID, status)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return records, nil
	}

	detailed := make([]aftersales.AfterSalesCase, 0, len(records))
	for _, record := range records {
		caseRecord, err := s.GetCase(record.ID)
		if err != nil {
			return nil, err
		}
		detailed = append(detailed, *caseRecord)
	}
	return detailed, nil
}

func (s *AfterSalesService) GetCase(id uint) (*aftersales.AfterSalesCase, error) {
	if s == nil || s.caseRepo == nil {
		return nil, errors.New("after-sales service is not configured")
	}
	if id == 0 {
		return nil, ErrAfterSalesCaseNotFound
	}

	record, err := s.caseRepo.FindByID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrAfterSalesCaseNotFound
		}
		return nil, err
	}
	s.populateEventOperatorNames(record)
	s.populateRefundReviewDetails(record)
	return record, nil
}

func (s *AfterSalesService) ListAdminCases(input ListAfterSalesCasesInput) ([]aftersales.AfterSalesCase, int64, error) {
	if s == nil || s.caseRepo == nil {
		return nil, 0, errors.New("after-sales service is not configured")
	}

	input.Status = strings.TrimSpace(input.Status)
	input.Type = strings.TrimSpace(input.Type)
	input.Search = strings.TrimSpace(input.Search)
	if input.Status == "all" {
		input.Status = ""
	}
	if input.Type == "all" {
		input.Type = ""
	}
	if input.Status != "" && !aftersales.IsValidStatus(input.Status) {
		return nil, 0, ErrAfterSalesStatusInvalid
	}
	if input.Type != "" && !aftersales.IsValidType(input.Type) {
		return nil, 0, ErrAfterSalesTypeInvalid
	}

	return s.caseRepo.List(input.Page, input.PageSize, input.Status, input.Type, input.Search)
}

type UpdateAfterSalesStatusInput struct {
	Status           string
	Resolution       string
	UpdatedBy        uint
	ReturnShipmentID uint
	WarehouseName    string
	WarehouseAddress string
	Carrier          string
	TrackingNumber   string
	TrackingURL      string
	LabelURL         string
	ShippedAt        *time.Time
	ReceivedAt       *time.Time
	ReceivedBy       *uint
}

func (s *AfterSalesService) UpdateStatus(id uint, status, resolution string, updatedBy uint) (*aftersales.AfterSalesCase, error) {
	return s.UpdateStatusWithReturnShipment(id, UpdateAfterSalesStatusInput{
		Status: status, Resolution: resolution, UpdatedBy: updatedBy,
	})
}

func (s *AfterSalesService) UpdateStatusWithReturnShipment(id uint, input UpdateAfterSalesStatusInput) (*aftersales.AfterSalesCase, error) {
	if s == nil || s.caseRepo == nil {
		return nil, errors.New("after-sales service is not configured")
	}
	if id == 0 {
		return nil, ErrAfterSalesCaseNotFound
	}
	status := strings.TrimSpace(input.Status)
	resolution := strings.TrimSpace(input.Resolution)
	if !aftersales.IsValidStatus(status) {
		return nil, ErrAfterSalesStatusInvalid
	}

	record, err := s.caseRepo.FindByIDForUpdate(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrAfterSalesCaseNotFound
		}
		return nil, err
	}
	if !record.CanTransitionTo(status) {
		return nil, fmt.Errorf(
			"%w: cannot move from %s to %s",
			ErrAfterSalesTransitionInvalid,
			record.Status,
			status,
		)
	}
	var shipment *aftersales.AfterSalesReturnShipment
	if status == aftersales.StatusAwaitingReturn || status == aftersales.StatusReturnInTransit || status == aftersales.StatusReceived || input.ReturnShipmentID != 0 || strings.TrimSpace(input.Carrier) != "" || strings.TrimSpace(input.TrackingNumber) != "" {
		shipment = &aftersales.AfterSalesReturnShipment{
			ID: input.ReturnShipmentID, WarehouseName: strings.TrimSpace(input.WarehouseName), WarehouseAddress: strings.TrimSpace(input.WarehouseAddress),
			Carrier: strings.TrimSpace(input.Carrier), TrackingNumber: strings.TrimSpace(input.TrackingNumber), TrackingURL: strings.TrimSpace(input.TrackingURL), LabelURL: strings.TrimSpace(input.LabelURL),
			ShippedAt: input.ShippedAt, ReceivedAt: input.ReceivedAt, ReceivedBy: input.ReceivedBy,
		}
		// Awaiting-return starts a new package. Later transitions continue to
		// update the most recent package unless the caller supplies an explicit
		// shipment ID (for a replacement label or another package).
		if len(record.ReturnShipments) > 0 && shipment.ID == 0 && status != aftersales.StatusAwaitingReturn {
			previous := record.ReturnShipments[len(record.ReturnShipments)-1]
			shipment.ID = previous.ID
			if shipment.WarehouseName == "" {
				shipment.WarehouseName = previous.WarehouseName
			}
			if shipment.WarehouseAddress == "" {
				shipment.WarehouseAddress = previous.WarehouseAddress
			}
			if shipment.Carrier == "" {
				shipment.Carrier = previous.Carrier
			}
			if shipment.TrackingNumber == "" {
				shipment.TrackingNumber = previous.TrackingNumber
			}
			if shipment.TrackingURL == "" {
				shipment.TrackingURL = previous.TrackingURL
			}
			if shipment.LabelURL == "" {
				shipment.LabelURL = previous.LabelURL
			}
			if shipment.ShippedAt == nil {
				shipment.ShippedAt = previous.ShippedAt
			}
		}
		if shipment.ID != 0 {
			for _, previous := range record.ReturnShipments {
				if previous.ID != shipment.ID {
					continue
				}
				if shipment.WarehouseName == "" {
					shipment.WarehouseName = previous.WarehouseName
				}
				if shipment.WarehouseAddress == "" {
					shipment.WarehouseAddress = previous.WarehouseAddress
				}
				if shipment.Carrier == "" {
					shipment.Carrier = previous.Carrier
				}
				if shipment.TrackingNumber == "" {
					shipment.TrackingNumber = previous.TrackingNumber
				}
				if shipment.TrackingURL == "" {
					shipment.TrackingURL = previous.TrackingURL
				}
				if shipment.LabelURL == "" {
					shipment.LabelURL = previous.LabelURL
				}
				if shipment.ShippedAt == nil {
					shipment.ShippedAt = previous.ShippedAt
				}
				if shipment.ReceivedAt == nil {
					shipment.ReceivedAt = previous.ReceivedAt
				}
				if shipment.ReceivedBy == nil {
					shipment.ReceivedBy = previous.ReceivedBy
				}
				break
			}
		}
	}
	if status == aftersales.StatusAwaitingReturn {
		if shipment == nil || shipment.WarehouseName == "" || shipment.WarehouseAddress == "" {
			return nil, ErrAfterSalesReturnWarehouseRequired
		}
	}
	if status == aftersales.StatusReturnInTransit {
		if shipment == nil || shipment.Carrier == "" {
			return nil, ErrAfterSalesReturnCarrierRequired
		}
		if shipment.TrackingNumber == "" {
			return nil, ErrAfterSalesReturnTrackingRequired
		}
		if shipment.WarehouseName == "" || shipment.WarehouseAddress == "" {
			return nil, ErrAfterSalesReturnWarehouseRequired
		}
		if shipment.ShippedAt == nil {
			now := time.Now().UTC()
			shipment.ShippedAt = &now
		}
	}
	if status == aftersales.StatusReceived {
		if shipment == nil || shipment.ID == 0 {
			return nil, ErrAfterSalesReturnReceiverRequired
		}
		if shipment.ReceivedBy == nil || *shipment.ReceivedBy == 0 {
			return nil, ErrAfterSalesReturnReceiverRequired
		}
		if shipment.ReceivedAt == nil {
			now := time.Now().UTC()
			shipment.ReceivedAt = &now
		}
	}
	updated, err := s.caseRepo.UpdateStatusAndSaveReturnShipmentIfCurrent(id, record.Status, status, resolution, input.UpdatedBy, shipment)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, fmt.Errorf(
			"%w: case status changed before update",
			ErrAfterSalesTransitionInvalid,
		)
	}
	return s.GetCase(id)
}

func (s *AfterSalesService) populateEventOperatorNames(record *aftersales.AfterSalesCase) {
	if record == nil || len(record.Events) == 0 {
		return
	}

	ids := make([]uint, 0, len(record.Events))
	for _, event := range record.Events {
		if event.UpdatedBy > 0 {
			ids = append(ids, event.UpdatedBy)
		}
	}

	namesByID := make(map[uint]string, len(ids))
	if s != nil && s.userRepo != nil && len(ids) > 0 {
		users, err := s.userRepo.FindByIDs(ids)
		if err == nil {
			for _, operator := range users {
				namesByID[operator.ID] = afterSalesOperatorName(operator)
			}
		}
	}

	for index := range record.Events {
		event := &record.Events[index]
		switch {
		case event.UpdatedBy == 0:
			event.OperatorName = "系统"
		case namesByID[event.UpdatedBy] != "":
			event.OperatorName = namesByID[event.UpdatedBy]
		default:
			event.OperatorName = fmt.Sprintf("账号 #%d", event.UpdatedBy)
		}
	}
}

func eligibleForAfterSales(record *order.Order) bool {
	if record == nil || record.PaymentStatus != "paid" {
		return false
	}
	switch record.Status {
	case "paid", "processing", "shipped", "completed":
		return true
	default:
		return false
	}
}
