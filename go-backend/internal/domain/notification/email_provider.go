package notification

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

const (
	EmailProviderDriverSMTP = "smtp"

	EmailProviderEncryptionNone     = "none"
	EmailProviderEncryptionSTARTTLS = "starttls"
	EmailProviderEncryptionTLS      = "tls"

	EmailProviderTestStatusHealthy = "healthy"
	EmailProviderTestStatusFailed  = "failed"
)

var (
	ErrEmailProviderCodeRequired      = errors.New("email provider code is required")
	ErrEmailProviderNameRequired      = errors.New("email provider name is required")
	ErrEmailProviderDriverInvalid     = errors.New("email provider driver is invalid")
	ErrEmailProviderHostRequired      = errors.New("email provider host is required")
	ErrEmailProviderPortInvalid       = errors.New("email provider port is invalid")
	ErrEmailProviderFromEmailInvalid  = errors.New("email provider from email is invalid")
	ErrEmailProviderReplyToInvalid    = errors.New("email provider reply-to email is invalid")
	ErrEmailProviderEncryptionInvalid = errors.New("email provider encryption type is invalid")
)

// EmailProviderConfig stores a configured outbound mail transport. Secret
// material is persisted only in PasswordEncrypted and is never serialized.
type EmailProviderConfig struct {
	ID                uint       `gorm:"primarykey" json:"id"`
	Code              string     `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Name              string     `gorm:"size:128;not null" json:"name"`
	Driver            string     `gorm:"size:32;not null;default:'smtp'" json:"driver"`
	Host              string     `gorm:"size:255;not null" json:"host"`
	Port              int        `gorm:"not null;default:587" json:"port"`
	Username          string     `gorm:"size:255;not null;default:''" json:"username"`
	PasswordEncrypted string     `gorm:"type:text;not null;default:''" json:"-"`
	FromName          string     `gorm:"size:128;not null" json:"from_name"`
	FromEmail         string     `gorm:"size:255;not null" json:"from_email"`
	ReplyTo           string     `gorm:"size:255;not null;default:''" json:"reply_to"`
	EncryptionType    string     `gorm:"size:16;not null;default:'starttls'" json:"encryption_type"`
	IsActive          bool       `gorm:"not null;default:false;index" json:"is_active"`
	IsDefault         bool       `gorm:"not null;default:false;index" json:"is_default"`
	LastTestedAt      *time.Time `json:"last_tested_at,omitempty"`
	LastTestStatus    string     `gorm:"size:32;not null;default:''" json:"last_test_status"`
	LastTestError     string     `gorm:"size:500;not null;default:''" json:"last_test_error,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (EmailProviderConfig) TableName() string { return "email_provider_configs" }

type EmailProviderView struct {
	ID             uint       `json:"id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Driver         string     `json:"driver"`
	Host           string     `json:"host"`
	Port           int        `json:"port"`
	Username       string     `json:"username"`
	HasPassword    bool       `json:"has_password"`
	FromName       string     `json:"from_name"`
	FromEmail      string     `json:"from_email"`
	ReplyTo        string     `json:"reply_to,omitempty"`
	EncryptionType string     `json:"encryption_type"`
	IsActive       bool       `json:"is_active"`
	IsDefault      bool       `json:"is_default"`
	LastTestedAt   *time.Time `json:"last_tested_at,omitempty"`
	LastTestStatus string     `json:"last_test_status,omitempty"`
	LastTestError  string     `json:"last_test_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (p *EmailProviderConfig) Validate() error {
	if p == nil || strings.TrimSpace(p.Code) == "" {
		return ErrEmailProviderCodeRequired
	}
	if strings.TrimSpace(p.Name) == "" {
		return ErrEmailProviderNameRequired
	}
	if strings.ToLower(strings.TrimSpace(p.Driver)) != EmailProviderDriverSMTP {
		return ErrEmailProviderDriverInvalid
	}
	if strings.TrimSpace(p.Host) == "" {
		return ErrEmailProviderHostRequired
	}
	if p.Port < 1 || p.Port > 65535 {
		return ErrEmailProviderPortInvalid
	}
	if !validEmailAddress(p.FromEmail) {
		return ErrEmailProviderFromEmailInvalid
	}
	if strings.TrimSpace(p.ReplyTo) != "" && !validEmailAddress(p.ReplyTo) {
		return ErrEmailProviderReplyToInvalid
	}
	switch strings.ToLower(strings.TrimSpace(p.EncryptionType)) {
	case EmailProviderEncryptionNone, EmailProviderEncryptionSTARTTLS, EmailProviderEncryptionTLS:
	default:
		return ErrEmailProviderEncryptionInvalid
	}
	return nil
}

func validEmailAddress(value string) bool {
	parsed, err := mail.ParseAddress(strings.TrimSpace(value))
	return err == nil && parsed.Address == strings.TrimSpace(value)
}

func (p EmailProviderConfig) View() EmailProviderView {
	return EmailProviderView{
		ID: p.ID, Code: p.Code, Name: p.Name, Driver: p.Driver, Host: p.Host,
		Port: p.Port, Username: p.Username, HasPassword: strings.TrimSpace(p.PasswordEncrypted) != "",
		FromName: p.FromName, FromEmail: p.FromEmail, ReplyTo: p.ReplyTo,
		EncryptionType: p.EncryptionType, IsActive: p.IsActive, IsDefault: p.IsDefault,
		LastTestedAt: p.LastTestedAt, LastTestStatus: p.LastTestStatus, LastTestError: p.LastTestError,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
