package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/pricing"

	"gorm.io/gorm"
)

const (
	defaultPricingSnapshotAuditBatchSize  = 200
	defaultPricingSnapshotAuditIssueLimit = 1000
)

// OrderPricingSnapshotAuditOptions controls a read-only historical snapshot
// audit. The audit never writes or locks order rows.
type OrderPricingSnapshotAuditOptions struct {
	BatchSize   int
	MaxOrders   int
	FromOrderID uint
	ToOrderID   uint
	MaxIssues   int
}

// OrderPricingSnapshotAuditIssue identifies one historical row that needs
// review. Details intentionally contain facts and field names only; raw
// customer/order payloads are never copied into the report.
type OrderPricingSnapshotAuditIssue struct {
	Scope       string `json:"scope"`
	Code        string `json:"code"`
	OrderID     uint   `json:"order_id"`
	OrderNumber string `json:"order_number,omitempty"`
	ItemID      uint   `json:"item_id,omitempty"`
	Detail      string `json:"detail"`
}

// OrderPricingSnapshotAuditReport is safe to serialize as JSON and is useful
// both for deployment preflight and for a production read-only audit run.
type OrderPricingSnapshotAuditReport struct {
	GeneratedAt time.Time                        `json:"generated_at"`
	Options     OrderPricingSnapshotAuditOptions `json:"options"`

	OrdersScanned            int                              `json:"orders_scanned"`
	OrdersWithValidSnapshot  int                              `json:"orders_with_valid_snapshot"`
	OrdersMissingSnapshot    int                              `json:"orders_missing_snapshot"`
	OrdersInvalidSnapshot    int                              `json:"orders_invalid_snapshot"`
	OrdersMismatchedSnapshot int                              `json:"orders_mismatched_snapshot"`
	ItemsScanned             int                              `json:"items_scanned"`
	ItemsWithValidSnapshot   int                              `json:"items_with_valid_snapshot"`
	ItemsMissingSnapshot     int                              `json:"items_missing_snapshot"`
	ItemsInvalidSnapshot     int                              `json:"items_invalid_snapshot"`
	ItemsMismatchedSnapshot  int                              `json:"items_mismatched_snapshot"`
	TotalIssues              int                              `json:"total_issues"`
	IssuesTruncated          bool                             `json:"issues_truncated"`
	OrdersTruncated          bool                             `json:"orders_truncated"`
	Issues                   []OrderPricingSnapshotAuditIssue `json:"issues"`
}

// OrderPricingSnapshotAuditService performs a bounded, read-only audit of
// order and order-item pricing snapshots. It is deliberately separate from
// refund/repair code: findings must be reviewed before any remediation.
type OrderPricingSnapshotAuditService struct {
	db *gorm.DB
}

func NewOrderPricingSnapshotAuditService(db *gorm.DB) *OrderPricingSnapshotAuditService {
	return &OrderPricingSnapshotAuditService{db: db}
}

func (s *OrderPricingSnapshotAuditService) Audit(ctx context.Context, options OrderPricingSnapshotAuditOptions) (OrderPricingSnapshotAuditReport, error) {
	if s == nil || s.db == nil {
		return OrderPricingSnapshotAuditReport{}, errors.New("pricing snapshot audit database is required")
	}
	options = normalizePricingSnapshotAuditOptions(options)
	report := OrderPricingSnapshotAuditReport{
		GeneratedAt: time.Now().UTC(),
		Options:     options,
		Issues:      make([]OrderPricingSnapshotAuditIssue, 0),
	}

	lastID := options.FromOrderID
	for {
		batchSize := options.BatchSize
		if options.MaxOrders > 0 {
			remaining := options.MaxOrders - report.OrdersScanned
			if remaining <= 0 {
				report.OrdersTruncated = true
				break
			}
			if remaining < batchSize {
				batchSize = remaining
			}
		}

		var orders []order.Order
		query := s.db.WithContext(ctx).Model(&order.Order{}).
			Select("id, order_number, currency, subtotal_amount_minor, shipping_fee_minor, tax_amount_minor, discount_amount_minor, total_amount_minor, points_value_minor, pricing_snapshot").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize)
		if options.ToOrderID > 0 {
			query = query.Where("id <= ?", options.ToOrderID)
		}
		if err := query.Find(&orders).Error; err != nil {
			return report, fmt.Errorf("load orders for pricing snapshot audit: %w", err)
		}
		if len(orders) == 0 {
			break
		}

		orderIDs := make([]uint, 0, len(orders))
		for _, record := range orders {
			orderIDs = append(orderIDs, record.ID)
		}
		var items []order.OrderItem
		if err := s.db.WithContext(ctx).Model(&order.OrderItem{}).
			Select("id, order_id, product_id, variant_id, currency, quantity, price_minor, subtotal_minor, tax_amount_minor, discount_minor, total_minor, pricing_snapshot").
			Where("order_id IN ?", orderIDs).
			Order("order_id ASC, id ASC").
			Find(&items).Error; err != nil {
			return report, fmt.Errorf("load order items for pricing snapshot audit: %w", err)
		}
		itemsByOrder := make(map[uint][]order.OrderItem, len(orderIDs))
		for _, item := range items {
			itemsByOrder[item.OrderID] = append(itemsByOrder[item.OrderID], item)
		}

		for _, record := range orders {
			report.OrdersScanned++
			auditPricingOrder(&report, record)
			for _, item := range itemsByOrder[record.ID] {
				report.ItemsScanned++
				auditPricingItem(&report, record, item)
			}
			lastID = record.ID
		}
		if len(orders) < batchSize {
			break
		}
	}

	return report, nil
}

