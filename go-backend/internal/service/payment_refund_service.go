package service

import (
	currencydomain "commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AdminRefundLineItemInput struct {
	OrderItemID uint
	Quantity    int
	Restock     bool
}

type refundLineItemTotals struct {
	Quantity           int
	LineSubtotalAmount float64
	LineTaxAmount      float64
	LineDiscountAmount float64
	LineTotalAmount    float64
}

type VerifiedGatewayRefundInput struct {
	Provider        string
	OrderNumber     string
	TransactionID   string
	RefundID        string
	Amount          float64
	Currency        string
	GatewayResponse string
}

func (s *PaymentService) GetRefund(id uint) (*payment.Refund, error) {
	return s.paymentRepo.FindRefundByID(id)
}

func (s *PaymentService) GetOrderRefunds(orderID uint) ([]payment.Refund, error) {
	return s.paymentRepo.FindRefundsByOrderID(orderID)
}

func (s *PaymentService) CreateAdminRefund(refund *payment.Refund, adminUserID uint) error {
	if err := validateAdminRefundInput(refund); err != nil {
		return err
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		return createAdminRefundInTx(repos, refund, adminUserID)
	})
}

func validateAdminRefundInput(refund *payment.Refund) error {
	if refund == nil {
		return errors.New("refund is required")
	}
	if refund.OrderID == 0 {
		return errors.New("order_id is required")
	}
	if refund.TransactionID == 0 {
		return errors.New("transaction_id is required")
	}
	if refund.Amount <= 0 && len(refund.LineItems) == 0 {
		return errors.New("amount must be greater than zero")
	}
	return nil
}

func createAdminRefundInTx(repos repository.TxRepositories, refund *payment.Refund, adminUserID uint) error {
	if err := validateAdminRefundInput(refund); err != nil {
		return err
	}
	transaction, err := repos.Payment.FindTransactionByIDForUpdate(refund.TransactionID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return errors.New("transaction not found")
		}
		return err
	}
	if transaction.OrderID != refund.OrderID {
		return errors.New("transaction does not belong to order")
	}
	if transaction.Status != "completed" {
		return errors.New("transaction is not refundable")
	}

	o, err := repos.Order.FindByIDForUpdateWithItems(refund.OrderID)
	if err != nil {
		return normalizeOrderError(err)
	}
	if o.PaymentStatus != "paid" {
		return errors.New("order is not paid")
	}
	if o.Status == "refunded" {
		return errors.New("order is already refunded")
	}

	requestedAmount := roundRefundMoney(refund.Amount)
	requestedSubtotalAmount := requestedAmount
	if len(refund.LineItems) > 0 {
		lineItems, totals, err := buildRefundLineItems(repos, o, refund.LineItems)
		if err != nil {
			return err
		}
		if requestedAmount > 0 && absRefundMoney(requestedAmount-totals.LineTotalAmount) > 0.01 {
			return fmt.Errorf("refund amount %.2f does not match selected line item total %.2f", requestedAmount, totals.LineTotalAmount)
		}
		refund.LineItems = lineItems
		requestedAmount = totals.LineTotalAmount
		requestedSubtotalAmount = totals.LineSubtotalAmount
	} else {
		hasLineItemRefunds, err := repos.Payment.HasLineItemRefundsByOrderID(o.ID, "pending", "completed")
		if err != nil {
			return err
		}
		if hasLineItemRefunds {
			return errors.New("amount-only refunds cannot be created after item-level refunds exist for this order")
		}
	}

	adjustment, err := calculateRefundPromotionAdjustment(repos, o, requestedAmount, requestedSubtotalAmount)
	if err != nil {
		return err
	}
	fxSnapshot, _, err := ensureRefundFXSnapshot(refund, o, transaction.Currency)
	if err != nil {
		return err
	}
	reservedAmount, err := repos.Payment.SumRefundAmountByTransactionID(transaction.ID, "pending", "completed")
	if err != nil {
		return err
	}
	if adjustment.NetAmount-(transaction.Amount-reservedAmount) > 0.01 {
		return fmt.Errorf("refund amount %.2f exceeds refundable amount %.2f", adjustment.NetAmount, transaction.Amount-reservedAmount)
	}
	if err := validateHistoricalRefundFXCap(fxSnapshot, transaction, adjustment.NetAmount, reservedAmount); err != nil {
		return err
	}

	refund.RequestedAmount = adjustment.RequestedAmount
	refund.Amount = adjustment.NetAmount
	refund.DiscountClawbackAmount = adjustment.DiscountClawbackAmount
	refund.CalculationSnapshot = adjustment.CalculationSnapshot
	refund.FXSnapshotData = currencydomain.OrderFXSnapshotJSON(fxSnapshot)
	refund.Status = "pending"
	refund.RefundID = nil
	refund.GatewayResponse = ""
	refund.CompletedAt = nil
	refund.RefundedBy = adminUserID

	return repos.Payment.CreateRefund(refund)
}

