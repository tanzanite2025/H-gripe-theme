package service

import (
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/pkg/email"
	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"
)

const EmailProviderMasterKeyEnv = "EMAIL_PROVIDER_MASTER_KEY"

var (
	ErrEmailProviderServiceNotConfigured = errors.New("email provider service is not configured")
	ErrEmailProviderMasterKeyRequired    = errors.New("email provider master key is required")
	ErrInvalidEmailProvider              = errors.New("invalid email provider")
	ErrEmailProviderCodeConflict         = errors.New("email provider code already exists")
	ErrEmailProviderTestFailed           = errors.New("email provider test failed")
	ErrEmailProviderPasswordUnavailable  = errors.New("email provider password cannot be decrypted")
)

type EmailProviderService struct {
	repo    *repository.EmailProviderRepository
	factory func(*email.SMTPConfig) (email.EmailService, error)
}

type SaveEmailProviderInput struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	Driver         string `json:"driver"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	FromName       string `json:"from_name"`
	FromEmail      string `json:"from_email"`
	ReplyTo        string `json:"reply_to"`
	EncryptionType string `json:"encryption_type"`
	IsActive       *bool  `json:"is_active"`
	IsDefault      *bool  `json:"is_default"`
}

func NewEmailProviderService(repo *repository.EmailProviderRepository) *EmailProviderService {
	return &EmailProviderService{repo: repo, factory: email.NewEmailService}
}

func (s *EmailProviderService) List() ([]notification.EmailProviderView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrEmailProviderServiceNotConfigured
	}
	records, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	views := make([]notification.EmailProviderView, 0, len(records))
	for _, record := range records {
		views = append(views, record.View())
	}
	return views, nil
}

func (s *EmailProviderService) Get(id uint) (*notification.EmailProviderView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrEmailProviderServiceNotConfigured
	}
	record, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	view := record.View()
	return &view, nil
}

func (s *EmailProviderService) Create(input SaveEmailProviderInput) (*notification.EmailProviderView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrEmailProviderServiceNotConfigured
	}
	record, err := s.normalize(input, nil)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.FindByCode(record.Code); err == nil {
		return nil, ErrEmailProviderCodeConflict
	} else if !repository.IsRecordNotFound(err) {
		return nil, err
	}
	wantsDefault := record.IsDefault
	record.IsDefault = false
	if err := s.repo.Create(&record); err != nil {
		return nil, err
	}
	if wantsDefault {
		if err := s.repo.SetDefault(record.ID); err != nil {
			return nil, err
		}
		reloaded, err := s.repo.FindByID(record.ID)
		if err != nil {
			return nil, err
		}
		record = *reloaded
	}
	view := record.View()
	return &view, nil
}

func (s *EmailProviderService) Update(id uint, input SaveEmailProviderInput) (*notification.EmailProviderView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrEmailProviderServiceNotConfigured
	}
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	record, err := s.normalize(input, existing)
	if err != nil {
		return nil, err
	}
	record.ID = id
	wantsDefault := record.IsDefault
	if wantsDefault {
		// Clear the flag in the row update first. SetDefault performs the
		// replacement transaction after this update, avoiding the partial
		// unique index rejecting a second default row.
		record.IsDefault = false
	}
	if other, findErr := s.repo.FindByCode(record.Code); findErr == nil && other.ID != id {
		return nil, ErrEmailProviderCodeConflict
	} else if findErr != nil && !repository.IsRecordNotFound(findErr) {
		return nil, findErr
	}
	if err := s.repo.Update(&record); err != nil {
		return nil, err
	}
	if wantsDefault {
		if err := s.repo.SetDefault(id); err != nil {
			return nil, err
		}
	}
	return s.Get(id)
}

func (s *EmailProviderService) SetActive(id uint, active bool) (*notification.EmailProviderView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrEmailProviderServiceNotConfigured
	}
	record, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	record.IsActive = active
	if !active {
		record.IsDefault = false
	}
	if err := s.repo.Update(record); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *EmailProviderService) SetDefault(id uint) (*notification.EmailProviderView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrEmailProviderServiceNotConfigured
	}
	if _, err := s.repo.FindByID(id); err != nil {
		return nil, err
	}
	if err := s.repo.SetDefault(id); err != nil {
		return nil, err
	}
	return s.Get(id)
}

// DefaultSMTPConfig resolves the active default provider for runtime delivery.
// The bool result is false when no database provider is configured, in which
// case callers may intentionally fall back to the environment SMTP settings.
// Once a provider row exists, all validation/decryption errors are returned and
// must not be hidden by that fallback.
func (s *EmailProviderService) DefaultSMTPConfig() (*email.SMTPConfig, bool, error) {
	if s == nil || s.repo == nil {
		return nil, false, ErrEmailProviderServiceNotConfigured
	}
	record, err := s.repo.FindDefaultActive()
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	cfg, err := s.smtpConfigForRecord(record)
	if err != nil {
		return nil, true, err
	}
	return cfg, true, nil
}

// Test sends a real message through the selected provider and persists the
// result. The target address is deliberately required by the API so a test can
// never accidentally broadcast to an implicit recipient.
func (s *EmailProviderService) Test(id uint, targetEmail string) (*notification.EmailProviderView, error) {
	if s == nil || s.repo == nil {
		return nil, ErrEmailProviderServiceNotConfigured
	}
	targetEmail = strings.TrimSpace(targetEmail)
	parsed, err := mail.ParseAddress(targetEmail)
	if err != nil || parsed.Address != targetEmail {
		return nil, fmt.Errorf("%w: invalid target email", ErrInvalidEmailProvider)
	}
	record, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	testedAt := time.Now().UTC()
	cfg, configErr := s.smtpConfigForRecord(record)
	if configErr == nil {
		var sender email.EmailService
		factory := s.factory
		if factory == nil {
			factory = email.NewEmailService
		}
		sender, configErr = factory(cfg)
		if configErr == nil {
			configErr = sender.SendEmail([]string{targetEmail}, "SMTP provider test", "This is a test email from the configured SMTP provider.")
		}
	}
	status := notification.EmailProviderTestStatusHealthy
	testError := ""
	if configErr != nil {
		status = notification.EmailProviderTestStatusFailed
		testError = truncateEmailProviderError(configErr.Error())
	}
	if persistErr := s.repo.UpdateTestState(id, status, testError, testedAt); persistErr != nil {
		return nil, fmt.Errorf("persist email provider test result: %w", persistErr)
	}
	view, viewErr := s.Get(id)
	if viewErr != nil {
		return nil, viewErr
	}
	if configErr != nil {
		return view, fmt.Errorf("%w: %v", ErrEmailProviderTestFailed, configErr)
	}
	return view, nil
}

func (s *EmailProviderService) smtpConfigForRecord(record *notification.EmailProviderConfig) (*email.SMTPConfig, error) {
	if record == nil {
		return nil, ErrInvalidEmailProvider
	}
	if err := record.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidEmailProvider, err)
	}
	password := ""
	if strings.TrimSpace(record.PasswordEncrypted) != "" {
		masterKey := EmailProviderMasterKey()
		if masterKey == "" {
			return nil, ErrEmailProviderPasswordUnavailable
		}
		var err error
		password, err = secretbox.DecryptString(record.PasswordEncrypted, masterKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrEmailProviderPasswordUnavailable, err)
		}
	}
	return &email.SMTPConfig{
		Host:           record.Host,
		Port:           record.Port,
		Username:       record.Username,
		Password:       password,
		From:           record.FromEmail,
		FromName:       record.FromName,
		ReplyTo:        record.ReplyTo,
		EncryptionType: record.EncryptionType,
		Timeout:        10 * time.Second,
	}, nil
}

func truncateEmailProviderError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 500 {
		return value
	}
	return value[:500]
}

func (s *EmailProviderService) normalize(input SaveEmailProviderInput, existing *notification.EmailProviderConfig) (notification.EmailProviderConfig, error) {
	record := notification.EmailProviderConfig{
		Code: strings.TrimSpace(input.Code), Name: strings.TrimSpace(input.Name), Driver: strings.ToLower(strings.TrimSpace(input.Driver)),
		Host: strings.TrimSpace(input.Host), Port: input.Port, Username: strings.TrimSpace(input.Username),
		FromName: strings.TrimSpace(input.FromName), FromEmail: strings.TrimSpace(input.FromEmail), ReplyTo: strings.TrimSpace(input.ReplyTo),
		EncryptionType: strings.ToLower(strings.TrimSpace(input.EncryptionType)), IsActive: true,
	}
	if existing != nil {
		record.ID = existing.ID
		record.PasswordEncrypted = existing.PasswordEncrypted
		record.IsActive = existing.IsActive
		record.IsDefault = existing.IsDefault
		record.LastTestedAt = existing.LastTestedAt
		record.LastTestStatus = existing.LastTestStatus
		record.LastTestError = existing.LastTestError
	}
	if record.Driver == "" {
		record.Driver = notification.EmailProviderDriverSMTP
	}
	if record.Port == 0 {
		record.Port = 587
	}
	if record.EncryptionType == "" {
		record.EncryptionType = notification.EmailProviderEncryptionSTARTTLS
	}
	if input.IsActive != nil {
		record.IsActive = *input.IsActive
	}
	if input.IsDefault != nil {
		record.IsDefault = *input.IsDefault
	}
	if err := record.Validate(); err != nil {
		return notification.EmailProviderConfig{}, fmt.Errorf("%w: %w", ErrInvalidEmailProvider, err)
	}
	if strings.TrimSpace(input.Password) != "" {
		masterKey := EmailProviderMasterKey()
		if masterKey == "" {
			return notification.EmailProviderConfig{}, ErrEmailProviderMasterKeyRequired
		}
		encrypted, err := secretbox.EncryptString(input.Password, masterKey)
		if err != nil {
			return notification.EmailProviderConfig{}, fmt.Errorf("encrypt email provider password: %w", err)
		}
		record.PasswordEncrypted = encrypted
	}
	if record.IsDefault {
		record.IsActive = true
	}
	return record, nil
}

func EmailProviderMasterKey() string {
	for _, env := range []string{EmailProviderMasterKeyEnv, "SECURITY_MASTER_KEY", "PAYMENT_CONFIG_MASTER_KEY"} {
		if value := strings.TrimSpace(os.Getenv(env)); value != "" {
			return value
		}
	}
	return ""
}
