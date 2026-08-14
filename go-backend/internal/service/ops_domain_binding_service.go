package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

var ErrInvalidOpsDomainBinding = errors.New("invalid operations domain binding")

type OpsDomainBindingService struct {
	repo          *repository.OpsDomainBindingRepository
	projectRepo   *repository.OpsProjectBindingRepository
	connectorRepo *repository.OpsConnectorRepository
}

type OpsDomainBindingInput struct {
	Domain              string `json:"domain"`
	ConnectorID         *uint  `json:"connector_id"`
	ConnectorIDSet      bool   `json:"-"`
	ProjectBindingID    *uint  `json:"project_binding_id"`
	ProjectBindingIDSet bool   `json:"-"`
	Role                string `json:"role"`
	Environment         string `json:"environment"`
	Provider            string `json:"provider"`
	Zone                string `json:"zone"`
	Target              string `json:"target"`
	ProxyMode           string `json:"proxy_mode"`
	TLSMode             string `json:"tls_mode"`
	RedirectTarget      string `json:"redirect_target"`
	Status              string `json:"status"`
	Enabled             *bool  `json:"enabled"`
	Notes               string `json:"notes"`
}

func NewOpsDomainBindingService(
	repo *repository.OpsDomainBindingRepository,
	projectRepo *repository.OpsProjectBindingRepository,
	connectorRepo *repository.OpsConnectorRepository,
) *OpsDomainBindingService {
	return &OpsDomainBindingService{
		repo:          repo,
		projectRepo:   projectRepo,
		connectorRepo: connectorRepo,
	}
}

func (s *OpsDomainBindingService) List() ([]ops.DomainBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations domain binding service is not configured")
	}
	return s.repo.List()
}

func (s *OpsDomainBindingService) Get(id uint) (*ops.DomainBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations domain binding service is not configured")
	}
	return s.repo.FindByID(id)
}

func (s *OpsDomainBindingService) Create(input OpsDomainBindingInput) (*ops.DomainBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations domain binding service is not configured")
	}
	record, err := normalizeOpsDomainBindingInput(input)
	if err != nil {
		return nil, err
	}
	if err := validateDeploymentDomainProject(record); err != nil {
		return nil, err
	}
	if err := s.validateProjectBinding(record.ProjectBindingID, record.Environment); err != nil {
		return nil, err
	}
	if err := s.validateConnectorBinding(record.ConnectorID, record.Provider, record.Environment); err != nil {
		return nil, err
	}
	if existing, err := s.repo.FindByDomain(record.Domain); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: domain already exists", ErrInvalidOpsDomainBinding)
	} else if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Create(&record); err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, fmt.Errorf("%w: domain already exists", ErrInvalidOpsDomainBinding)
		}
		return nil, err
	}
	return &record, nil
}

func (s *OpsDomainBindingService) Update(id uint, input OpsDomainBindingInput) (*ops.DomainBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations domain binding service is not configured")
	}
	existingRecord, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	record, err := normalizeOpsDomainBindingInput(input)
	if err != nil {
		return nil, err
	}
	record.ID = id
	if !input.ConnectorIDSet {
		record.ConnectorID = existingRecord.ConnectorID
	}
	if !input.ProjectBindingIDSet {
		record.ProjectBindingID = existingRecord.ProjectBindingID
	}
	if err := validateDeploymentDomainProject(record); err != nil {
		return nil, err
	}
	if err := s.validateProjectBinding(record.ProjectBindingID, record.Environment); err != nil {
		return nil, err
	}
	if err := s.validateConnectorBinding(record.ConnectorID, record.Provider, record.Environment); err != nil {
		return nil, err
	}
	existing, err := s.repo.FindByDomain(record.Domain)
	if err == nil && existing != nil && existing.ID != id {
		return nil, fmt.Errorf("%w: domain already exists", ErrInvalidOpsDomainBinding)
	}
	if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	if err := s.repo.Update(&record); err != nil {
		if repository.IsDuplicatedKey(err) {
			return nil, fmt.Errorf("%w: domain already exists", ErrInvalidOpsDomainBinding)
		}
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *OpsDomainBindingService) validateProjectBinding(projectID *uint, environment string) error {
	if projectID == nil || *projectID == 0 {
		return nil
	}
	if s.projectRepo == nil {
		return errors.New("operations project binding repository is not configured")
	}
	project, err := s.projectRepo.FindByID(*projectID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return fmt.Errorf("%w: project binding does not exist", ErrInvalidOpsDomainBinding)
		}
		return err
	}
	if project.Environment != environment {
		return fmt.Errorf(
			"%w: project environment %s does not match domain environment %s",
			ErrInvalidOpsDomainBinding,
			project.Environment,
			environment,
		)
	}
	return nil
}