func buildRefundLineItems(repos repository.TxRepositories, o *order.Order, requestedItems []payment.RefundLineItem) ([]payment.RefundLineItem, refundLineItemTotals, error) {
	if len(requestedItems) == 0 {
		return nil, refundLineItemTotals{}, errors.New("refund line items are required")
	}

	hasAmountOnlyRefunds, err := repos.Payment.HasAmountOnlyRefundsByOrderID(o.ID, "pending", "completed")
	if err != nil {
		return nil, refundLineItemTotals{}, err
	}
	if hasAmountOnlyRefunds {
		return nil, refundLineItemTotals{}, errors.New("item-level refunds cannot be created after amount-only refunds exist for this order")
	}

	refundedQuantities, err := repos.Payment.SumRefundLineItemQuantitiesByOrderID(o.ID, "pending", "completed")
	if err != nil {
		return nil, refundLineItemTotals{}, err
	}

	itemsByID := make(map[uint]order.OrderItem, len(o.Items))
	for _, item := range o.Items {
		itemsByID[item.ID] = item
	}

	seen := make(map[uint]struct{}, len(requestedItems))
	lineItems := make([]payment.RefundLineItem, 0, len(requestedItems))
	totals := refundLineItemTotals{}
	for _, requested := range requestedItems {
		if requested.OrderItemID == 0 {
			return nil, totals, errors.New("order_item_id is required for item-level refunds")
		}
		if _, exists := seen[requested.OrderItemID]; exists {
			return nil, totals, fmt.Errorf("duplicate refund line item for order_item_id %d", requested.OrderItemID)
		}
		seen[requested.OrderItemID] = struct{}{}

		if requested.Quantity <= 0 {
			return nil, totals, fmt.Errorf("refund quantity for order_item_id %d must be greater than zero", requested.OrderItemID)
		}
		item, ok := itemsByID[requested.OrderItemID]
		if !ok {
			return nil, totals, fmt.Errorf("order_item_id %d does not belong to order", requested.OrderItemID)
		}
		if requested.Restock && item.VariantID == nil {
			return nil, totals, fmt.Errorf("order_item_id %d cannot be restocked because it has no variant snapshot", requested.OrderItemID)
		}
		alreadyRefunded := refundedQuantities[item.ID]
		availableQuantity := item.Quantity - alreadyRefunded
		if requested.Quantity > availableQuantity {
			return nil, totals, fmt.Errorf("refund quantity %d exceeds available quantity %d for order_item_id %d", requested.Quantity, availableQuantity, item.ID)
		}

		lineItem := buildRefundLineItemSnapshot(item, requested.Quantity, requested.Restock)
		lineItems = append(lineItems, lineItem)
		totals.Quantity += lineItem.Quantity
		totals.LineSubtotalAmount = roundRefundMoney(totals.LineSubtotalAmount + lineItem.LineSubtotalAmount)
		totals.LineTaxAmount = roundRefundMoney(totals.LineTaxAmount + lineItem.LineTaxAmount)
		totals.LineDiscountAmount = roundRefundMoney(totals.LineDiscountAmount + lineItem.LineDiscountAmount)
		totals.LineTotalAmount = roundRefundMoney(totals.LineTotalAmount + lineItem.LineTotalAmount)
	}

	if totals.LineTotalAmount <= 0 {
		return nil, totals, errors.New("selected line items do not produce a refundable amount")
	}
	return lineItems, totals, nil
}

func buildRefundLineItemSnapshot(item order.OrderItem, quantity int, restock bool) payment.RefundLineItem {
	ratio := 0.0
	if item.Quantity <= 0 {
		ratio = 0
	} else {
		ratio = float64(quantity) / float64(item.Quantity)
	}
	lineSubtotal := roundRefundMoney(item.Subtotal * ratio)
	lineTax := roundRefundMoney(item.TaxAmount * ratio)
	lineDiscount := roundRefundMoney(item.Discount * ratio)
	lineTotal := roundRefundMoney(item.Total * ratio)
	if lineTotal <= 0 {
		lineTotal = roundRefundMoney(lineSubtotal + lineTax - lineDiscount)
	}

	return payment.RefundLineItem{
		OrderID:            item.OrderID,
		OrderItemID:        item.ID,
		ProductID:          item.ProductID,
		VariantID:          item.VariantID,
		ProductName:        item.ProductName,
		SKU:                item.SKU,
		Quantity:           quantity,
		UnitPrice:          roundRefundMoney(item.Price),
		LineSubtotalAmount: lineSubtotal,
		LineTaxAmount:      lineTax,
		LineDiscountAmount: lineDiscount,
		LineTotalAmount:    lineTotal,
		Restock:            restock,
	}
}

func absRefundMoney(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func restoreRefundLineItemStock(repos repository.TxRepositories, lineItems []payment.RefundLineItem, restockedAt time.Time) error {
	for _, item := range lineItems {
		if !item.Restock || item.Quantity <= 0 || item.VariantID == nil {
			continue
		}

		claimed, err := repos.Payment.MarkRefundLineItemRestocked(item.ID, restockedAt)
		if err != nil {
			return err
		}
		if !claimed {
			continue
		}
		if err := repos.Product.IncrementVariantStock(*item.VariantID, item.Quantity); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to restore stock for refunded variant %d: %w", *item.VariantID, err)
		}
	}
	return nil
}

