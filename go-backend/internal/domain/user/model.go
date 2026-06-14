package user

import (
	"strconv"
	"strings"
	"time"

	"tanzanite/internal/domain/auth"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Username  string         `gorm:"uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"not null" json:"-"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Role      string         `gorm:"default:'user'" json:"role"`
	Locale    string         `gorm:"default:'en'" json:"locale"`
	Status    string         `gorm:"default:'active'" json:"status"` // active, inactive, suspended
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// HashPassword 加密密码
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// BeforeCreate GORM钩子：创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Locale == "" {
		u.Locale = "en"
	}
	if u.Role == "" {
		u.Role = "user"
	}
	if u.Status == "" {
		u.Status = "active"
	}
	return nil
}

// UserResponse 用户响应结构（不包含敏感信息）
type UserResponse struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	Roles       []string  `json:"roles"`
	Locale      string    `json:"locale"`
	Status      string    `json:"status"`
	IsAgent     bool      `json:"is_agent"`
	AgentID     string    `json:"agent_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToResponse 转换为响应结构
func (u *User) ToResponse() *UserResponse {
	role := auth.NormalizeRole(u.Role)
	isAgent := auth.IsCustomerServiceAgentRole(u.Role)
	agentID := ""
	if isAgent {
		agentID = strconv.FormatUint(uint64(u.ID), 10)
	}

	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: displayName(u.FirstName, u.LastName, u.Username, u.Email),
		Role:        string(role),
		Roles:       []string{string(role)},
		Locale:      u.Locale,
		Status:      u.Status,
		IsAgent:     isAgent,
		AgentID:     agentID,
		CreatedAt:   u.CreatedAt,
	}
}

func displayName(firstName, lastName, username, email string) string {
	fullName := strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
	if fullName != "" {
		return fullName
	}
	if strings.TrimSpace(username) != "" {
		return strings.TrimSpace(username)
	}
	return strings.TrimSpace(email)
}
