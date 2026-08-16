package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

var (
	ErrInvalidOpsProjectBinding     = errors.New("invalid operations project binding")
	ErrInvalidOpsProjectEnvironment = errors.New("invalid operations project environment")
)

type OpsProjectBindingInput struct {
	Name                    string `json:"name"`
	VPSBindingID            uint   `json:"vps_binding_id"`
	ConnectorID             *uint  `json:"connector_id"`
	ConnectorIDSet          bool   `json:"-"`
	ProviderResourceID      string `json:"provider_resource_id"`
	Environment             string `json:"environment"`
	ComposeSource           string `json:"compose_source"`
	ComposeProjectName      string `json:"compose_project_name"`
	GatewayNetwork          string `json:"gateway_network"`
	GatewayAlias            string `json:"gateway_alias"`
	Services                string `json:"services"`
	Networks                string `json:"networks"`
	Volumes                 string `json:"volumes"`
	CurrentImageTag         string `json:"current_image_tag"`
	CurrentCommitSHA        string `json:"current_commit_sha"`
	Status                  string `json:"status"`
	Enabled                 *bool  `json:"enabled"`
	LastDeploymentAt        string `json:"last_deployment_at"`
	BackupPolicy            string `json:"backup_policy"`
	RestoreNotes            string `json:"restore_notes"`
	QuickBuyRateLimitPolicy string `json:"quick_buy_rate_limit_policy"`
	Notes                   string `json:"notes"`
}

type OpsProjectBindingService struct {
	repo          *repository.OpsProjectBindingRepository
	vpsRepo       *repository.OpsVPSBindingRepository
	connectorRepo *repository.OpsConnectorRepository
}

func NewOpsProjectBindingService(
	repo *repository.OpsProjectBindingRepository,
	vpsRepo *repository.OpsVPSBindingRepository,
	connectorRepo *repository.OpsConnectorRepository,
) *OpsProjectBindingService {
	return &OpsProjectBindingService{
		repo:          repo,
		vpsRepo:       vpsRepo,
		connectorRepo: connectorRepo,
	}
}

func (s *OpsProjectBindingService) List() ([]ops.ProjectBindingView, error) {
	return s.ListForEnvironment("")
}

func (s *OpsProjectBindingService) ListForEnvironment(environment string) ([]ops.ProjectBindingView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations project service is not configured")
	}
	environment, err := normalizeOpsProjectEnvironment(environment)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByEnvironment(environment)
}

func (s *OpsProjectBindingService) Get(id uint) (*ops.ProjectBindingView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations project service is not configured")
	}
	return s.repo.FindByID(id)
}

func (s *OpsProjectBindingService) Create(input OpsProjectBindingInput) (*ops.ProjectBindingView, error) {
	if s == nil || s.repo == nil || s.vpsRepo == nil {
		return nil, errors.New("operations project service is not configured")
	}
	record, err := s.normalizeInput(input, nil)
	if err != nil {
		return nil, err
	}
	vps, err := s.vpsRepo.FindByID(record.VPSBindingID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, fmt.Errorf("%w: VPS binding does not exist", ErrInvalidOpsProjectBinding)
		}
		return nil, err
	}
	if err := validateProjectVPSBinding(record.Environment, *vps); err != nil {
		return nil, err
	}
	if err := s.validateProjectConnector(record.ConnectorID, *vps, record.Environment); err != nil {
		return nil, err
	}
	if existing, err := s.repo.FindByName(record.Name); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: project name already exists", ErrInvalidOpsProjectBinding)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Create(&record); err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, fmt.Errorf("%w: project name already exists", ErrInvalidOpsProjectBinding)
		}
		return nil, err
	}
	return s.repo.FindByID(record.ID)
}