func (s *OpsDomainBindingService) validateConnectorBinding(connectorID *uint, provider, environment string) error {
	if connectorID == nil || *connectorID == 0 {
		return nil
	}
	if s.connectorRepo == nil {
		return errors.New("operations connector repository is not configured")
	}
	connector, err := s.connectorRepo.FindByID(*connectorID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return fmt.Errorf("%w: connector does not exist", ErrInvalidOpsDomainBinding)
		}
		return err
	}
	if connector.Provider != provider {
		return fmt.Errorf(
			"%w: domain provider %s does not match connector provider %s",
			ErrInvalidOpsDomainBinding,
			provider,
			connector.Provider,
		)
	}
	if connector.Environment != "" && connector.Environment != environment {
		return fmt.Errorf(
			"%w: domain environment %s does not match connector environment %s",
			ErrInvalidOpsDomainBinding,
			environment,
			connector.Environment,
		)
	}
	return nil
}

func (s *OpsDomainBindingService) SetEnabled(id uint, enabled bool) (*ops.DomainBinding, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("operations domain binding service is not configured")
	}
	status := ops.DomainStatusDisabled
	if enabled {
		status = ops.DomainStatusActive
	}
	if err := s.repo.UpdateEnabled(id, enabled, status); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func normalizeOpsDomainBindingInput(input OpsDomainBindingInput) (ops.DomainBinding, error) {
	domain, err := normalizeDomain(input.Domain)
	if err != nil {
		return ops.DomainBinding{}, err
	}

	role := normalizeEnum(input.Role, map[string]struct{}{
		ops.DomainRoleCanonical:    {},
		ops.DomainRoleAlias:        {},
		ops.DomainRoleAdmin:        {},
		ops.DomainRoleRedirect:     {},
		ops.DomainRoleVerification: {},
		ops.DomainRoleInternal:     {},
	})
	if role == "" {
		return ops.DomainBinding{}, fmt.Errorf("%w: role is invalid", ErrInvalidOpsDomainBinding)
	}

	environment := normalizeEnum(input.Environment, map[string]struct{}{
		ops.DomainEnvironmentProduction: {},
		ops.DomainEnvironmentStaging:    {},
		ops.DomainEnvironmentTest:       {},
		ops.DomainEnvironmentLocal:      {},
	})
	if environment == "" {
		return ops.DomainBinding{}, fmt.Errorf("%w: environment is invalid", ErrInvalidOpsDomainBinding)
	}

	provider := normalizeEnum(input.Provider, map[string]struct{}{
		ops.DomainProviderCloudflare: {},
		ops.DomainProviderHostinger:  {},
		ops.DomainProviderOther:      {},
	})
	if provider == "" {
		return ops.DomainBinding{}, fmt.Errorf("%w: provider is invalid", ErrInvalidOpsDomainBinding)
	}

	proxyMode := normalizeEnum(input.ProxyMode, map[string]struct{}{
		ops.DomainProxyProxied: {},
		ops.DomainProxyDNSOnly: {},
		ops.DomainProxyUnknown: {},
	})
	if proxyMode == "" {
		proxyMode = ops.DomainProxyUnknown
	}

	tlsMode := normalizeEnum(input.TLSMode, map[string]struct{}{
		ops.DomainTLSFullStrict: {},
		ops.DomainTLSFull:       {},
		ops.DomainTLSFlexible:   {},
		ops.DomainTLSOff:        {},
		ops.DomainTLSUnknown:    {},
	})
	if tlsMode == "" {
		tlsMode = ops.DomainTLSUnknown
	}

	status := normalizeEnum(input.Status, map[string]struct{}{
		ops.DomainStatusActive:   {},
		ops.DomainStatusPending:  {},
		ops.DomainStatusDisabled: {},
		ops.DomainStatusDrifted:  {},
		ops.DomainStatusError:    {},
	})
	if status == "" {
		status = ops.DomainStatusPending
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	if !enabled {
		status = ops.DomainStatusDisabled
	}
	if enabled && status == ops.DomainStatusDisabled {
		status = ops.DomainStatusActive
	}

	redirectTarget := strings.TrimSpace(input.RedirectTarget)
	if role == ops.DomainRoleRedirect && redirectTarget == "" {
		return ops.DomainBinding{}, fmt.Errorf("%w: redirect_target is required for redirect domains", ErrInvalidOpsDomainBinding)
	}
	projectBindingID := normalizeOpsDomainBindingID(input.ProjectBindingID)

	return ops.DomainBinding{
		Domain:           domain,
		ConnectorID:      normalizeOpsDomainBindingID(input.ConnectorID),
		ProjectBindingID: projectBindingID,
		Role:             role,
		Environment:      environment,
		Provider:         provider,
		Zone:             strings.ToLower(strings.TrimSpace(input.Zone)),
		Target:           strings.TrimSpace(input.Target),
		ProxyMode:        proxyMode,
		TLSMode:          tlsMode,
		RedirectTarget:   redirectTarget,
		Status:           status,
		Enabled:          enabled,
		Notes:            strings.TrimSpace(input.Notes),
	}, nil
}

func validateDeploymentDomainProject(record ops.DomainBinding) error {
	if isDeploymentDomainRole(record.Role) && record.ProjectBindingID == nil {
		return fmt.Errorf("%w: project_binding_id is required for deployment domains", ErrInvalidOpsDomainBinding)
	}
	return nil
}

func normalizeOpsDomainBindingID(id *uint) *uint {
	if id == nil || *id == 0 {
		return nil
	}
	value := *id
	return &value
}

func isDeploymentDomainRole(role string) bool {
	switch role {
	case ops.DomainRoleCanonical, ops.DomainRoleAlias, ops.DomainRoleAdmin, ops.DomainRoleRedirect:
		return true
	default:
		return false
	}
}

func normalizeDomain(raw string) (string, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return "", fmt.Errorf("%w: domain is required", ErrInvalidOpsDomainBinding)
	}
	if strings.ContainsAny(value, " \t\r\n/") {
		return "", fmt.Errorf("%w: domain must be a hostname", ErrInvalidOpsDomainBinding)
	}
	if strings.Contains(value, "://") {
		return "", fmt.Errorf("%w: domain must not include a scheme", ErrInvalidOpsDomainBinding)
	}
	value = strings.TrimSuffix(value, ".")
	if value == "" || len(value) > 255 {
		return "", fmt.Errorf("%w: domain length is invalid", ErrInvalidOpsDomainBinding)
	}
	if parsed, err := url.Parse("https://" + value); err != nil || parsed.Host != value {
		return "", fmt.Errorf("%w: domain is invalid", ErrInvalidOpsDomainBinding)
	}
	return value, nil
}

func normalizeEnum(value string, allowed map[string]struct{}) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if _, ok := allowed[normalized]; !ok {
		return ""
	}
	return normalized
}
