package audit

import (
	"time"
)

// AuditLog 审计日志
type AuditLog struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`
	Username    string    `json:"username"`
	Action      string    `gorm:"not null;index" json:"action"` // create, update, delete, view
	Resource    string    `gorm:"not null;index" json:"resource"` // order, product, user, etc.
	ResourceID  uint      `gorm:"index" json:"resource_id"`
	Method      string    `json:"method"` // GET, POST, PUT, DELETE
	Path        string    `json:"path"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `gorm:"type:text" json:"user_agent"`
	Changes     string    `gorm:"type:text" json:"changes"` // JSON格式的变更内容
	OldValue    string    `gorm:"type:text" json:"old_value"` // JSON格式
	NewValue    string    `gorm:"type:text" json:"new_value"` // JSON格式
	Status      string    `json:"status"` // success, failed
	ErrorMessage string   `gorm:"type:text" json:"error_message"`
	Duration    int       `json:"duration"` // 毫秒
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}
