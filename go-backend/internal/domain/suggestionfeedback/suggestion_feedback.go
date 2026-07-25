package suggestionfeedback

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SuggestionFeedback struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	UserID              uint           `gorm:"index" json:"user_id"`
	FullName            string         `gorm:"size:120" json:"full_name"`
	Email               string         `gorm:"size:190" json:"email"`
	Country             string         `gorm:"size:80" json:"country"`
	OrderNumber         string         `gorm:"size:80" json:"order_number"`
	ProductCategory     string         `gorm:"size:60" json:"product_category"`
	RequestType         string         `gorm:"size:60" json:"request_type"`
	Message             string         `gorm:"type:text;not null" json:"message"`
	Attachments         datatypes.JSON `gorm:"type:json" json:"attachments"`
	Meta                datatypes.JSON `gorm:"type:json" json:"meta"`
	Status              string         `gorm:"size:25;not null;default:'new';index" json:"status"`
	MemberLevelRequired string         `gorm:"size:60" json:"member_level_required"`
	MemberLevelMet      bool           `gorm:"default:false" json:"member_level_met"`
	EligibilityHash     string         `gorm:"size:190" json:"-"`
	ReviewedBy          *uint          `json:"reviewed_by"`
	ReviewedAt          *time.Time     `json:"reviewed_at"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SuggestionFeedback) TableName() string {
	return "suggestion_feedback"
}

type Attachment struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

func JSONFromAttachments(attachments []Attachment) datatypes.JSON {
	data, err := json.Marshal(attachments)
	if err != nil {
		return datatypes.JSON([]byte("[]"))
	}
	return datatypes.JSON(data)
}

func JSONFromMeta(meta map[string]string) datatypes.JSON {
	data, err := json.Marshal(meta)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(data)
}
