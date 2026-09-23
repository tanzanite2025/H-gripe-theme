package service

import (
	"fmt"

	"commerce-platform/internal/pkg/email"
)

// RuntimeEmailService selects the active database provider for every delivery.
// This keeps admin changes effective without restarting the application while
// retaining the existing environment SMTP configuration as an explicit
// fallback when no active default provider has been configured yet.
type RuntimeEmailService struct {
	providers *EmailProviderService
	fallback  email.EmailService
	factory   func(*email.SMTPConfig) (email.EmailService, error)
}

func NewRuntimeEmailService(providers *EmailProviderService, fallback email.EmailService) *RuntimeEmailService {
	return &RuntimeEmailService{
		providers: providers,
		fallback:  fallback,
		factory:   email.NewEmailService,
	}
}

func (s *RuntimeEmailService) resolve() (email.EmailService, error) {
	if s == nil {
		return nil, ErrEmailProviderServiceNotConfigured
	}
	if s.providers == nil {
		if s.fallback == nil {
			return nil, ErrEmailProviderServiceNotConfigured
		}
		return s.fallback, nil
	}
	config, configured, err := s.providers.DefaultSMTPConfig()
	if err != nil {
		return nil, fmt.Errorf("resolve database email provider: %w", err)
	}
	if !configured {
		if s.fallback == nil {
			return nil, ErrEmailProviderServiceNotConfigured
		}
		return s.fallback, nil
	}
	factory := s.factory
	if factory == nil {
		factory = email.NewEmailService
	}
	selected, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("initialize database email provider: %w", err)
	}
	return selected, nil
}

func (s *RuntimeEmailService) SendEmail(to []string, subject, body string) error {
	sender, err := s.resolve()
	if err != nil {
		return err
	}
	return sender.SendEmail(to, subject, body)
}

// SendRenderedEmail preserves the HTML/text multipart path used by canonical
// transactional notifications. A third-party test double that only implements
// EmailService receives the text alternative instead.
func (s *RuntimeEmailService) SendRenderedEmail(to []string, subject, htmlBody, textBody string) error {
	sender, err := s.resolve()
	if err != nil {
		return err
	}
	if rendered, ok := sender.(interface {
		SendRenderedEmail([]string, string, string, string) error
	}); ok {
		return rendered.SendRenderedEmail(to, subject, htmlBody, textBody)
	}
	return sender.SendEmail(to, subject, textBody)
}

func (s *RuntimeEmailService) SendHTMLEmail(to []string, subject, templateName string, data interface{}) error {
	sender, err := s.resolve()
	if err != nil {
		return err
	}
	return sender.SendHTMLEmail(to, subject, templateName, data)
}

func (s *RuntimeEmailService) SendPasswordReset(to string, resetData interface{}) error {
	sender, err := s.resolve()
	if err != nil {
		return err
	}
	return sender.SendPasswordReset(to, resetData)
}

func (s *RuntimeEmailService) SendWelcomeEmail(to string, userData interface{}) error {
	sender, err := s.resolve()
	if err != nil {
		return err
	}
	return sender.SendWelcomeEmail(to, userData)
}
