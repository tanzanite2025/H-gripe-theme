package aftersales

import "time"

const (
	RefundReviewStatusPending   = "pending"
	RefundReviewStatusApproved  = "approved"
	RefundReviewStatusRejected  = "rejected"
	RefundReviewStatusCancelled = "cancelled"
)

type AfterSalesRefundReview struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	CaseID         uint       `gorm:"not null;uniqueIndex" json:"case_id"`
	Status         string     `gorm:"not null;index" json:"status"`
	ProposedAmount float64    `gorm:"type:numeric(18,2);not null" json:"proposed_amount"`
	Currency       string     `gorm:"size:8;not null" json:"currency"`
	RequestNotes   string     `gorm:"type:text;not null;default:''" json:"request_notes"`
	DecisionNotes  string     `gorm:"type:text;not null;default:''" json:"decision_notes"`
	CreatedBy      uint       `gorm:"not null;default:0" json:"created_by"`
	UpdatedBy      uint       `gorm:"not null;default:0" json:"updated_by"`
	ReviewedByID   *uint      `gorm:"index" json:"reviewed_by_id,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	LinkedRefundID *uint      `gorm:"index" json:"linked_refund_id,omitempty"`

	CreatorName  string    `gorm:"-" json:"creator_name,omitempty"`
	ReviewerName string    `gorm:"-" json:"reviewer_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (AfterSalesRefundReview) TableName() string {
	return "after_sales_refund_reviews"
}

func IsRefundReviewCaseType(caseType string) bool {
	switch caseType {
	case TypeReturnRefund, TypeRefundOnly:
		return true
	default:
		return false
	}
}

func IsValidRefundReviewStatus(status string) bool {
	switch status {
	case RefundReviewStatusPending,
		RefundReviewStatusApproved,
		RefundReviewStatusRejected,
		RefundReviewStatusCancelled:
		return true
	default:
		return false
	}
}

func IsRefundReviewDecisionStatus(status string) bool {
	switch status {
	case RefundReviewStatusApproved,
		RefundReviewStatusRejected,
		RefundReviewStatusCancelled:
		return true
	default:
		return false
	}
}
