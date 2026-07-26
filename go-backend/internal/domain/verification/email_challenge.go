package verification

import (
	"time"

	"gorm.io/gorm"
)

type EmailChallenge struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Purpose   string         `gorm:"size:80;not null;index" json:"purpose"`
	Email     string         `gorm:"size:255;not null;index" json:"email"`
	Subject   string         `gorm:"size:500;not null;index" json:"subject"`
	TokenHash string         `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expires_at"`
	UsedAt    *time.Time     `gorm:"index" json:"used_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (EmailChallenge) TableName() string {
	return "email_verification_challenges"
}
