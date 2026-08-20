package aftersales

import "time"

const (
	AttachmentKindImage = "image"
	AttachmentKindVideo = "video"
)

type AfterSalesCaseAttachment struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	CaseID      uint   `gorm:"not null;index" json:"case_id"`
	Kind        string `gorm:"size:16;not null" json:"kind"`
	StorageURL  string `gorm:"type:text;not null" json:"-"`
	Filename    string `gorm:"size:255;not null;default:''" json:"filename"`
	ContentType string `gorm:"size:128;not null;default:''" json:"content_type"`
	SizeBytes   int64  `gorm:"not null;default:0" json:"size_bytes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AfterSalesCaseAttachment) TableName() string {
	return "after_sales_case_attachments"
}