func (s *OpsProjectBindingService) Update(id uint, input OpsProjectBindingInput) (*ops.ProjectBindingView, error) {
	if s == nil || s.repo == nil || s.vpsRepo == nil {
		return nil, errors.New("operations project service is not configured")
	}
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	record, err := s.normalizeInput(input, &existing.ProjectBinding)
	if err != nil {
		return nil, err
	}
	record.ID = id
	vps, err := s.vpsRepo.FindByID(record.VPSBindingID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, fmt.Errorf("%w: VPS binding does not exist", ErrInvalidOpsProjectBinding)
		}
		return nil, err
	}
	if err := validateProjectVPSBinding(record.Environment, *vps); err != nil {
		return nil, err
	}
	if err := s.validateProjectConnector(record.ConnectorID, *vps, record.Environment); err != nil {
		return nil, err
	}
	if other, err := s.repo.FindByName(record.Name); err == nil && other != nil && other.ID != id {
		return nil, fmt.Errorf("%w: project name already exists", ErrInvalidOpsProjectBinding)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Update(&record); err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, fmt.Errorf("%w: project name already exists", ErrInvalidOpsProjectBinding)
		}
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *OpsProjectBindingService) SetEnabled(id uint, enabled bool) (*ops.ProjectBindingView, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations project service is not configured")
	}
	status := ops.ProjectStatusDisabled
	if enabled {
		status = ops.ProjectStatusPending
	}
	if err := s.repo.UpdateEnabled(id, enabled, status); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *OpsProjectBindingService) normalizeInput(input OpsProjectBindingInput, existing *ops.ProjectBinding) (ops.ProjectBinding, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 120 {
		return ops.ProjectBinding{}, fmt.Errorf("%w: name is required and must be at most 120 characters", ErrInvalidOpsProjectBinding)
	}
	if input.VPSBindingID == 0 {
		return ops.ProjectBinding{}, fmt.Errorf("%w: VPS binding is required", ErrInvalidOpsProjectBinding)
	}

	environment := normalizeConnectorEnum(input.Environment, map[string]struct{}{
		ops.ProjectEnvironmentProduction: {},
		ops.ProjectEnvironmentStaging:    {},
		ops.ProjectEnvironmentTest:       {},
		ops.ProjectEnvironmentLocal:      {},
	})
	if environment == "" {
		environment = ops.ProjectEnvironmentProduction
	}

	status := normalizeConnectorEnum(input.Status, map[string]struct{}{
		ops.ProjectStatusActive:   {},
		ops.ProjectStatusPending:  {},
		ops.ProjectStatusDisabled: {},
		ops.ProjectStatusDrifted:  {},
		ops.ProjectStatusError:    {},
	})
	if status == "" {
		status = ops.ProjectStatusPending
	}

	enabled := true
	if existing != nil {
		enabled = existing.Enabled
	}
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	if !enabled {
		status = ops.ProjectStatusDisabled
	}
	if enabled && status == ops.ProjectStatusDisabled {
		status = ops.ProjectStatusPending
	}

	lastDeploymentAt, err := parseOpsOptionalTime(input.LastDeploymentAt)
	if err != nil {
		return ops.ProjectBinding{}, fmt.Errorf("%w: last deployment time is invalid", ErrInvalidOpsProjectBinding)
	}
	if existing != nil {
		if lastDeploymentAt == nil {
			lastDeploymentAt = existing.LastDeploymentAt
		}
	}

	quickBuyRateLimitPolicy := strings.TrimSpace(input.QuickBuyRateLimitPolicy)
	if quickBuyRateLimitPolicy == "" && existing != nil && strings.TrimSpace(existing.QuickBuyRateLimitPolicy) != "" {
		quickBuyRateLimitPolicy = existing.QuickBuyRateLimitPolicy
	}
	quickBuyRateLimitPolicy, _, err = ops.NormalizeQuickBuyRateLimitPolicyJSON(quickBuyRateLimitPolicy)
	if err != nil {
		return ops.ProjectBinding{}, fmt.Errorf("%w: %v", ErrInvalidOpsProjectBinding, err)
	}

	record := ops.ProjectBinding{
		Name:                    name,
		VPSBindingID:            input.VPSBindingID,
		ConnectorID:             input.ConnectorID,
		ProviderResourceID:      strings.TrimSpace(input.ProviderResourceID),
		Environment:             environment,
		ComposeSource:           strings.TrimSpace(input.ComposeSource),
		ComposeProjectName:      strings.TrimSpace(input.ComposeProjectName),
		GatewayNetwork:          strings.TrimSpace(input.GatewayNetwork),
		GatewayAlias:            strings.TrimSpace(input.GatewayAlias),
		Services:                normalizeOpsList(input.Services),
		Networks:                normalizeOpsList(input.Networks),
		Volumes:                 normalizeOpsList(input.Volumes),
		CurrentImageTag:         strings.TrimSpace(input.CurrentImageTag),
		CurrentCommitSHA:        strings.TrimSpace(input.CurrentCommitSHA),
		Status:                  status,
		HealthStatus:            ops.ProjectHealthUnknown,
		Enabled:                 enabled,
		LastDeploymentAt:        lastDeploymentAt,
		BackupPolicy:            strings.TrimSpace(input.BackupPolicy),
		RestoreNotes:            strings.TrimSpace(input.RestoreNotes),
		QuickBuyRateLimitPolicy: quickBuyRateLimitPolicy,
		Notes:                   strings.TrimSpace(input.Notes),
	}
	if existing != nil {
		record.ID = existing.ID
		if !input.ConnectorIDSet {
			record.ConnectorID = existing.ConnectorID
		}
	}
	return record, nil
}

func validateProjectVPSBinding(environment string, vps ops.VPSBinding) error {
	if strings.TrimSpace(vps.Environment) != strings.TrimSpace(environment) {
		return fmt.Errorf(
			"%w: project environment %s does not match VPS environment %s",
			ErrInvalidOpsProjectBinding,
			environment,
			vps.Environment,
		)
	}
	return nil
}

func (s *OpsProjectBindingService) validateProjectConnector(
	connectorID *uint,
	vps ops.VPSBinding,
	environment string,
) error {
	if connectorID == nil || *connectorID == 0 {
		connectorID = vps.ConnectorID
	}
	if connectorID == nil || *connectorID == 0 {
		return nil
	}
	if s.connectorRepo == nil {
		return errors.New("operations connector repository is not configured")
	}
	connector, err := s.connectorRepo.FindByID(*connectorID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return fmt.Errorf("%w: effective project connector does not exist", ErrInvalidOpsProjectBinding)
		}
		return err
	}
	if connector.Provider != ops.ConnectorProviderHostinger {
		return fmt.Errorf("%w: effective project connector must be Hostinger", ErrInvalidOpsProjectBinding)
	}
	if connector.Environment != "" && connector.Environment != environment {
		return fmt.Errorf(
			"%w: project environment %s does not match connector environment %s",
			ErrInvalidOpsProjectBinding,
			environment,
			connector.Environment,
		)
	}
	return nil
}

func normalizeOpsList(raw string) string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	values := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return strings.Join(values, ", ")
}

func parseOpsOptionalTime(raw string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			utc := parsed.UTC()
			return &utc, nil
		}
	}
	return nil, errors.New("invalid time")
}

func normalizeOpsProjectEnvironment(environment string) (string, error) {
	environment = strings.ToLower(strings.TrimSpace(environment))
	if environment == "" {
		return "", nil
	}
	switch environment {
	case ops.ProjectEnvironmentProduction,
		ops.ProjectEnvironmentStaging,
		ops.ProjectEnvironmentTest,
		ops.ProjectEnvironmentLocal:
		return environment, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidOpsProjectEnvironment, environment)
	}
}