func (s *PaymentService) RecordVerifiedGatewayRefund(input VerifiedGatewayRefundInput) error {
	if input.Provider == "" {
		return errors.New("provider is required")
	}
	if input.TransactionID == "" {
		return errors.New("transaction_id is required")
	}
	if input.RefundID == "" {
		return errors.New("refund_id is required")
	}
	if input.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if existing, err := repos.Payment.FindRefundByRefundID(input.RefundID); err == nil {
			if existing.Status == "completed" {
				return restoreRefundLineItemStock(repos, existing.LineItems, time.Now())
			}
			return errors.New("refund id is already used")
		} else if !repository.IsRecordNotFound(err) {
			return err
		}

		transaction, err := repos.Payment.FindTransactionByTransactionIDForUpdate(input.TransactionID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return errors.New("transaction not found")
			}
			return err
		}
		if transaction.Status == "refunded" {
			return errors.New("transaction is already fully refunded")
		}
		if transaction.Status != "completed" {
			return errors.New("transaction is not refundable")
		}
		if input.Currency != "" && transaction.Currency != "" && !strings.EqualFold(input.Currency, transaction.Currency) {
			return fmt.Errorf("refund currency %s does not match transaction currency %s", input.Currency, transaction.Currency)
		}

		o, err := repos.Order.FindByIDForUpdate(transaction.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		if input.OrderNumber != "" && o.OrderNumber != input.OrderNumber {
			return errors.New("refund order_number does not match transaction order")
		}
		if o.PaymentStatus == "refunded" {
			return errors.New("order is already refunded")
		}
		if o.PaymentStatus != "paid" {
			return errors.New("order is not paid")
		}

		fxSnapshot, _, err := ensureRefundFXSnapshot(&payment.Refund{}, o, transaction.Currency)
		if err != nil {
			return err
		}
		reservedAmount, err := repos.Payment.SumRefundAmountByTransactionID(transaction.ID, "pending", "completed")
		if err != nil {
			return err
		}

		now := time.Now()
		refundID := input.RefundID
		pendingRefund, err := repos.Payment.FindPendingRefundByTransactionAndAmount(transaction.ID, input.Amount)
		if err != nil && !repository.IsRecordNotFound(err) {
			return err
		}
		reservedBeforeCurrent := reservedAmount
		if pendingRefund != nil {
			reservedBeforeCurrent -= pendingRefund.Amount
		}
		if input.Amount-(transaction.Amount-reservedBeforeCurrent) > 0.01 {
			return fmt.Errorf("refund amount %.2f exceeds refundable amount %.2f", input.Amount, transaction.Amount-reservedBeforeCurrent)
		}
		if err := validateHistoricalRefundFXCap(fxSnapshot, transaction, input.Amount, reservedBeforeCurrent); err != nil {
			return err
		}
		if pendingRefund != nil && err == nil {
			wasCompleted := pendingRefund.Status == "completed"
			pendingRefund.Status = "completed"
			pendingRefund.RefundID = &refundID
			pendingRefund.GatewayResponse = input.GatewayResponse
			pendingRefund.CompletedAt = &now
			pendingRefund.FXSnapshotData = currencydomain.OrderFXSnapshotJSON(fxSnapshot)
			if err := repos.Payment.UpdateRefund(pendingRefund); err != nil {
				return err
			}
			if !wasCompleted {
				if err := restoreRefundLineItemStock(repos, pendingRefund.LineItems, now); err != nil {
					return err
				}
			}
		} else {
			refund := &payment.Refund{
				OrderID:         transaction.OrderID,
				TransactionID:   transaction.ID,
				RefundID:        &refundID,
				Amount:          input.Amount,
				RequestedAmount: input.Amount,
				Status:          "completed",
				GatewayResponse: input.GatewayResponse,
				FXSnapshotData:  currencydomain.OrderFXSnapshotJSON(fxSnapshot),
				CompletedAt:     &now,
			}
			if err := repos.Payment.CreateRefund(refund); err != nil {
				return err
			}
		}

		completedAmount, err := repos.Payment.SumRefundAmountByTransactionID(transaction.ID, "completed")
		if err != nil {
			return err
		}
		if completedAmount >= transaction.Amount-0.01 {
			transaction.Status = "refunded"
			if err := repos.Payment.UpdateTransaction(transaction); err != nil {
				return err
			}
		}

		orderRefundedAmount, err := repos.Payment.SumRefundAmountByOrderID(o.ID, "completed")
		if err != nil {
			return err
		}
		if orderRefundedAmount >= o.TotalAmount-0.01 {
			if err := repos.Order.UpdatePaymentStatus(o.ID, "refunded"); err != nil {
				return err
			}
			return repos.Order.UpdateStatus(o.ID, "refunded")
		}

		return nil
	})
}