func normalizePricingSnapshotAuditOptions(options OrderPricingSnapshotAuditOptions) OrderPricingSnapshotAuditOptions {
	if options.BatchSize <= 0 {
		options.BatchSize = defaultPricingSnapshotAuditBatchSize
	}
	if options.MaxIssues <= 0 {
		options.MaxIssues = defaultPricingSnapshotAuditIssueLimit
	}
	return options
}

func auditPricingOrder(report *OrderPricingSnapshotAuditReport, record order.Order) {
	raw := strings.TrimSpace(string(record.PricingSnapshotData))
	if pricingSnapshotAuditMissing(raw) {
		report.OrdersMissingSnapshot++
		appendPricingSnapshotIssue(report, OrderPricingSnapshotAuditIssue{
			Scope: "order", Code: "missing", OrderID: record.ID, OrderNumber: record.OrderNumber,
			Detail: "pricing_snapshot is empty",
		})
		return
	}
	payload, err := pricing.ParseOrderPricingSnapshot([]byte(raw))
	if err != nil {
		report.OrdersInvalidSnapshot++
		appendPricingSnapshotIssue(report, OrderPricingSnapshotAuditIssue{
			Scope: "order", Code: "invalid", OrderID: record.ID, OrderNumber: record.OrderNumber,
			Detail: err.Error(),
		})
		return
	}
	report.OrdersWithValidSnapshot++
	var mismatches []string
	if payload.Currency != record.Currency {
		mismatches = append(mismatches, "currency")
	}
	if payload.SubtotalMinor != record.SubtotalAmountMinor {
		mismatches = append(mismatches, "subtotal_minor")
	}
	if payload.ShippingMinor != record.ShippingFeeMinor {
		mismatches = append(mismatches, "shipping_minor")
	}
	if payload.TaxMinor != record.TaxAmountMinor {
		mismatches = append(mismatches, "tax_minor")
	}
	if payload.DiscountTotalMinor != record.DiscountAmountMinor {
		mismatches = append(mismatches, "discount_total_minor")
	}
	if payload.PointsDiscountMinor != record.PointsValueMinor {
		mismatches = append(mismatches, "points_discount_minor")
	}
	if payload.TotalMinor != record.TotalAmountMinor {
		mismatches = append(mismatches, "total_minor")
	}
	if len(mismatches) > 0 {
		report.OrdersMismatchedSnapshot++
		appendPricingSnapshotIssue(report, OrderPricingSnapshotAuditIssue{
			Scope: "order", Code: "mismatch", OrderID: record.ID, OrderNumber: record.OrderNumber,
			Detail: "persisted fields differ: " + strings.Join(mismatches, ","),
		})
	}
}

func auditPricingItem(report *OrderPricingSnapshotAuditReport, record order.Order, item order.OrderItem) {
	raw := strings.TrimSpace(string(item.PricingSnapshotData))
	if pricingSnapshotAuditMissing(raw) {
		report.ItemsMissingSnapshot++
		appendPricingSnapshotIssue(report, OrderPricingSnapshotAuditIssue{
			Scope: "order_item", Code: "missing", OrderID: record.ID, OrderNumber: record.OrderNumber, ItemID: item.ID,
			Detail: "pricing_snapshot is empty",
		})
		return
	}
	line, err := pricing.ParseLineSnapshot([]byte(raw))
	if err != nil {
		report.ItemsInvalidSnapshot++
		appendPricingSnapshotIssue(report, OrderPricingSnapshotAuditIssue{
			Scope: "order_item", Code: "invalid", OrderID: record.ID, OrderNumber: record.OrderNumber, ItemID: item.ID,
			Detail: err.Error(),
		})
		return
	}
	report.ItemsWithValidSnapshot++
	var mismatches []string
	if line.UnitPrice().Currency().String() != item.Currency || line.UnitPrice().Currency().String() != record.Currency {
		mismatches = append(mismatches, "currency")
	}
	if line.ProductID() != item.ProductID {
		mismatches = append(mismatches, "product_id")
	}
	if line.VariantID() != orderItemVariantID(item) {
		mismatches = append(mismatches, "variant_id")
	}
	if line.Quantity() != item.Quantity {
		mismatches = append(mismatches, "quantity")
	}
	if line.UnitPrice().AmountMinor() != item.PriceMinor {
		mismatches = append(mismatches, "unit_price_minor")
	}
	if line.BaseSubtotal().AmountMinor() != item.SubtotalMinor {
		mismatches = append(mismatches, "base_subtotal_minor")
	}
	if line.DiscountTotal().AmountMinor() != item.DiscountMinor {
		mismatches = append(mismatches, "discount_total_minor")
	}
	if line.Tax().AmountMinor() != item.TaxAmountMinor {
		mismatches = append(mismatches, "tax_minor")
	}
	lineTotal, addErr := line.NetSubtotal().Add(line.Tax())
	if addErr != nil || lineTotal.AmountMinor() != item.TotalMinor {
		mismatches = append(mismatches, "total_minor")
	}
	if len(mismatches) > 0 {
		report.ItemsMismatchedSnapshot++
		appendPricingSnapshotIssue(report, OrderPricingSnapshotAuditIssue{
			Scope: "order_item", Code: "mismatch", OrderID: record.ID, OrderNumber: record.OrderNumber, ItemID: item.ID,
			Detail: "persisted fields differ: " + strings.Join(mismatches, ","),
		})
	}
}

func pricingSnapshotAuditMissing(raw string) bool {
	return raw == "" || raw == "{}" || raw == "null"
}

func appendPricingSnapshotIssue(report *OrderPricingSnapshotAuditReport, issue OrderPricingSnapshotAuditIssue) {
	report.TotalIssues++
	if len(report.Issues) >= report.Options.MaxIssues {
		report.IssuesTruncated = true
		return
	}
	report.Issues = append(report.Issues, issue)
}
