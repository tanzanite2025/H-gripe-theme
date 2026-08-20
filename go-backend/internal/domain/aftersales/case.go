package aftersales

import (
	"errors"
	"strings"
	"time"
)

const (
	TypeReturnRefund    = "return_refund"
	TypeExchange        = "exchange"
	TypeRefundOnly      = "refund_only"
	TypeReshipment      = "reshipment"
	TypeCustomerRequest = "customer_request"
)

const (
	StatusRequested       = "requested"
	StatusReviewing       = "reviewing"
	StatusApproved        = "approved"
	StatusAwaitingReturn  = "awaiting_return"
	StatusReturnInTransit = "return_in_transit"
	StatusReceived        = "received"
	StatusInspecting      = "inspecting"
	StatusResolving       = "resolving"
	StatusCompleted       = "completed"
	StatusRejected        = "rejected"
	StatusCancelled       = "cancelled"
	StatusException       = "exception"
)

var validTypes = map[string]struct{}{
	TypeReturnRefund:    {},
	TypeExchange:        {},
	TypeRefundOnly:      {},
	TypeReshipment:      {},
	TypeCustomerRequest: {},
}

var validStatuses = map[string]struct{}{
	StatusRequested:       {},
	StatusReviewing:       {},
	StatusApproved:        {},
	StatusAwaitingReturn:  {},
	StatusReturnInTransit: {},
	StatusReceived:        {},
	StatusInspecting:      {},
	StatusResolving:       {},
	StatusCompleted:       {},
	StatusRejected:        {},
	StatusCancelled:       {},
	StatusException:       {},
}

var statusTransitions = map[string][]string{
	StatusRequested: {
		StatusReviewing,
		StatusCancelled,
	},
	StatusReviewing: {
		StatusApproved,
		StatusRejected,
		StatusCancelled,
	},
	StatusApproved: {
		StatusAwaitingReturn,
		StatusResolving,
		StatusCancelled,
	},
	StatusAwaitingReturn: {
		StatusReturnInTransit,
		StatusCancelled,
	},
	StatusReturnInTransit: {
		StatusReceived,
		StatusException,
		StatusCancelled,
	},
	StatusReceived: {
		StatusInspecting,
		StatusException,
	},
	StatusInspecting: {
		StatusResolving,
		StatusException,
	},
	StatusResolving: {
		StatusCompleted,
		StatusException,
	},
	StatusException: {
		StatusReviewing,
		StatusCancelled,
	},
	StatusCompleted: {},
	StatusRejected:  {},
	StatusCancelled: {},
}

type AfterSalesCase struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	OrderID     uint       `gorm:"not null;index" json:"order_id"`
	OrderNumber string     `gorm:"->" json:"order_number"`
	Type        string     `gorm:"not null;index" json:"type"`
	Status      string     `gorm:"not null;index" json:"status"`
	Reason      string     `gorm:"type:text;not null" json:"reason"`
	Description string     `gorm:"type:text" json:"description"`
	Resolution  string     `gorm:"type:text" json:"resolution"`
	CreatedBy   uint       `gorm:"not null;default:0" json:"created_by"`
	UpdatedBy   uint       `gorm:"not null;default:0" json:"updated_by"`
	ClosedAt    *time.Time `gorm:"index" json:"closed_at,omitempty"`

	Items        []AfterSalesCaseItem       `gorm:"foreignKey:CaseID" json:"items"`
	Events       []AfterSalesCaseEvent      `gorm:"foreignKey:CaseID" json:"events"`
	Attachments  []AfterSalesCaseAttachment `gorm:"foreignKey:CaseID" json:"attachments"`
	RefundReview *AfterSalesRefundReview    `gorm:"foreignKey:CaseID" json:"refund_review,omitempty"`

	RefundReviewMaximumAmount float64 `gorm:"-" json:"refund_review_maximum_amount"`
	RefundReviewCurrency      string  `gorm:"-" json:"refund_review_currency"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AfterSalesCase) TableName() string {
	return "after_sales_cases"
}

type AfterSalesCaseItem struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	CaseID      uint   `gorm:"not null;index" json:"case_id"`
	OrderID     uint   `gorm:"not null;index" json:"order_id"`
	OrderItemID uint   `gorm:"not null;index" json:"order_item_id"`
	ProductID   uint   `gorm:"not null;index" json:"product_id"`
	VariantID   *uint  `gorm:"index" json:"variant_id,omitempty"`
	ProductName string `gorm:"not null" json:"product_name"`
	SKU         string `json:"sku"`
	Quantity    int    `gorm:"not null" json:"quantity"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AfterSalesCaseItem) TableName() string {
	return "after_sales_case_items"
}

func IsValidType(value string) bool {
	_, ok := validTypes[strings.TrimSpace(value)]
	return ok
}

func IsValidStatus(value string) bool {
	_, ok := validStatuses[strings.TrimSpace(value)]
	return ok
}

func (c *AfterSalesCase) CanTransitionTo(targetStatus string) bool {
	if c == nil {
		return false
	}
	allowedStatuses, exists := statusTransitions[c.Status]
	if !exists {
		return false
	}
	for _, status := range allowedStatuses {
		if status == targetStatus {
			return true
		}
	}
	return false
}

func (c *AfterSalesCase) Validate() error {
	if c == nil {
		return errors.New("after-sales case is required")
	}
	if !IsValidType(c.Type) {
		return errors.New("invalid after-sales case type")
	}
	if !IsValidStatus(c.Status) {
		return errors.New("invalid after-sales case status")
	}
	if strings.TrimSpace(c.Reason) == "" {
		return errors.New("after-sales case reason is required")
	}
	return nil
}

func IsTerminalStatus(status string) bool {
	switch status {
	case StatusCompleted, StatusRejected, StatusCancelled:
		return true
	default:
		return false
	}
}
