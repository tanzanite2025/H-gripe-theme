package feedback

import (
	"time"

	"gorm.io/gorm"
)

type Feedback struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ThreadKey    string         `gorm:"not null;index" json:"thread_key"`
	PagePath     string         `gorm:"index" json:"page_path"`
	PageTitle    string         `json:"page_title"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	Name         string         `json:"name"`
	Email        string         `json:"-"`
	SourceHash   string         `gorm:"size:80;index" json:"-"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	Status       string         `gorm:"index;default:'pending'" json:"status"`
	Locale       string         `json:"locale"`
	ReplyContent string         `gorm:"type:text" json:"reply_content"`
	RepliedAt    *time.Time     `json:"replied_at"`
	RepliedBy    uint           `json:"replied_by"`
	ReviewedAt   *time.Time     `json:"reviewed_at"`
	ReviewedBy   uint           `json:"reviewed_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Feedback) TableName() string {
	return "feedback"
}

type Response struct {
	ID           uint       `json:"id"`
	ThreadKey    string     `json:"thread_key"`
	PagePath     *string    `json:"page_path"`
	PageTitle    *string    `json:"page_title"`
	Name         *string    `json:"name"`
	Content      string     `json:"content"`
	Status       string     `json:"status"`
	Locale       *string    `json:"locale"`
	ReplyContent *string    `json:"reply_content"`
	RepliedAt    *time.Time `json:"replied_at"`
	CreatedAt    time.Time  `json:"created_at"`
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
	var pagePath *string
	if f.PagePath != "" {
		pagePath = &f.PagePath
	}
	var pageTitle *string
	if f.PageTitle != "" {
		pageTitle = &f.PageTitle
	}
	var replyContent *string
	if f.ReplyContent != "" {
		replyContent = &f.ReplyContent
	}

	return Response{
		ID:           f.ID,
		ThreadKey:    f.ThreadKey,
		PagePath:     pagePath,
		PageTitle:    pageTitle,
		Name:         name,
		Content:      f.Content,
		Status:       f.Status,
		Locale:       locale,
		ReplyContent: replyContent,
		RepliedAt:    f.RepliedAt,
		CreatedAt:    f.CreatedAt,
	}
}
