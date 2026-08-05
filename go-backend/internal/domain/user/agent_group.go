package user

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type AgentGroup struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Code        string         `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"type:text;not null;default:''" json:"description"`
	Status      string         `gorm:"size:20;default:'active';index" json:"status"`
	SortOrder   int            `gorm:"default:0;index" json:"sort_order"`
	Agents      []AgentProfile `gorm:"many2many:customer_service_agent_group_members;joinForeignKey:GroupID;joinReferences:AgentProfileID" json:"agents,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AgentGroup) TableName() string {
	return "customer_service_agent_groups"
}

func (g *AgentGroup) BeforeCreate(tx *gorm.DB) error {
	g.applyDefaults()
	return nil
}

func (g *AgentGroup) BeforeUpdate(tx *gorm.DB) error {
	g.applyDefaults()
	return nil
}

func (g *AgentGroup) applyDefaults() {
	g.Code = normalizeAgentGroupCode(g.Code)
	g.Name = strings.TrimSpace(g.Name)
	g.Description = strings.TrimSpace(g.Description)
	if strings.TrimSpace(g.Status) == "" {
		g.Status = "active"
	}
}

func normalizeAgentGroupCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

type AgentGroupMember struct {
	ID             uint         `gorm:"primarykey" json:"id"`
	GroupID        uint         `gorm:"uniqueIndex:idx_customer_service_agent_group_member;not null;index" json:"group_id"`
	AgentProfileID uint         `gorm:"uniqueIndex:idx_customer_service_agent_group_member;not null;index" json:"agent_profile_id"`
	Group          AgentGroup   `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	AgentProfile   AgentProfile `gorm:"foreignKey:AgentProfileID" json:"agent_profile,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

func (AgentGroupMember) TableName() string {
	return "customer_service_agent_group_members"
}
