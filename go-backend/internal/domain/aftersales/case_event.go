package aftersales

import "time"

type AfterSalesCaseEvent struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CaseID       uint      `gorm:"not null;index" json:"case_id"`
	FromStatus   string    `gorm:"not null;default:''" json:"from_status"`
	ToStatus     string    `gorm:"not null" json:"to_status"`
	Resolution   string    `gorm:"type:text;not null;default:''" json:"resolution"`
	UpdatedBy    uint      `gorm:"not null;default:0" json:"updated_by"`
	OperatorName string    `gorm:"-" json:"operator_name"`
	CreatedAt    time.Time `json:"created_at"`
}

func (AfterSalesCaseEvent) TableName() string {
	return "after_sales_case_events"
}

// AfterSalesCaseEventArchive is the cold-storage copy of a terminal case
// timeline event.
type AfterSalesCaseEventArchive struct {
	AfterSalesCaseEvent
}

func (AfterSalesCaseEventArchive) TableName() string {
	return "after_sales_case_events_archive"
}
