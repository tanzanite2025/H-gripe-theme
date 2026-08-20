package productiontest

import "time"

const (
	TestAccountStatusActive   = "active"
	TestAccountStatusDisabled = "disabled"
)

// TestAccount grants one ordinary user access to production-only test flows.
type TestAccount struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	UserID    uint       `gorm:"not null;uniqueIndex" json:"user_id"`
	Label     string     `gorm:"size:120;not null" json:"label"`
	Purpose   string     `gorm:"type:text;not null;default:''" json:"purpose"`
	Status    string     `gorm:"size:24;not null;default:'active';index" json:"status"`
	ExpiresAt *time.Time `gorm:"index" json:"expires_at,omitempty"`
	CreatedBy uint       `gorm:"not null;index" json:"created_by"`
	UpdatedBy uint       `gorm:"not null;index" json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (TestAccount) TableName() string {
	return "production_test_accounts"
}

func (a TestAccount) IsActiveAt(now time.Time) bool {
	if a.ID == 0 || a.UserID == 0 || a.Status != TestAccountStatusActive {
		return false
	}
	if a.ExpiresAt != nil && !a.ExpiresAt.After(now) {
		return false
	}
	return true
}
