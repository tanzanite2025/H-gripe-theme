package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	selectionconfiguration "commerce-platform/internal/domain/selectionconfiguration"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrSelectionConfigurationKeyInvalid         = errors.New("invalid selection configuration key")
	ErrSelectionConfigurationKeyNotFound        = errors.New("selection configuration key not found")
	ErrSelectionConfigurationKeyAlreadyExists   = errors.New("selection configuration key already exists")
	ErrSelectionConfigurationKeyCodeImmutable   = errors.New("selection configuration key code is immutable")
	ErrSelectionConfigurationKeyKindUnsupported = errors.New("selection configuration key kind is unsupported")
)

const (
	SelectionConfigurationKeyKindQuestionKey = selectionconfiguration.SelectionConfigurationKeyKindQuestionKey
	SelectionConfigurationKeyKindAnswerKey   = selectionconfiguration.SelectionConfigurationKeyKindAnswerKey
)

var selectionConfigurationKeyCodePattern = regexp.MustCompile(`^[a-z0-9]+(?:_[a-z0-9]+)*$`)

type SelectionConfigurationKeyService struct {
	repo *repository.SelectionConfigurationKeyRepository
}

type SelectionConfigurationKeyInput struct {
	ID           uint
	Kind         string
	Code         string
	DisplayLabel string
	Description  string
	IsEnabled    bool
	SortOrder    int
}

type SelectionConfigurationKeyOption struct {
	ID           uint   `json:"id"`
	Code         string `json:"code"`
	DisplayLabel string `json:"display_label"`
}

type SelectionConfigurationKeyListItem struct {
	ID           uint   `json:"id"`
	Kind         string `json:"kind"`
	Code         string `json:"code"`
	DisplayLabel string `json:"display_label"`
	Description  string `json:"description"`
	IsEnabled    bool   `json:"is_enabled"`
	SortOrder    int    `json:"sort_order"`
}

func NewSelectionConfigurationKeyService(repo *repository.SelectionConfigurationKeyRepository) *SelectionConfigurationKeyService {
	return &SelectionConfigurationKeyService{repo: repo}
}

func (s *SelectionConfigurationKeyService) ListSelectionConfigurationKeys(kind string, includeDisabled bool) ([]SelectionConfigurationKeyListItem, error) {
	normalizedKind, err := normalizeSelectionConfigurationKeyKind(kind)
	if err != nil {
		return nil, err
	}
	keys, err := s.repo.ListSelectionConfigurationKeysByKind(normalizedKind, includeDisabled)
	if err != nil {
		return nil, err
	}
	result := make([]SelectionConfigurationKeyListItem, 0, len(keys))
	for _, key := range keys {
		result = append(result, SelectionConfigurationKeyListItem{
			ID:           key.ID,
			Kind:         key.Kind,
			Code:         key.Code,
			DisplayLabel: key.DisplayLabel,
			Description:  key.Description,
			IsEnabled:    key.IsEnabled,
			SortOrder:    key.SortOrder,
		})
	}
	return result, nil
}

func (s *SelectionConfigurationKeyService) ListEnabledSelectionConfigurationKeyOptions(kind string) ([]SelectionConfigurationKeyOption, error) {
	normalizedKind, err := normalizeSelectionConfigurationKeyKind(kind)
	if err != nil {
		return nil, err
	}
	keys, err := s.repo.ListSelectionConfigurationKeysByKind(normalizedKind, false)
	if err != nil {
		return nil, err
	}
	result := make([]SelectionConfigurationKeyOption, 0, len(keys))
	for _, key := range keys {
		result = append(result, SelectionConfigurationKeyOption{
			ID:           key.ID,
			Code:         key.Code,
			DisplayLabel: key.DisplayLabel,
		})
	}
	return result, nil
}

func (s *SelectionConfigurationKeyService) CreateSelectionConfigurationKey(input SelectionConfigurationKeyInput) (*selectionconfiguration.SelectionConfigurationKey, error) {
	normalized, err := normalizeSelectionConfigurationKeyInput(input)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.FindSelectionConfigurationKeyByKindAndCode(normalized.Kind, normalized.Code); err == nil {
		return nil, ErrSelectionConfigurationKeyAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	key := &selectionconfiguration.SelectionConfigurationKey{
		Kind:         normalized.Kind,
		Code:         normalized.Code,
		DisplayLabel: normalized.DisplayLabel,
		Description:  normalized.Description,
		IsEnabled:    normalized.IsEnabled,
		SortOrder:    normalized.SortOrder,
	}
	if err := s.repo.CreateSelectionConfigurationKey(key); err != nil {
		return nil, err
	}
	return key, nil
}

func (s *SelectionConfigurationKeyService) UpdateSelectionConfigurationKey(id uint, input SelectionConfigurationKeyInput) (*selectionconfiguration.SelectionConfigurationKey, error) {
	if id == 0 {
		return nil, fmt.Errorf("%w: key id is required", ErrSelectionConfigurationKeyInvalid)
	}
	normalized, err := normalizeSelectionConfigurationKeyInput(input)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.FindSelectionConfigurationKeyByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSelectionConfigurationKeyNotFound
		}
		return nil, err
	}
	if existing.Kind != normalized.Kind {
		return nil, ErrSelectionConfigurationKeyCodeImmutable
	}
	if existing.Code != normalized.Code {
		return nil, ErrSelectionConfigurationKeyCodeImmutable
	}

	existing.DisplayLabel = normalized.DisplayLabel
	existing.Description = normalized.Description
	existing.IsEnabled = normalized.IsEnabled
	existing.SortOrder = normalized.SortOrder
	if err := s.repo.SaveSelectionConfigurationKey(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func normalizeSelectionConfigurationKeyInput(input SelectionConfigurationKeyInput) (SelectionConfigurationKeyInput, error) {
	kind, err := normalizeSelectionConfigurationKeyKind(input.Kind)
	if err != nil {
		return SelectionConfigurationKeyInput{}, err
	}
	code := normalizeSelectionConfigurationKeyCode(input.Code)
	if code == "" {
		return SelectionConfigurationKeyInput{}, fmt.Errorf("%w: code is required and must use lowercase snake_case", ErrSelectionConfigurationKeyInvalid)
	}
	displayLabel := strings.TrimSpace(input.DisplayLabel)
	if displayLabel == "" {
		return SelectionConfigurationKeyInput{}, fmt.Errorf("%w: display_label is required", ErrSelectionConfigurationKeyInvalid)
	}
	if input.SortOrder <= 0 {
		input.SortOrder = 10
	}
	input.Kind = kind
	input.Code = code
	input.DisplayLabel = displayLabel
	input.Description = strings.TrimSpace(input.Description)
	return input, nil
}

func normalizeSelectionConfigurationKeyKind(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case selectionconfiguration.SelectionConfigurationKeyKindQuestionKey, selectionconfiguration.SelectionConfigurationKeyKindAnswerKey:
		return value, nil
	default:
		return "", fmt.Errorf("%w: kind must be %q or %q", ErrSelectionConfigurationKeyKindUnsupported, selectionconfiguration.SelectionConfigurationKeyKindQuestionKey, selectionconfiguration.SelectionConfigurationKeyKindAnswerKey)
	}
}

func normalizeSelectionConfigurationKeyCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if !selectionConfigurationKeyCodePattern.MatchString(value) {
		return ""
	}
	return value
}
