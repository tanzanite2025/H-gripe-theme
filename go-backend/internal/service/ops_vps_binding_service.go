package service

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

var ErrInvalidOpsVPSBinding = errors.New("invalid operations VPS binding")

type OpsVPSBindingInput struct {
	Name               string `json:"name"`
	Provider           string `json:"provider"`
	Environment        string `json:"environment"`
	ConnectorID        *uint  `json:"connector_id"`
	ConnectorIDSet     bool   `json:"-"`
	ProviderResourceID string `json:"provider_resource_id"`
	Hostname           string `json:"hostname"`
	IPv4               string `json:"ipv4"`
	Region             string `json:"region"`
	OperatingSystem    string `json:"operating_system"`
	Status             string `json:"status"`
	Enabled            *bool  `json:"enabled"`
	Notes              string `json:"notes"`
}

type OpsVPSBindingService struct {
	repo          *repository.OpsVPSBindingRepository
	connectorRepo *repository.OpsConnectorRepository
}

func NewOpsVPSBindingService(
	repo *repository.OpsVPSBindingRepository,
	connectorRepo *repository.OpsConnectorRepository,
) *OpsVPSBindingService {
	return &OpsVPSBindingService{repo: repo, connectorRepo: connectorRepo}
}

func (s *OpsVPSBindingService) List() ([]ops.VPSBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations VPS service is not configured")
	}
	return s.repo.List()
}

func (s *OpsVPSBindingService) Get(id uint) (*ops.VPSBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations VPS service is not configured")
	}
	return s.repo.FindByID(id)
}

func (s *OpsVPSBindingService) Create(input OpsVPSBindingInput) (*ops.VPSBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations VPS service is not configured")
	}
	record, err := normalizeOpsVPSBindingInput(input, nil)
	if err != nil {
		return nil, err
	}
	if err := s.validateConnector(record.ConnectorID, record.Provider, record.Environment); err != nil {
		return nil, err
	}
	if existing, err := s.repo.FindByName(record.Name); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: VPS name already exists", ErrInvalidOpsVPSBinding)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Create(&record); err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, fmt.Errorf("%w: VPS name already exists", ErrInvalidOpsVPSBinding)
		}
		return nil, err
	}
	return &record, nil
}

func (s *OpsVPSBindingService) Update(id uint, input OpsVPSBindingInput) (*ops.VPSBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations VPS service is not configured")
	}
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	record, err := normalizeOpsVPSBindingInput(input, existing)
	if err != nil {
		return nil, err
	}
	record.ID = id
	if err := s.validateConnector(record.ConnectorID, record.Provider, record.Environment); err != nil {
		return nil, err
	}
	if other, err := s.repo.FindByName(record.Name); err == nil && other != nil && other.ID != id {
		return nil, fmt.Errorf("%w: VPS name already exists", ErrInvalidOpsVPSBinding)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Update(&record); err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, fmt.Errorf("%w: VPS name already exists", ErrInvalidOpsVPSBinding)
		}
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *OpsVPSBindingService) SetEnabled(id uint, enabled bool) (*ops.VPSBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations VPS service is not configured")
	}
	status := ops.VPSStatusDisabled
	if enabled {
		status = ops.VPSStatusPending
	}
	if err := s.repo.UpdateEnabled(id, enabled, status); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func normalizeOpsVPSBindingInput(input OpsVPSBindingInput, existing *ops.VPSBinding) (ops.VPSBinding, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 120 {
		return ops.VPSBinding{}, fmt.Errorf("%w: name is required and must be at most 120 characters", ErrInvalidOpsVPSBinding)
	}

	provider := normalizeConnectorEnum(input.Provider, map[string]struct{}{
		ops.VPSProviderHostinger: {},
		ops.VPSProviderOther:     {},
	})
	if provider == "" {
		return ops.VPSBinding{}, fmt.Errorf("%w: provider is invalid", ErrInvalidOpsVPSBinding)
	}

	environment := normalizeConnectorEnum(input.Environment, map[string]struct{}{
		ops.VPSEnvironmentProduction: {},
		ops.VPSEnvironmentStaging:    {},
		ops.VPSEnvironmentTest:       {},
		ops.VPSEnvironmentLocal:      {},
	})
	if environment == "" {
		environment = ops.VPSEnvironmentProduction
	}

	status := normalizeConnectorEnum(input.Status, map[string]struct{}{
		ops.VPSStatusActive:   {},
		ops.VPSStatusPending:  {},
		ops.VPSStatusDisabled: {},
		ops.VPSStatusDrifted:  {},
		ops.VPSStatusError:    {},
	})
	if status == "" {
		status = ops.VPSStatusPending
	}

	enabled := true
	if existing != nil {
		enabled = existing.Enabled
	}
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	if !enabled {
		status = ops.VPSStatusDisabled
	}
	if enabled && status == ops.VPSStatusDisabled {
		status = ops.VPSStatusPending
	}

	hostname, err := normalizeOpsHostname(input.Hostname)
	if err != nil {
		return ops.VPSBinding{}, err
	}
	ipv4 := strings.TrimSpace(input.IPv4)
	if ipv4 != "" {
		parsed := net.ParseIP(ipv4)
		if parsed == nil || parsed.To4() == nil {
			return ops.VPSBinding{}, fmt.Errorf("%w: ipv4 is invalid", ErrInvalidOpsVPSBinding)
		}
		ipv4 = parsed.To4().String()
	}

	record := ops.VPSBinding{
		Name:               name,
		Provider:           provider,
		Environment:        environment,
		ConnectorID:        input.ConnectorID,
		ProviderResourceID: strings.TrimSpace(input.ProviderResourceID),
		Hostname:           hostname,
		IPv4:               ipv4,
		Region:             strings.TrimSpace(input.Region),
		OperatingSystem:    strings.TrimSpace(input.OperatingSystem),
		Status:             status,
		ObservedStatus:     ops.VPSObservedUnknown,
		Enabled:            enabled,
		Notes:              strings.TrimSpace(input.Notes),
	}
	if existing != nil {
		record.ID = existing.ID
		record.LastObservedAt = existing.LastObservedAt
		if !input.ConnectorIDSet {
			record.ConnectorID = existing.ConnectorID
		}
	}
	return record, nil
}

func (s *OpsVPSBindingService) validateConnector(connectorID *uint, provider, environment string) error {
	if connectorID == nil || *connectorID == 0 {
		return nil
	}
	if s.connectorRepo == nil {
		return errors.New("operations connector repository is not configured")
	}
	connector, err := s.connectorRepo.FindByID(*connectorID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return fmt.Errorf("%w: VPS connector does not exist", ErrInvalidOpsVPSBinding)
		}
		return err
	}
	if connector.Provider != provider {
		return fmt.Errorf(
			"%w: VPS provider %s does not match connector provider %s",
			ErrInvalidOpsVPSBinding,
			provider,
			connector.Provider,
		)
	}
	if connector.Environment != "" && connector.Environment != environment {
		return fmt.Errorf(
			"%w: VPS environment %s does not match connector environment %s",
			ErrInvalidOpsVPSBinding,
			environment,
			connector.Environment,
		)
	}
	return nil
}

func normalizeOpsHostname(raw string) (string, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return "", nil
	}
	if len(value) > 255 || strings.ContainsAny(value, " \t\r\n/:") || strings.Contains(value, "://") {
		return "", fmt.Errorf("%w: hostname is invalid", ErrInvalidOpsVPSBinding)
	}
	return strings.TrimSuffix(value, "."), nil
}
