package feedback

import (
	"time"

	"gorm.io/gorm"
)

type Feedback struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ThreadKey string         `gorm:"not null;index" json:"thread_key"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Name      string         `json:"name"`
	Email     string         `json:"-"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Status    string         `gorm:"index;default:'pending'" json:"status"`
	Locale    string         `json:"locale"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Feedback) TableName() string {
	return "feedback"
}

type Response struct {
	ID        uint      `json:"id"`
	ThreadKey string    `json:"thread_key"`
	UserID    uint      `json:"user_id"`
	Name      *string   `json:"name"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	Locale    *string   `json:"locale"`
	CreatedAt time.Time `json:"created_at"`
}

func (f Feedback) ToResponse() Response {
	var name *string
	if f.Name != "" {
		name = &f.Name
	}
	var locale *string
	if f.Locale != "" {
		locale = &f.Locale
	}

	return Response{
		ID:        f.ID,
		ThreadKey: f.ThreadKey,
		UserID:    f.UserID,
		Name:      name,
		Content:   f.Content,
		Status:    f.Status,
		Locale:    locale,
		CreatedAt: f.CreatedAt,
	}
}
