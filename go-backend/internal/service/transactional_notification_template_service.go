package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

// TransactionalNotificationTemplateService owns template persistence and
// locale selection. It deliberately has no EmailService dependency: loading or
// saving a template must never open an SMTP connection.
type TransactionalNotificationTemplateService struct {
	repo *repository.NotificationTemplateRepository
}

func NewTransactionalNotificationTemplateService(repo *repository.NotificationTemplateRepository) *TransactionalNotificationTemplateService {
	return &TransactionalNotificationTemplateService{repo: repo}
}

type SaveTransactionalNotificationTemplateInput struct {
	Code              string
	Locale            string
	Category          string
	Name              string
	SubjectTemplate   string
	BodyHTML          string
	BodyText          string
	AllowedVariables  []string
	RequiredVariables []string
	IsEnabled         bool
	Version           int
	ChangedByUserID   *uint
	ChangeReason      string
}

func (s *TransactionalNotificationTemplateService) Save(input SaveTransactionalNotificationTemplateInput) (*notification.EmailTemplate, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("notification template service is not configured")
	}
	definition, err := LookupTransactionalNotificationTemplate(input.Code)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Category) == "" {
		input.Category = definition.Category
	}
	if input.Category != definition.Category {
		return nil, ErrNotificationTemplateDefinitionMismatch
	}
	locale := locales.ResolveSupported(input.Locale)
	if locale == "" {
		return nil, fmt.Errorf("%w: %s", ErrNotificationTemplateLocaleUnavailable, input.Locale)
	}
	if !sameStringSet(input.RequiredVariables, definition.RequiredVariables) ||
		!sameStringSet(input.AllowedVariables, definition.AllowedVariables) {
		return nil, ErrNotificationTemplateDefinitionMismatch
	}
	allowedJSON, err := json.Marshal(input.AllowedVariables)
	if err != nil {
		return nil, fmt.Errorf("encode allowed template variables: %w", err)
	}
	requiredJSON, err := json.Marshal(input.RequiredVariables)
	if err != nil {
		return nil, fmt.Errorf("encode required template variables: %w", err)
	}
	template := &notification.EmailTemplate{
		Code:              definition.Code,
		Locale:            locale,
		Category:          input.Category,
		Name:              strings.TrimSpace(input.Name),
		SubjectTemplate:   input.SubjectTemplate,
		BodyHTML:          input.BodyHTML,
		BodyText:          input.BodyText,
		AllowedVariables:  datatypes.JSON(allowedJSON),
		RequiredVariables: datatypes.JSON(requiredJSON),
		IsEnabled:         input.IsEnabled,
		Version:           input.Version,
	}
	if template.Version <= 0 {
		template.Version = 1
	}
	if err := template.Validate(); err != nil {
		return nil, err
	}
	if existing, findErr := s.repo.FindByCodeLocale(template.Code, template.Locale); findErr == nil {
		template.ID = existing.ID
	} else if !repository.IsRecordNotFound(findErr) {
		return nil, findErr
	}
	version := &notification.EmailTemplateVersion{
		Version:         template.Version,
		ChangedByUserID: input.ChangedByUserID,
		ChangeReason:    strings.TrimSpace(input.ChangeReason),
	}
	if err := s.repo.SaveWithVersion(template, version); err != nil {
		return nil, err
	}
	return template, nil
}

// Resolve returns an enabled template using exact locale first and English as
// the final fallback. Disabled locale rows are never selected.
func (s *TransactionalNotificationTemplateService) Resolve(code, requestedLocale string) (*notification.EmailTemplate, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("notification template service is not configured")
	}
	definition, err := LookupTransactionalNotificationTemplate(code)
	if err != nil {
		return nil, err
	}
	templates, err := s.repo.ListByCode(definition.Code)
	if err != nil {
		return nil, err
	}
	availableLocales := make([]string, 0, len(templates))
	for _, template := range templates {
		if template.IsEnabled {
			availableLocales = append(availableLocales, template.Locale)
		}
	}
	resolved, err := ResolveTransactionalNotificationTemplateDefinition(definition.Code, requestedLocale, availableLocales)
	if err != nil {
		return nil, err
	}
	for index := range templates {
		candidate := &templates[index]
		if candidate.IsEnabled && candidate.Locale == resolved.Locale {
			if err := validatePersistedTemplateDefinition(*candidate, definition); err != nil {
				return nil, err
			}
			return candidate, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrNotificationTemplateLocaleUnavailable, resolved.Locale)
}

