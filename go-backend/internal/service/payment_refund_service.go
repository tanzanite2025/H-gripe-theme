package service

import (
	currencydomain "commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/payment"
	domainpricing "commerce-platform/internal/domain/pricing"
	"commerce-platform/internal/repository"
	"encoding/json"
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

const adminRefundIdempotencyScope = "admin_refund_create"

var (
	ErrPaymentRefundIdempotencyUnavailable  = errors.New("payment refund idempotency is not configured")
	ErrPaymentRefundIdempotencyHashRequired = errors.New("payment refund idempotency request hash is required")
	ErrPaymentRefundIdempotencyConflict     = errors.New("idempotency key was already used for a different refund request")
	ErrPaymentRefundIdempotencyInProgress   = errors.New("idempotent refund request is already being processed")
)

type refundLineItemTotals struct {
	Quantity           int
	LineSubtotalAmount domainmoney.Money
	LineTaxAmount      domainmoney.Money
	LineDiscountAmount domainmoney.Money
	LineTotalAmount    domainmoney.Money
}

type refundPaymentSplit struct {
	GatewayAmount  domainmoney.Money
	TotalAvailable domainmoney.Money
}

type VerifiedGatewayRefundInput struct {
	Provider              string
	OrderNumber           string
	TransactionID         string
	RefundID              string
	ProviderStatus        string            // provider status carried by the verified webhook resource
	ProviderRefundAmount  domainmoney.Money // actual net amount confirmed by the payment provider
	RequestedRefundAmount domainmoney.Money // original refund request before coupon or loyalty deductions
	// SettlementAmountMinor is the positive net amount deducted from the
	// provider's settlement balance. It is usually the absolute value of a
	// Stripe BalanceTransaction.Net and may be absent for providers that do not
	// expose a settlement transaction in the webhook.
	SettlementAmountMinor          int64
	SettlementCurrency             string
	SettlementBalanceTransactionID string
	ErrorMessage                   string // provider-declared failure/exception detail
	GatewayResponse                string
}

func applyRefundSettlementFacts(
	refund *payment.Refund,
	snapshot currencydomain.OrderFXSnapshot,
	refundMoney domainmoney.Money,
	settlementAmountMinor int64,
	settlementCurrency string,
	settlementBalanceTransactionID string,
) error {
	if refund == nil || settlementAmountMinor == 0 || strings.TrimSpace(settlementCurrency) == "" {
		return nil
	}
	if settlementAmountMinor < 0 {
		if settlementAmountMinor == -1<<63 {
			return errors.New("provider settlement amount overflows")
		}
		settlementAmountMinor = -settlementAmountMinor
	}
	settlementCode, err := currencydomain.ParseCode(settlementCurrency)
	if err != nil {
		return fmt.Errorf("invalid settlement currency: %w", err)
	}
	refund.SettlementAmountMinor = settlementAmountMinor
	refund.SettlementCurrency = settlementCode.String()
	refund.SettlementBalanceTransactionID = strings.TrimSpace(settlementBalanceTransactionID)
	fxGainLoss, fxCurrency, err := calculateRefundFXGainLoss(snapshot, refundMoney, settlementAmountMinor, settlementCode.String())
	if err != nil {
		return err
	}
	refund.FXGainLossMinor = fxGainLoss
	refund.FXGainLossCurrency = fxCurrency
	return nil
}

// optionalRefundMoney validates an optional Money input without serializing it
// through a major-unit float. The zero-value Money is used to represent an
// omitted requested amount at this service boundary.
func optionalRefundMoney(value domainmoney.Money) (domainmoney.Money, string, error) {
	code := strings.TrimSpace(value.Currency().String())
	if code == "" {
		return domainmoney.Money{}, "", nil
	}
	if err := value.Validate(); err != nil {
		return domainmoney.Money{}, "", err
	}
	return value, normalizePaymentCurrency(code), nil
}

func (s *PaymentService) GetRefund(id uint) (*payment.Refund, error) {
	return s.paymentRepo.FindRefundByID(id)
}

func (s *PaymentService) GetOrderRefunds(orderID uint) ([]payment.Refund, error) {
	return s.paymentRepo.FindRefundsByOrderID(orderID)
}

func (s *PaymentService) CreateAdminRefund(refund *payment.Refund, adminUserID uint) error {
	_, err := s.CreateAdminRefundWithIdempotency(refund, adminUserID, "", "")
	return err
}

func (s *PaymentService) CreateAdminRefundWithIdempotency(
	refund *payment.Refund,
	adminUserID uint,
	idempotencyKey string,
	requestHash string,
) (*payment.Refund, error) {
	if err := validateAdminRefundInput(refund); err != nil {
		return nil, err
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	requestHash = strings.TrimSpace(requestHash)
	if idempotencyKey != "" && requestHash == "" {
		return nil, ErrPaymentRefundIdempotencyHashRequired
	}

	var result *payment.Refund
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		var idempotencyRecord *payment.RefundIdempotency
		if idempotencyKey != "" {
			if repos.RefundIdempotency == nil {
				return ErrPaymentRefundIdempotencyUnavailable
			}
			record := &payment.RefundIdempotency{
				AdminUserID:    adminUserID,
				Scope:          adminRefundIdempotencyScope,
				IdempotencyKey: idempotencyKey,
				RequestHash:    requestHash,
			}
			claimed, err := repos.RefundIdempotency.TryCreate(record)
			if err != nil {
				return err
			}
			if !claimed {
				existing, err := repos.RefundIdempotency.FindByScopeKey(adminRefundIdempotencyScope, idempotencyKey)
				if err != nil {
					return err
				}
				if existing.RequestHash != requestHash {
					return ErrPaymentRefundIdempotencyConflict
				}
				if existing.RefundID == nil || *existing.RefundID == 0 {
					return ErrPaymentRefundIdempotencyInProgress
				}
				existingRefund, err := repos.Payment.FindRefundByID(*existing.RefundID)
				if err != nil {
					return err
				}
				result = existingRefund
				return nil
			}
			idempotencyRecord = record
		}

		if err := createAdminRefundInTx(repos, refund, adminUserID); err != nil {
			return err
		}
		if idempotencyRecord != nil {
			if err := repos.RefundIdempotency.BindRefundID(idempotencyRecord.ID, refund.ID); err != nil {
				return err
			}
		}
		result = refund
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
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
	if refund.AmountMinor <= 0 && len(refund.LineItems) == 0 {
		return errors.New("amount must be greater than zero")
	}
	return nil
}

func isRefundableGatewayTransactionStatus(status string) bool {
	return status == "completed" || status == payment.TransactionStatusDuplicatePaid
}

func isDuplicatePaidRefund(refund *payment.Refund) bool {
	return refund != nil && refund.Reason == duplicatePaidRefundReason
}

func createAdminRefundInTx(repos repository.TxRepositories, refund *payment.Refund, adminUserID uint) error {
	if err := validateAdminRefundInput(refund); err != nil {
		return err
	}
	// Serialize refund creation with fulfillment on the order row first. The
	// fulfillment workflow takes this same lock before checking pending refunds;
	// locking the transaction first would allow a shipment to pass the check
	// while this refund intent is still being created.
	o, err := repos.Order.FindByIDForUpdateWithItems(refund.OrderID)
	if err != nil {
		return normalizeOrderError(err)
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
	if !isRefundableGatewayTransactionStatus(transaction.Status) {
		return errors.New("transaction is not refundable")
	}

	if o.PaymentStatus != "paid" {
		return errors.New("order is not paid")
	}
	if o.Status == "refunded" {
		return errors.New("order is already refunded")
	}

	requestedAmount, err := domainmoney.New(refund.AmountMinor, o.Currency)
	if err != nil {
		return fmt.Errorf("parse refund amount: %w", err)
	}
	requestedSubtotalAmount := requestedAmount
	linePricingIncludesCoupon := false
	if len(refund.LineItems) > 0 {
		lineItems, totals, err := buildRefundLineItems(repos, o, refund.LineItems)
		if err != nil {
			return err
		}
		if requestedAmount.AmountMinor() > 0 {
			equal, compareErr := refundAmountsEqualInMinorUnits(requestedAmount, totals.LineTotalAmount)
			if compareErr != nil {
				return fmt.Errorf("compare refund amount with selected line item total: %w", compareErr)
			}
			if !equal {
				return fmt.Errorf(
					"refund amount %s does not match selected line item total %s",
					formatRefundMoney(requestedAmount),
					formatRefundMoney(totals.LineTotalAmount),
				)
			}
		}
		refund.LineItems = lineItems
		requestedAmount = totals.LineTotalAmount
		requestedSubtotalAmount = totals.LineSubtotalAmount
		// A pricing snapshot is the source of truth for order-level discount
		// allocation. Once a selected line carries a coupon allocation, do not
		// recalculate the coupon against the post-refund subtotal below.
		for _, lineItem := range lineItems {
			for _, orderItem := range o.Items {
				if orderItem.ID != lineItem.OrderItemID {
					continue
				}
				lineSnapshot, snapshotErr := parseOrderItemPricingSnapshot(orderItem)
				if snapshotErr != nil {
					return snapshotErr
				}
				for _, allocation := range lineSnapshot.DiscountAllocations() {
					if allocation.Kind() == domainpricing.DiscountKindCoupon && allocation.Amount().AmountMinor() > 0 {
						linePricingIncludesCoupon = true
						break
					}
				}
				break
			}
			if linePricingIncludesCoupon {
				break
			}
		}
	}

	var adjustment refundPromotionAdjustment
	if linePricingIncludesCoupon {
		originalCouponDiscount := zeroRefundMoney(o.Currency)
		if _, couponDiscount, present, snapshotErr := readOrderPricingRefundBaseline(o); snapshotErr != nil {
			return snapshotErr
		} else if present {
			originalCouponDiscount = couponDiscount
		}
		adjustment, err = calculateRefundPromotionAdjustmentFromPersistedPricing(
			o,
			requestedAmount,
			requestedSubtotalAmount,
			originalCouponDiscount,
		)
	} else {
		adjustment, err = calculateRefundPromotionAdjustment(repos, o, requestedAmount, requestedSubtotalAmount)
	}
	if err != nil {
		return err
	}
	fxSnapshot, _, err := ensureRefundFXSnapshot(refund, o, transaction.Currency)
	if err != nil {
		return err
	}
	reservedAmountMinor, err := repos.Payment.SumRefundAmountMinorByTransactionID(transaction.ID, "pending", "completed")
	if err != nil {
		return err
	}
	transactionMoney, err := transaction.AmountMoney()
	if err != nil {
		return err
	}
	reservedGatewayMoney, err := domainmoney.New(reservedAmountMinor, transaction.Currency)
	if err != nil {
		return err
	}
	split, err := calculateRefundPaymentSplit(
		adjustment.NetAmount,
		transactionMoney,
		reservedGatewayMoney,
	)
	if err != nil {
		return err
	}
	if err := validateHistoricalRefundFXCap(fxSnapshot, transaction, split.GatewayAmount, reservedGatewayMoney); err != nil {
		return err
	}

	// Persist the exact pricing-pipeline result in minor units. Refund caps,
	// idempotency, and gateway settlement all compare these canonical values.
	refund.RequestedAmountMinor = adjustment.RequestedAmount.AmountMinor()
	refund.AmountMinor = split.GatewayAmount.AmountMinor()
	refund.DiscountClawbackAmountMinor = adjustment.DiscountClawbackAmount.AmountMinor()
	refund.Currency = transaction.Currency
	refund.CalculationSnapshot = adjustment.CalculationSnapshot
	refund.FXSnapshotData = currencydomain.OrderFXSnapshotJSON(fxSnapshot)
	refund.Status = "pending"
	refund.RefundID = nil
	refund.GatewayResponse = ""
	refund.CompletedAt = nil
	refund.RefundedBy = adminUserID

	if err := repos.Payment.CreateRefund(refund); err != nil {
		return err
	}
	// A pending gateway refund is a financial hold. Keep it on the order until
	// the pending intent is completed or failed so warehouse fulfillment cannot
	// race the provider's asynchronous money movement.
	if err := repos.Order.SetFulfillmentHold(refund.OrderID, true); err != nil {
		return err
	}
	return enqueuePaymentRefundPendingOutboxEvent(
		repos.Outbox,
		refund,
		transaction.Currency,
		transaction.PaymentMethod,
		"",
		0,
		time.Now().UTC(),
	)
}

// releaseRefundPendingHoldIfClear drops the refund-created hold only after no
// pending intents remain. Dispute and payment-review projections are
// deliberately preserved through their terminal order statuses.
func releaseRefundPendingHoldIfClear(repos repository.TxRepositories, orderID uint) error {
	if repos.Order == nil || repos.Payment == nil {
		return nil
	}
	pending, err := repos.Payment.HasPendingRefundByOrderID(orderID)
	if err != nil || pending {
		return err
	}
	orderRecord, err := repos.Order.FindByIDForUpdate(orderID)
	if err != nil {
		return err
	}
	if orderRecord.Status == "disputed" || orderRecord.Status == "needs_review" ||
		orderRecord.Status == "shipped" || orderRecord.Status == "completed" ||
		orderRecord.ShippingStatus == "shipped" || orderRecord.ShippingStatus == "delivered" {
		return nil
	}
	// Disputes and payment reviews project their own terminal status onto the
	// order ("disputed"/"needs_review") before setting this shared flag. The
	// status guard above therefore avoids a second set of cross-table reads in
	// this cleanup path and remains compatible with older schemas.
	return repos.Order.SetFulfillmentHold(orderID, false)
}

func refundCanRestockPhysicalItems(orderRecord *order.Order) bool {
	if orderRecord == nil {
		return false
	}
	// A gateway refund does not prove that goods came back. Once a parcel has
	// entered physical fulfillment, inventory is restored only by the returns
	// receiving workflow after warehouse inspection.
	return orderRecord.Status != "shipped" && orderRecord.Status != "completed" &&
		orderRecord.ShippingStatus != "shipped" && orderRecord.ShippingStatus != "delivered" &&
		(orderRecord.DeliveredAt == nil || orderRecord.DeliveredAt.IsZero()) &&
		(orderRecord.ShippedAt == nil || orderRecord.ShippedAt.IsZero())
}

func calculateRefundPaymentSplit(
	totalRefund domainmoney.Money,
	transactionTotal domainmoney.Money,
	reservedGateway domainmoney.Money,
) (refundPaymentSplit, error) {
	if err := totalRefund.Validate(); err != nil {
		return refundPaymentSplit{}, fmt.Errorf("invalid refund amount: %w", err)
	}
	if err := transactionTotal.Validate(); err != nil {
		return refundPaymentSplit{}, fmt.Errorf("invalid transaction amount: %w", err)
	}
	if err := reservedGateway.Validate(); err != nil {
		return refundPaymentSplit{}, fmt.Errorf("invalid reserved gateway amount: %w", err)
	}
	transactionCurrency := transactionTotal.Currency().String()
	if totalRefund.Currency() != transactionTotal.Currency() || reservedGateway.Currency() != transactionTotal.Currency() {
		return refundPaymentSplit{}, domainmoney.ErrCurrencyMismatch
	}
	if reservedGateway.AmountMinor() < 0 {
		var err error
		reservedGateway, err = domainmoney.New(0, transactionCurrency)
		if err != nil {
			return refundPaymentSplit{}, err
		}
	}

	remainingGateway, err := transactionTotal.Subtract(reservedGateway)
	if err != nil {
		return refundPaymentSplit{}, err
	}
	if remainingGateway.AmountMinor() < 0 {
		remainingGateway, err = domainmoney.New(0, transactionCurrency)
		if err != nil {
			return refundPaymentSplit{}, err
		}
	}
	if totalRefund.AmountMinor() > remainingGateway.AmountMinor() {
		return refundPaymentSplit{}, fmt.Errorf(
			"refund amount %s exceeds refundable amount %s",
			formatRefundMoney(totalRefund),
			formatRefundMoney(remainingGateway),
		)
	}
	return refundPaymentSplit{
		GatewayAmount:  totalRefund,
		TotalAvailable: remainingGateway,
	}, nil
}

func formatRefundMoney(value domainmoney.Money) string {
	formatted, err := value.FormatMajor()
	if err != nil {
		return "<invalid>"
	}
	return formatted
}

func buildRefundLineItems(repos repository.TxRepositories, o *order.Order, requestedItems []payment.RefundLineItem) ([]payment.RefundLineItem, refundLineItemTotals, error) {
	if len(requestedItems) == 0 {
		return nil, refundLineItemTotals{}, errors.New("refund line items are required")
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
	zero, err := domainmoney.New(0, o.Currency)
	if err != nil {
		return nil, refundLineItemTotals{}, fmt.Errorf("refund order currency: %w", err)
	}
	totals := refundLineItemTotals{
		LineSubtotalAmount: zero,
		LineTaxAmount:      zero,
		LineDiscountAmount: zero,
		LineTotalAmount:    zero,
	}
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

		lineItem, allocation, err := buildRefundLineItemSnapshot(
			item,
			alreadyRefunded,
			requested.Quantity,
			requested.Restock,
			o.Currency,
		)
		if err != nil {
			return nil, totals, fmt.Errorf("build refund line item %d: %w", item.ID, err)
		}
		lineItems = append(lineItems, lineItem)
		totals.Quantity += allocation.Quantity
		totals.LineSubtotalAmount, err = totals.LineSubtotalAmount.Add(allocation.LineSubtotalAmount)
		if err != nil {
			return nil, totals, err
		}
		totals.LineTaxAmount, err = totals.LineTaxAmount.Add(allocation.LineTaxAmount)
		if err != nil {
			return nil, totals, err
		}
		totals.LineDiscountAmount, err = totals.LineDiscountAmount.Add(allocation.LineDiscountAmount)
		if err != nil {
			return nil, totals, err
		}
		totals.LineTotalAmount, err = totals.LineTotalAmount.Add(allocation.LineTotalAmount)
		if err != nil {
			return nil, totals, err
		}
	}

	if totals.LineTotalAmount.AmountMinor() <= 0 {
		return nil, totals, errors.New("selected line items do not produce a refundable amount")
	}
	return lineItems, totals, nil
}

func buildRefundLineItemSnapshot(
	item order.OrderItem,
	alreadyRefunded int,
	quantity int,
	restock bool,
	currencyCode string,
) (payment.RefundLineItem, refundLineItemTotals, error) {
	currencyCode = currencydomain.NormalizeCode(currencyCode)
	if !currencydomain.IsCatalogCode(currencyCode) {
		return payment.RefundLineItem{}, refundLineItemTotals{}, fmt.Errorf("invalid refund currency %s", currencyCode)
	}
	snapshot, err := parseOrderItemPricingSnapshot(item)
	if err != nil {
		return payment.RefundLineItem{}, refundLineItemTotals{}, err
	}
	if snapshot.ProductID() != item.ProductID ||
		snapshot.VariantID() != orderItemVariantID(item) ||
		snapshot.Quantity() != item.Quantity {
		return payment.RefundLineItem{}, refundLineItemTotals{}, fmt.Errorf(
			"%w: snapshot identity does not match order item",
			errInvalidOrderItemPricingSnapshot,
		)
	}
	if snapshot.UnitPrice().Currency().String() != currencyCode {
		return payment.RefundLineItem{}, refundLineItemTotals{}, fmt.Errorf("%w: snapshot currency does not match order currency", errInvalidOrderItemPricingSnapshot)
	}
	lineSubtotal, lineDiscount, lineTotal, err := allocateRefundLineAmounts(snapshot, alreadyRefunded, quantity)
	if err != nil {
		return payment.RefundLineItem{}, refundLineItemTotals{}, err
	}
	itemTax, err := item.TaxAmountMoney()
	if pricingSnapshotIncludesTax(item.PricingSnapshotData) {
		if snapshotTax := snapshot.Tax(); snapshotTax.Validate() == nil && snapshotTax.AmountMinor() > 0 {
			itemTax = snapshotTax
		}
	}
	if err != nil {
		return payment.RefundLineItem{}, refundLineItemTotals{}, fmt.Errorf("parse order item tax amount: %w", err)
	}
	lineTax, err := allocateRefundMoneyRange(itemTax, alreadyRefunded, quantity, item.Quantity)
	if err != nil {
		return payment.RefundLineItem{}, refundLineItemTotals{}, err
	}
	lineTotal, err = lineTotal.Add(lineTax)
	if err != nil {
		return payment.RefundLineItem{}, refundLineItemTotals{}, err
	}
	lineItem := payment.RefundLineItem{
		OrderID:           item.OrderID,
		OrderItemID:       item.ID,
		ProductID:         item.ProductID,
		VariantID:         item.VariantID,
		ProductName:       item.ProductName,
		SKU:               item.SKU,
		Quantity:          quantity,
		Currency:          currencyCode,
		UnitPriceMinor:    snapshot.UnitPrice().AmountMinor(),
		LineSubtotalMinor: lineSubtotal.AmountMinor(),
		LineTaxMinor:      lineTax.AmountMinor(),
		LineDiscountMinor: lineDiscount.AmountMinor(),
		LineTotalMinor:    lineTotal.AmountMinor(),
		Restock:           restock,
	}
	allocation := refundLineItemTotals{
		Quantity:           quantity,
		LineSubtotalAmount: lineSubtotal,
		LineTaxAmount:      lineTax,
		LineDiscountAmount: lineDiscount,
		LineTotalAmount:    lineTotal,
	}
	return lineItem, allocation, nil
}

func pricingSnapshotIncludesTax(raw []byte) bool {
	if len(raw) == 0 || string(raw) == "{}" {
		return false
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return false
	}
	_, ok := fields["tax_minor"]
	return ok
}

var errInvalidOrderItemPricingSnapshot = errors.New("invalid order item pricing snapshot")

func parseOrderItemPricingSnapshot(item order.OrderItem) (domainpricing.LineSnapshot, error) {
	if len(item.PricingSnapshotData) == 0 || string(item.PricingSnapshotData) == "{}" {
		return domainpricing.LineSnapshot{}, fmt.Errorf("%w: pricing snapshot is required", errInvalidOrderItemPricingSnapshot)
	}
	snapshot, err := domainpricing.ParseLineSnapshot(item.PricingSnapshotData)
	if err != nil {
		return domainpricing.LineSnapshot{}, fmt.Errorf("%w: %v", errInvalidOrderItemPricingSnapshot, err)
	}
	return snapshot, nil
}

func allocateRefundLineAmounts(
	snapshot domainpricing.LineSnapshot,
	alreadyRefunded int,
	quantity int,
) (domainmoney.Money, domainmoney.Money, domainmoney.Money, error) {
	if quantity <= 0 || snapshot.Quantity() <= 0 || alreadyRefunded < 0 ||
		alreadyRefunded > snapshot.Quantity() || quantity > snapshot.Quantity()-alreadyRefunded {
		return domainmoney.Money{}, domainmoney.Money{}, domainmoney.Money{}, errors.New("refund quantity is inconsistent with pricing snapshot")
	}
	base, err := snapshot.UnitPrice().MultiplyInt(int64(quantity))
	if err != nil {
		return domainmoney.Money{}, domainmoney.Money{}, domainmoney.Money{}, fmt.Errorf("allocate base subtotal: %w", err)
	}
	discount, err := allocateRefundMoneyRange(
		snapshot.DiscountTotal(),
		alreadyRefunded,
		quantity,
		snapshot.Quantity(),
	)
	if err != nil {
		return domainmoney.Money{}, domainmoney.Money{}, domainmoney.Money{}, fmt.Errorf("allocate discount: %w", err)
	}
	net, err := base.Subtract(discount)
	if err != nil {
		return domainmoney.Money{}, domainmoney.Money{}, domainmoney.Money{}, fmt.Errorf("calculate net subtotal: %w", err)
	}
	return base, discount, net, nil
}

func allocateRefundMoneyRange(
	amount domainmoney.Money,
	alreadyAllocated int,
	quantity int,
	totalQuantity int,
) (domainmoney.Money, error) {
	if err := amount.Validate(); err != nil {
		return domainmoney.Money{}, err
	}
	if alreadyAllocated < 0 || quantity < 0 || totalQuantity <= 0 ||
		alreadyAllocated > totalQuantity || quantity > totalQuantity-alreadyAllocated {
		return domainmoney.Money{}, errors.New("refund quantity is inconsistent with source quantity")
	}
	allocatedBefore, err := amount.MultiplyRatio(int64(alreadyAllocated), int64(totalQuantity))
	if err != nil {
		return domainmoney.Money{}, err
	}
	allocatedThroughCurrent, err := amount.MultiplyRatio(int64(alreadyAllocated+quantity), int64(totalQuantity))
	if err != nil {
		return domainmoney.Money{}, err
	}
	return allocatedThroughCurrent.Subtract(allocatedBefore)
}

func refundAmountsEqualInMinorUnits(left, right domainmoney.Money) (bool, error) {
	if err := validateRefundMoneyOperand(left, "left"); err != nil {
		return false, err
	}
	if err := validateRefundMoneyOperand(right, "right"); err != nil {
		return false, err
	}
	if left.Currency() != right.Currency() {
		return false, domainmoney.ErrCurrencyMismatch
	}
	return left.AmountMinor() == right.AmountMinor(), nil
}

func addRefundAmounts(left, right domainmoney.Money) (domainmoney.Money, error) {
	if err := validateRefundMoneyOperand(left, "left"); err != nil {
		return domainmoney.Money{}, err
	}
	if err := validateRefundMoneyOperand(right, "right"); err != nil {
		return domainmoney.Money{}, err
	}
	return left.Add(right)
}

func subtractRefundAmounts(left, right domainmoney.Money) (domainmoney.Money, error) {
	if err := validateRefundMoneyOperand(left, "left"); err != nil {
		return domainmoney.Money{}, err
	}
	if err := validateRefundMoneyOperand(right, "right"); err != nil {
		return domainmoney.Money{}, err
	}
	return left.Subtract(right)
}

func refundAmountAtLeastInMinorUnits(total, target domainmoney.Money) (bool, error) {
	if err := validateRefundMoneyOperand(total, "total"); err != nil {
		return false, err
	}
	if err := validateRefundMoneyOperand(target, "target"); err != nil {
		return false, err
	}
	if total.Currency() != target.Currency() {
		return false, domainmoney.ErrCurrencyMismatch
	}
	return total.AmountMinor() >= target.AmountMinor(), nil
}

func refundAmountExceedsInMinorUnits(left, right domainmoney.Money) (bool, error) {
	if err := validateRefundMoneyOperand(left, "left"); err != nil {
		return false, err
	}
	if err := validateRefundMoneyOperand(right, "right"); err != nil {
		return false, err
	}
	if left.Currency() != right.Currency() {
		return false, domainmoney.ErrCurrencyMismatch
	}
	return left.AmountMinor() > right.AmountMinor(), nil
}

func validateRefundMoneyOperand(value domainmoney.Money, name string) error {
	if err := value.Validate(); err != nil {
		return fmt.Errorf("invalid %s refund amount: %w", name, err)
	}
	return nil
}

func restoreRefundLineItemStock(repos repository.TxRepositories, orderRecord *order.Order, lineItems []payment.RefundLineItem, restockedAt time.Time) ([]uint, error) {
	var affectedProductIDs []uint
	variantItemsMap := make(map[uint]int)
	orderItemsByID := make(map[uint]order.OrderItem)
	if orderRecord != nil {
		for _, orderItem := range orderRecord.Items {
			orderItemsByID[orderItem.ID] = orderItem
		}
	}
	for _, item := range lineItems {
		if !item.Restock || item.Quantity <= 0 || item.VariantID == nil {
			continue
		}

		claimed, err := repos.Payment.MarkRefundLineItemRestocked(item.ID, restockedAt)
		if err != nil {
			return nil, err
		}
		if !claimed {
			continue
		}
		variantItemsMap[*item.VariantID] += item.Quantity
		if orderItem, ok := orderItemsByID[item.OrderItemID]; ok && len(orderItem.ConfigurationSnapshotData) > 0 && string(orderItem.ConfigurationSnapshotData) != "{}" {
			var configuration ProductConfigurationSnapshot
			if err := json.Unmarshal(orderItem.ConfigurationSnapshotData, &configuration); err != nil {
				return nil, fmt.Errorf("[CRITICAL] Failed to parse configuration snapshot for refunded order item %d: %w", item.OrderItemID, err)
			}
			for _, allocation := range configuration.InventoryAllocations {
				if allocation.VariantID == 0 || allocation.Quantity <= 0 {
					return nil, fmt.Errorf("[CRITICAL] Invalid component inventory allocation for refunded order item %d", item.OrderItemID)
				}
				variantItemsMap[allocation.VariantID] += allocation.Quantity * item.Quantity
			}
		}
	}
	if len(variantItemsMap) > 0 {
		productIDs, err := repos.Product.IncrementVariantStocks(variantItemsMap)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Failed to restore stock for refunded variants: %w", err)
		}
		affectedProductIDs = append(affectedProductIDs, productIDs...)
	}
	return affectedProductIDs, nil
}

func (s *PaymentService) RecordVerifiedGatewayRefund(input VerifiedGatewayRefundInput) error {
	input.RefundID = strings.TrimSpace(input.RefundID)
	// A gateway may deliver the same refund webhook more than once. Resolve
	// the provider refund ID before validating the rest of the payload. A
	// previously processed webhook may have completed the local refund while
	// leaving an execution attempt stale, so reconcile that attempt instead of
	// returning early and preserving the stale state.
	if input.RefundID != "" {
		if existingRefund, err := s.paymentRepo.FindRefundByRefundID(input.RefundID); err == nil {
			return s.reconcileExistingGatewayRefund(input, existingRefund)
		} else if !repository.IsRecordNotFound(err) {
			return err
		}
	}

	if input.Provider == "" {
		return errors.New("provider is required")
	}
	if input.TransactionID == "" {
		return errors.New("transaction_id is required")
	}
	if input.RefundID == "" {
		return errors.New("refund_id is required")
	}
	if err := input.ProviderRefundAmount.Validate(); err != nil {
		return err
	}
	providerRefundMoney, providerCurrency, amountErr := optionalRefundMoney(input.ProviderRefundAmount)
	if amountErr != nil {
		return amountErr
	}
	if providerRefundMoney.AmountMinor() <= 0 || providerCurrency == "" {
		return errors.New("provider refund amount must be greater than zero")
	}
	requestedRefundMoney, requestedCurrency, amountErr := optionalRefundMoney(input.RequestedRefundAmount)
	if amountErr != nil {
		return amountErr
	}
	if requestedCurrency != "" && requestedCurrency != providerCurrency {
		return fmt.Errorf("requested refund currency %s does not match provider refund currency %s", requestedCurrency, providerCurrency)
	}

	var affectedProductIDs []uint
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		// Resolve the transaction without locking it so we can acquire the order
		// serialization point first. Admin refund creation and fulfillment use
		// this same order-first lock ordering.
		transaction, err := repos.Payment.FindTransactionByTransactionID(input.TransactionID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return errors.New("transaction not found")
			}
			return err
		}
		o, err := repos.Order.FindByIDForUpdate(transaction.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		transaction, err = repos.Payment.FindTransactionByTransactionIDForUpdate(input.TransactionID)
		if err != nil {
			return err
		}
		// Recheck after locking the transaction so concurrent deliveries
		// serialize behind the first commit before deciding to create a refund.
		if _, err := repos.Payment.FindRefundByRefundID(input.RefundID); err == nil {
			return nil
		} else if !repository.IsRecordNotFound(err) {
			return err
		}
		if transaction.Status == "refunded" {
			return errors.New("transaction is already fully refunded")
		}
		if !isRefundableGatewayTransactionStatus(transaction.Status) {
			return errors.New("transaction is not refundable")
		}
		if providerCurrency != "" && transaction.Currency != "" && !strings.EqualFold(providerCurrency, transaction.Currency) {
			return fmt.Errorf("refund currency %s does not match transaction currency %s", providerCurrency, transaction.Currency)
		}

		if input.OrderNumber != "" && o.OrderNumber != input.OrderNumber {
			return errors.New("refund order_number does not match transaction order")
		}
		duplicatePaidRefund := transaction.Status == payment.TransactionStatusDuplicatePaid
		if o.PaymentStatus == "refunded" && !duplicatePaidRefund {
			return errors.New("order is already refunded")
		}
		if o.PaymentStatus != "paid" && !duplicatePaidRefund {
			return errors.New("order is not paid")
		}

		fxSnapshot, _, err := ensureRefundFXSnapshot(&payment.Refund{}, o, transaction.Currency)
		if err != nil {
			return err
		}
		reservedAmountMinor, err := repos.Payment.SumRefundAmountMinorByTransactionID(transaction.ID, "pending", "completed")
		if err != nil {
			return err
		}

		now := time.Now()
		refundID := input.RefundID
		pendingRefund, err := repos.Payment.FindPendingRefundByTransactionAndAmount(transaction.ID, input.ProviderRefundAmount)
		if repository.IsRecordNotFound(err) {
			pendingRefund, err = repos.Payment.FindPendingRefundByTransactionIDForUpdate(transaction.ID)
		}
		if err != nil && !repository.IsRecordNotFound(err) {
			return err
		}
		reservedBeforeCurrentMoney, moneyErr := domainmoney.New(reservedAmountMinor, transaction.Currency)
		if moneyErr != nil {
			return moneyErr
		}
		if pendingRefund != nil {
			pendingMoney, moneyErr := pendingRefund.AmountMoney()
			if moneyErr != nil {
				return moneyErr
			}
			if !strings.EqualFold(pendingMoney.Currency().String(), transaction.Currency) {
				return fmt.Errorf("pending refund currency %s does not match transaction currency %s", pendingMoney.Currency(), transaction.Currency)
			}
			reservedBeforeCurrentMoney, moneyErr = subtractRefundAmounts(reservedBeforeCurrentMoney, pendingMoney)
			if moneyErr != nil {
				return moneyErr
			}
		}
		transactionMoney, moneyErr := transaction.AmountMoney()
		if moneyErr != nil {
			return moneyErr
		}
		remainingAmountMoney, remainingErr := subtractRefundAmounts(transactionMoney, reservedBeforeCurrentMoney)
		if remainingErr != nil {
			return remainingErr
		}
		exceeds, compareErr := refundAmountExceedsInMinorUnits(providerRefundMoney, remainingAmountMoney)
		if compareErr != nil {
			return compareErr
		}
		if exceeds {
			return fmt.Errorf(
				"provider refund amount %s exceeds refundable amount %s",
				formatRefundMoney(providerRefundMoney),
				formatRefundMoney(remainingAmountMoney),
			)
		}
		if err := validateHistoricalRefundFXCap(fxSnapshot, transaction, providerRefundMoney, reservedBeforeCurrentMoney); err != nil {
			return err
		}
		if pendingRefund != nil {
			wasCompleted := pendingRefund.Status == "completed"
			requestedMoney := requestedRefundMoney
			if requestedRefundMoney.AmountMinor() <= 0 {
				requestedMoney, err = domainmoney.New(0, transaction.Currency)
				if err != nil {
					return err
				}
			}
			pendingRequestedMoney, moneyErr := pendingRefund.RequestedAmountMoney()
			if moneyErr != nil {
				return moneyErr
			}
			requestedMismatch, compareErr := refundAmountsEqualInMinorUnits(requestedMoney, pendingRequestedMoney)
			if compareErr != nil {
				return compareErr
			}
			if requestedRefundMoney.AmountMinor() > 0 && !requestedMismatch {
				return fmt.Errorf(
					"requested refund amount %s does not match local requested refund amount %s",
					formatRefundMoney(requestedRefundMoney),
					formatRefundMoney(pendingRequestedMoney),
				)
			}
			if err := prepareRefundLoyaltySettlementInTx(repos, o, pendingRefund); err != nil {
				return err
			}
			pendingAmountMoney, moneyErr := pendingRefund.AmountMoney()
			if moneyErr != nil {
				return moneyErr
			}
			amountMismatch, compareErr := refundAmountsEqualInMinorUnits(providerRefundMoney, pendingAmountMoney)
			if compareErr != nil {
				return compareErr
			}
			if !amountMismatch {
				return fmt.Errorf(
					"provider refund amount %s does not match local net refund amount %s",
					formatRefundMoney(providerRefundMoney),
					formatRefundMoney(pendingAmountMoney),
				)
			}
			pendingRefund.Status = "completed"
			pendingRefund.RefundID = &refundID
			pendingRefund.GatewayResponse = input.GatewayResponse
			pendingRefund.CompletedAt = &now
			pendingRefund.FXSnapshotData = currencydomain.OrderFXSnapshotJSON(fxSnapshot)
			if err := applyRefundSettlementFacts(
				pendingRefund,
				fxSnapshot,
				providerRefundMoney,
				input.SettlementAmountMinor,
				input.SettlementCurrency,
				input.SettlementBalanceTransactionID,
			); err != nil {
				return err
			}
			if err := repos.Payment.UpdateRefund(pendingRefund); err != nil {
				return err
			}
			if err := finalizeRefundLoyaltySettlementInTx(repos, o, pendingRefund); err != nil {
				return err
			}
			if !wasCompleted && refundCanRestockPhysicalItems(o) {
				productIDs, err := restoreRefundLineItemStock(repos, o, pendingRefund.LineItems, now)
				if err != nil {
					return err
				}
				affectedProductIDs = append(affectedProductIDs, productIDs...)
				if err := s.enqueueProductCacheInvalidationInTx(repos, productIDs, "refund stock restored"); err != nil {
					return err
				}
			}

			// A timeout can leave the local execution row marked failed even
			// though the provider completed the refund. Reconcile that row from
			// the verified webhook while the refund and linked after-sales case
			// are still in the same transaction.
			execution, err := markRefundExecutionSucceededInTx(
				repos,
				pendingRefund,
				refundID,
				input.ProviderStatus,
				input.GatewayResponse,
				now,
			)
			if err != nil {
				return err
			}
			updatedBy := uint(0)
			if execution != nil {
				updatedBy = execution.RequestedByID
			}
			if err := completeLinkedAfterSalesCaseInTx(repos, pendingRefund, updatedBy); err != nil {
				return err
			}
			if err := enqueuePaymentRefundCompletedOutboxEvent(
				repos.Outbox,
				pendingRefund,
				transaction.Currency,
				input.Provider,
				input.RefundID,
				func() string {
					if execution == nil {
						return ""
					}
					return execution.Status
				}(),
				now,
				o.OrderNumber,
			); err != nil {
				return err
			}
			if err := releaseRefundPendingHoldIfClear(repos, pendingRefund.OrderID); err != nil {
				return err
			}
		} else {
			requestedMoney := requestedRefundMoney
			if requestedMoney.AmountMinor() <= 0 {
				requestedMoney = providerRefundMoney
			}
			refund := &payment.Refund{
				OrderID:              transaction.OrderID,
				TransactionID:        transaction.ID,
				Currency:             transaction.Currency,
				RefundID:             nil,
				AmountMinor:          providerRefundMoney.AmountMinor(),
				RequestedAmountMinor: providerRefundMoney.AmountMinor(),
				Status:               "completed",
				Reason: func() string {
					if duplicatePaidRefund {
						return duplicatePaidRefundReason
					}
					return ""
				}(),
				GatewayResponse: input.GatewayResponse,
				FXSnapshotData:  currencydomain.OrderFXSnapshotJSON(fxSnapshot),
				CompletedAt:     &now,
			}
			refund.AmountMinor = providerRefundMoney.AmountMinor()
			refund.RequestedAmountMinor = providerRefundMoney.AmountMinor()
			if requestedMoney.AmountMinor() > 0 {
				transactionMoney, moneyErr := transaction.AmountMoney()
				if moneyErr != nil {
					return moneyErr
				}
				reservedMinor, queryErr := repos.Payment.SumRefundAmountMinorByTransactionID(transaction.ID, "pending", "completed")
				if queryErr != nil {
					return queryErr
				}
				reservedMoney, moneyErr := domainmoney.New(reservedMinor, transaction.Currency)
				if moneyErr != nil {
					return moneyErr
				}
				split, err := calculateRefundPaymentSplit(
					requestedMoney,
					transactionMoney,
					reservedMoney,
				)
				if err != nil {
					return err
				}
				refund.RequestedAmountMinor = requestedMoney.AmountMinor()
				refund.AmountMinor = split.GatewayAmount.AmountMinor()
			}
			if refundMoney, moneyErr := refund.AmountMoney(); moneyErr != nil {
				return moneyErr
			} else if err := applyRefundSettlementFacts(
				refund,
				fxSnapshot,
				refundMoney,
				input.SettlementAmountMinor,
				input.SettlementCurrency,
				input.SettlementBalanceTransactionID,
			); err != nil {
				return err
			}
			if err := repos.Payment.CreateRefund(refund); err != nil {
				return err
			}
			if err := prepareRefundLoyaltySettlementInTx(repos, o, refund); err != nil {
				return err
			}
			refundAmountMoney, moneyErr := refund.AmountMoney()
			if moneyErr != nil {
				return moneyErr
			}
			amountMismatch, compareErr := refundAmountsEqualInMinorUnits(providerRefundMoney, refundAmountMoney)
			if compareErr != nil {
				return compareErr
			}
			if !amountMismatch {
				return fmt.Errorf(
					"provider refund amount %s does not match loyalty-adjusted local net refund amount %s",
					formatRefundMoney(providerRefundMoney),
					formatRefundMoney(refundAmountMoney),
				)
			}
			refund.RefundID = &refundID
			if err := repos.Payment.UpdateRefund(refund); err != nil {
				return err
			}
			if err := finalizeRefundLoyaltySettlementInTx(repos, o, refund); err != nil {
				return err
			}
			if err := enqueuePaymentRefundCompletedOutboxEvent(
				repos.Outbox,
				refund,
				transaction.Currency,
				input.Provider,
				input.RefundID,
				"",
				now,
				o.OrderNumber,
			); err != nil {
				return err
			}
			if err := releaseRefundPendingHoldIfClear(repos, refund.OrderID); err != nil {
				return err
			}
		}

		completedAmountMinor, err := repos.Payment.SumRefundTotalAmountMinorByTransactionID(transaction.ID, "completed")
		if err != nil {
			return err
		}
		transactionMoney, err = transaction.AmountMoney()
		if err != nil {
			return err
		}
		completedMoney, err := domainmoney.New(completedAmountMinor, transaction.Currency)
		if err != nil {
			return err
		}
		transactionFullyRefunded, compareErr := refundAmountAtLeastInMinorUnits(completedMoney, transactionMoney)
		if compareErr != nil {
			return compareErr
		}
		if transactionFullyRefunded && transaction.Status != payment.TransactionStatusDuplicatePaid {
			transaction.Status = "refunded"
			if err := repos.Payment.UpdateTransaction(transaction); err != nil {
				return err
			}
		}

		if !duplicatePaidRefund {
			orderRefundedAmountMinor, err := repos.Payment.SumRefundTotalAmountMinorByOrderID(o.ID, "completed")
			if err != nil {
				return err
			}
			orderRefundedMoney, err := domainmoney.New(orderRefundedAmountMinor, o.Currency)
			if err != nil {
				return err
			}
			orderTotalMoney, err := o.TotalMoney()
			if err != nil {
				return err
			}
			orderFullyRefunded, compareErr := refundAmountAtLeastInMinorUnits(orderRefundedMoney, orderTotalMoney)
			if compareErr != nil {
				return compareErr
			}
			if orderFullyRefunded {
				if err := repos.Order.UpdatePaymentStatus(o.ID, "refunded"); err != nil {
					return err
				}
				if refundCanRestockPhysicalItems(o) {
					if err := repos.Order.UpdateStatus(o.ID, o.Status, "refunded"); err != nil {
						return err
					}
				}
				return enqueueReferralOrderInvalidatedOutboxEvent(
					repos.Outbox,
					o.ID,
					time.Now().UTC(),
					"order fully refunded",
					"refund_webhook",
					input.RefundID,
				)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	s.invalidateProductCacheAfterStockCommit(affectedProductIDs)
	return nil
}

// RecordGatewayRefundFailure persists a provider-declared refund failure. It
// is intentionally separate from RecordVerifiedGatewayRefund: a failed
// provider outcome must never be interpreted as a completed money movement,
// but it must remain visible and retryable.
func (s *PaymentService) RecordGatewayRefundFailure(input VerifiedGatewayRefundInput) error {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.TransactionID = strings.TrimSpace(input.TransactionID)
	input.RefundID = strings.TrimSpace(input.RefundID)
	if input.Provider == "" {
		return errors.New("provider is required")
	}
	if input.TransactionID == "" {
		return errors.New("transaction_id is required")
	}
	if input.RefundID == "" {
		return errors.New("refund_id is required")
	}
	message := strings.TrimSpace(input.ErrorMessage)
	if message == "" {
		message = strings.TrimSpace(input.ProviderStatus)
	}
	if message == "" {
		message = "payment provider reported a refund failure"
	}
	if err := input.ProviderRefundAmount.Validate(); err != nil {
		return err
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		transaction, err := repos.Payment.FindTransactionByTransactionIDForUpdate(input.TransactionID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return errors.New("transaction not found")
			}
			return err
		}
		o, err := repos.Order.FindByIDForUpdate(transaction.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		if input.OrderNumber != "" && strings.TrimSpace(input.OrderNumber) != o.OrderNumber {
			return errors.New("refund order_number does not match transaction order")
		}
		if _, lookupErr := repos.Payment.FindRefundByRefundIDForUpdate(input.RefundID); lookupErr == nil {
			return nil
		} else if !repository.IsRecordNotFound(lookupErr) {
			return lookupErr
		}

		now := time.Now().UTC()
		var refund *payment.Refund
		refund, err = repos.Payment.FindPendingRefundByTransactionIDForUpdate(transaction.ID)
		if repository.IsRecordNotFound(err) {
			// A synchronous gateway failure now marks the local intent failed so
			// its amount is released. If the provider reports that same failure
			// asynchronously, enrich the failed intent instead of creating a
			// duplicate refund row.
			refund, err = repos.Payment.FindFailedRefundByTransactionIDForUpdate(transaction.ID)
		}
		if err == nil {
			if err := releaseRefundLoyaltyReservationInTx(repos, o, refund); err != nil {
				return err
			}
			refund.Status = "failed"
			// Bind the provider refund identifier to the local intent even for a
			// failed outcome. This makes repeated failure webhooks idempotent;
			// retry execution clears this field after retaining it on the execution
			// audit row.
			providerRefundID := input.RefundID
			refund.RefundID = &providerRefundID
			refund.CompletedAt = nil
			refund.GatewayResponse = input.GatewayResponse
			if repos.RefundExecution != nil {
				if execution, executionErr := repos.RefundExecution.FindByRefundIDForUpdate(refund.ID); executionErr == nil {
					if execution.Status != payment.PaymentRefundExecutionStatusSucceeded {
						execution.Status = payment.PaymentRefundExecutionStatusFailed
						execution.ProviderRefundID = input.RefundID
						execution.ProviderStatus = strings.TrimSpace(input.ProviderStatus)
						execution.GatewayResponseJSON = input.GatewayResponse
						execution.ErrorMessage = message
						execution.CompletedAt = &now
						if err := repos.RefundExecution.Update(execution); err != nil {
							return err
						}
					}
				} else if !repository.IsRecordNotFound(executionErr) {
					return executionErr
				}
			}
			if strings.TrimSpace(refund.Reason) == "" {
				refund.Reason = fmt.Sprintf("provider refund failure: %s", message)
			}
			if err := repos.Payment.UpdateRefund(refund); err != nil {
				return err
			}
			if err := releaseRefundPendingHoldIfClear(repos, refund.OrderID); err != nil {
				return err
			}
			attempt := 0
			if repos.RefundExecution != nil {
				if execution, executionErr := repos.RefundExecution.FindByRefundIDForUpdate(refund.ID); executionErr == nil {
					attempt = execution.AttemptCount
				} else if !repository.IsRecordNotFound(executionErr) {
					return executionErr
				}
			}
			return enqueuePaymentRefundFailedOutboxEvent(
				repos.Outbox,
				refund,
				transaction.Currency,
				input.Provider,
				input.RefundID,
				payment.PaymentRefundExecutionStatusFailed,
				message,
				attempt,
				now,
			)
		}
		if !repository.IsRecordNotFound(err) {
			return err
		}

		// A provider can report a failure for a refund request that was created
		// outside this process or before the local intent was persisted. Keep the
		// verified failure visible instead of acknowledging it as success or
		// losing it entirely.
		fxSnapshot, _, snapshotErr := ensureRefundFXSnapshot(&payment.Refund{}, o, transaction.Currency)
		if snapshotErr != nil {
			return snapshotErr
		}
		providerRefundID := input.RefundID
		failedRefund := &payment.Refund{
			OrderID:              transaction.OrderID,
			TransactionID:        transaction.ID,
			Currency:             transaction.Currency,
			RefundID:             &providerRefundID,
			AmountMinor:          input.ProviderRefundAmount.AmountMinor(),
			RequestedAmountMinor: input.ProviderRefundAmount.AmountMinor(),
			Reason:               fmt.Sprintf("provider refund failure: %s", message),
			Status:               "failed",
			GatewayResponse:      input.GatewayResponse,
			FXSnapshotData:       currencydomain.OrderFXSnapshotJSON(fxSnapshot),
		}
		if err := repos.Payment.CreateRefund(failedRefund); err != nil {
			return err
		}
		return enqueuePaymentRefundFailedOutboxEvent(
			repos.Outbox,
			failedRefund,
			transaction.Currency,
			input.Provider,
			input.RefundID,
			"",
			message,
			0,
			now,
		)
	})
}

// reconcileExistingGatewayRefund repairs the local execution/after-sales
// state when a duplicate provider webhook arrives after an earlier webhook
// already completed the refund row. This is intentionally idempotent: it
// does not re-run loyalty, stock, transaction, or order effects.
func (s *PaymentService) reconcileExistingGatewayRefund(
	input VerifiedGatewayRefundInput,
	refund *payment.Refund,
) error {
	if refund == nil || refund.Status != "completed" {
		return nil
	}
	if s == nil || s.txManager == nil {
		return errors.New("payment refund transaction manager is not configured")
	}
	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		lockedRefund, err := repos.Payment.FindRefundByIDForUpdate(refund.ID)
		if err != nil {
			return err
		}
		if lockedRefund.Status != "completed" {
			return nil
		}
		// A replay can contain an expanded balance transaction that was absent
		// from the first webhook delivery. Enrich the already-completed refund
		// without rerunning any monetary side effects. Older rows may not have a
		// historical snapshot; those still retain the provider settlement fact,
		// while FX gain/loss remains zero because it cannot be reconstructed.
		if input.SettlementAmountMinor != 0 && strings.TrimSpace(input.SettlementCurrency) != "" {
			refundMoney, moneyErr := lockedRefund.AmountMoney()
			if moneyErr != nil {
				return moneyErr
			}
			if snapshot, snapshotErr := currencydomain.ParseOrderFXSnapshot(lockedRefund.FXSnapshotData); snapshotErr == nil {
				if err := applyRefundSettlementFacts(
					lockedRefund,
					snapshot,
					refundMoney,
					input.SettlementAmountMinor,
					input.SettlementCurrency,
					input.SettlementBalanceTransactionID,
				); err != nil {
					return err
				}
			} else {
				settlementCode, codeErr := currencydomain.ParseCode(input.SettlementCurrency)
				if codeErr != nil {
					return fmt.Errorf("invalid settlement currency: %w", codeErr)
				}
				settlementAmount := input.SettlementAmountMinor
				if settlementAmount < 0 {
					if settlementAmount == -1<<63 {
						return errors.New("provider settlement amount overflows")
					}
					settlementAmount = -settlementAmount
				}
				lockedRefund.SettlementAmountMinor = settlementAmount
				lockedRefund.SettlementCurrency = settlementCode.String()
				lockedRefund.SettlementBalanceTransactionID = strings.TrimSpace(input.SettlementBalanceTransactionID)
			}
			if err := repos.Payment.UpdateRefund(lockedRefund); err != nil {
				return err
			}
		}
		execution, err := markRefundExecutionSucceededInTx(
			repos,
			lockedRefund,
			input.RefundID,
			input.ProviderStatus,
			input.GatewayResponse,
			time.Now().UTC(),
		)
		if err != nil {
			return err
		}
		updatedBy := uint(0)
		if execution != nil {
			updatedBy = execution.RequestedByID
		}
		return completeLinkedAfterSalesCaseInTx(repos, lockedRefund, updatedBy)
	})
}
