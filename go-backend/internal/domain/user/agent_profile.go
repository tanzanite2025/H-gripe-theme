package user

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type AgentProfile struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	AgentID      string         `gorm:"uniqueIndex;size:50;not null" json:"agent_id"`
	UserID       *uint          `gorm:"index" json:"user_id"`
	User         *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Email        string         `gorm:"size:100;index" json:"email"`
	Avatar       string         `gorm:"size:500" json:"avatar"`
	WhatsApp     string         `gorm:"size:50" json:"whatsapp"`
	Status       string         `gorm:"size:20;default:'active';index" json:"status"`
	OnlineStatus string         `gorm:"size:20;default:'offline';index" json:"online_status"`
	LastActiveAt *time.Time     `json:"last_active_at"`
	LastLogin    *time.Time     `json:"last_login"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AgentProfile) TableName() string {
	return "customer_service_agent_profiles"
}

func (p *AgentProfile) BeforeCreate(tx *gorm.DB) error {
	p.applyDefaults()
	return nil
}

func (p *AgentProfile) BeforeUpdate(tx *gorm.DB) error {
	p.applyDefaults()
	return nil
}

func (p *AgentProfile) applyDefaults() {
	if strings.TrimSpace(p.Status) == "" {
		p.Status = "active"
	}
	if strings.TrimSpace(p.OnlineStatus) == "" {
		p.OnlineStatus = "offline"
	}
}

func (p AgentProfile) DisplayName() string {
	if strings.TrimSpace(p.Name) != "" {
		return strings.TrimSpace(p.Name)
	}
	if p.User != nil {
		return displayName(p.User.FirstName, p.User.LastName, p.User.Username, p.User.Email)
	}
	return strings.TrimSpace(p.Email)
}

func (p AgentProfile) PublicEmail() string {
	if strings.TrimSpace(p.Email) != "" {
		return strings.TrimSpace(p.Email)
	}
	if p.User != nil {
		return strings.TrimSpace(p.User.Email)
	}
	return ""
}