func (s *TransactionalNotificationTemplateService) List() ([]notification.EmailTemplate, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("notification template service is not configured")
	}
	return s.repo.ListAll()
}

func (s *TransactionalNotificationTemplateService) Get(id uint) (*notification.EmailTemplate, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("notification template service is not configured")
	}
	templateRecord, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	definition, err := LookupTransactionalNotificationTemplate(templateRecord.Code)
	if err != nil {
		return nil, err
	}
	if err := validatePersistedTemplateDefinition(*templateRecord, definition); err != nil {
		return nil, err
	}
	return templateRecord, nil
}

func (s *TransactionalNotificationTemplateService) Versions(id uint) ([]notification.EmailTemplateVersion, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("notification template service is not configured")
	}
	if _, err := s.Get(id); err != nil {
		return nil, err
	}
	return s.repo.ListVersions(id)
}

// Rollback restores a historical snapshot by writing it as a new current
// version. Existing snapshots remain immutable and the current row's version
// is checked inside SaveWithVersion for optimistic concurrency.
func (s *TransactionalNotificationTemplateService) Rollback(id uint, version int, changedByUserID *uint, changeReason string) (*notification.EmailTemplate, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("notification template service is not configured")
	}
	current, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if version <= 0 || version >= current.Version {
		return nil, notification.ErrTemplateVersionConflict
	}
	snapshot, err := s.repo.FindVersion(id, version)
	if err != nil {
		return nil, err
	}
	allowed, err := decodeTemplateVariables(snapshot.AllowedVariables)
	if err != nil {
		return nil, fmt.Errorf("decode rollback allowed variables: %w", err)
	}
	required, err := decodeTemplateVariables(snapshot.RequiredVariables)
	if err != nil {
		return nil, fmt.Errorf("decode rollback required variables: %w", err)
	}
	reason := strings.TrimSpace(changeReason)
	if reason == "" {
		reason = fmt.Sprintf("rollback to version %d", version)
	}
	return s.Save(SaveTransactionalNotificationTemplateInput{
		Code:              current.Code,
		Locale:            current.Locale,
		Category:          current.Category,
		Name:              snapshot.Name,
		SubjectTemplate:   snapshot.SubjectTemplate,
		BodyHTML:          snapshot.BodyHTML,
		BodyText:          snapshot.BodyText,
		AllowedVariables:  allowed,
		RequiredVariables: required,
		IsEnabled:         current.IsEnabled,
		Version:           current.Version + 1,
		ChangedByUserID:   changedByUserID,
		ChangeReason:      reason,
	})
}

// Render resolves an enabled locale-specific template and renders it without
// acquiring a provider connection. Delivery workers can call this method after
// the surrounding Outbox event has been claimed.
func (s *TransactionalNotificationTemplateService) Render(code, requestedLocale string, variables map[string]string) (RenderedTransactionalNotification, error) {
	templateRecord, err := s.Resolve(code, requestedLocale)
	if err != nil {
		return RenderedTransactionalNotification{}, err
	}
	return RenderTransactionalNotificationTemplate(templateRecord, variables)
}

func validatePersistedTemplateDefinition(template notification.EmailTemplate, definition NotificationTemplateDefinition) error {
	var allowed, required []string
	if err := json.Unmarshal(template.AllowedVariables, &allowed); err != nil {
		return fmt.Errorf("decode persisted allowed template variables: %w", err)
	}
	if err := json.Unmarshal(template.RequiredVariables, &required); err != nil {
		return fmt.Errorf("decode persisted required template variables: %w", err)
	}
	if !sameStringSet(allowed, definition.AllowedVariables) || !sameStringSet(required, definition.RequiredVariables) {
		return ErrNotificationTemplateDefinitionMismatch
	}
	return nil
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	values := make(map[string]struct{}, len(left))
	for _, value := range left {
		values[strings.TrimSpace(value)] = struct{}{}
	}
	for _, value := range right {
		if _, ok := values[strings.TrimSpace(value)]; !ok {
			return false
		}
	}
	return true
}

func decodeTemplateVariables(value []byte) ([]string, error) {
	var variables []string
	if err := json.Unmarshal(value, &variables); err != nil {
		return nil, err
	}
	return variables, nil
}
