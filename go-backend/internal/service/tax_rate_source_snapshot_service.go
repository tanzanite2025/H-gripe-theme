package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/repository"
)

const (
	vatcomplyTaxRateProviderCode      = taxrate.DefaultTaxRateSourceProviderCode
	taxRateSourceScheduledRetryWindow = time.Hour
)

var (
	ErrTaxRateSourceDisabled       = errors.New("tax rate source synchronization is disabled")
	ErrTaxRateSourceSyncInProgress = errors.New("tax rate source synchronization is already running")
	ErrTaxRateSourceConfigInvalid  = errors.New("tax rate source configuration is invalid")
)

// TaxRateSourceSnapshotService owns provider configuration, external data
// validation, version creation and read-only snapshot access. It deliberately
// does not write the tax_rates records used by checkout.
type TaxRateSourceSnapshotService struct {
	repository      *repository.TaxRateSourceSnapshotRepository
	providerAdapter *vatcomplyTaxRateSourceAdapter
	syncMu          sync.Mutex
}

type TaxRateSourceConfiguration struct {
	ProviderCode         string     `json:"provider_code"`
	ProviderName         string     `json:"provider_name"`
	Endpoint             string     `json:"endpoint"`
	DocumentationURL     string     `json:"documentation_url"`
	Enabled              bool       `json:"enabled"`
	RefreshIntervalHours int        `json:"refresh_interval_hours"`
	RequiresAPIKey       bool       `json:"requires_api_key"`
	LastCheckedAt        *time.Time `json:"last_checked_at,omitempty"`
	LastSuccessfulSyncAt *time.Time `json:"last_successful_sync_at,omitempty"`
	LastSyncError        string     `json:"last_sync_error,omitempty"`
	CoverageDescription  string     `json:"coverage_description"`
}

type TaxRateSourceConfigurationUpdate struct {
	Enabled              bool `json:"enabled"`
	RefreshIntervalHours int  `json:"refresh_interval_hours"`
}

type CurrentTaxRateSourceSnapshot struct {
	Snapshot *taxrate.TaxRateSourceSnapshot       `json:"snapshot"`
	Entries  []taxrate.TaxRateSourceSnapshotEntry `json:"entries"`
}

type TaxRateSourceSnapshotSyncResult struct {
	Snapshot  *taxrate.TaxRateSourceSnapshot `json:"snapshot"`
	Changed   bool                           `json:"changed"`
	CheckedAt time.Time                      `json:"checked_at"`
}

func NewTaxRateSourceSnapshotService(
	snapshotRepository *repository.TaxRateSourceSnapshotRepository,
) *TaxRateSourceSnapshotService {
	return &TaxRateSourceSnapshotService{
		repository:      snapshotRepository,
		providerAdapter: newVATComplyTaxRateSourceAdapter(),
	}
}

// ConfigureHTTPClient replaces the provider client for deterministic tests.
func (s *TaxRateSourceSnapshotService) ConfigureHTTPClient(client *http.Client) {
	if s == nil || client == nil {
		return
	}
	if s.providerAdapter != nil {
		s.providerAdapter.httpClient = client
	}
}

func (s *TaxRateSourceSnapshotService) GetSourceConfiguration() (TaxRateSourceConfiguration, error) {
	if s == nil || s.repository == nil {
		return TaxRateSourceConfiguration{}, errors.New("tax rate source snapshot service is not configured")
	}
	record, err := s.repository.GetOrCreateSourceConfig()
	if err != nil {
		return TaxRateSourceConfiguration{}, err
	}
	return taxRateSourceConfigurationView(record), nil
}

func taxRateSourceConfigurationView(record *taxrate.TaxRateSourceConfig) TaxRateSourceConfiguration {
	if record == nil {
		return TaxRateSourceConfiguration{}
	}
	return TaxRateSourceConfiguration{
		ProviderCode:         record.ProviderCode,
		ProviderName:         vatcomplyTaxRateProviderName,
		Endpoint:             vatcomplyTaxRateAPIEndpoint,
		DocumentationURL:     vatcomplyTaxRateAPIDocURL,
		Enabled:              record.Enabled,
		RefreshIntervalHours: record.RefreshIntervalHours,
		RequiresAPIKey:       false,
		LastCheckedAt:        record.LastCheckedAt,
		LastSuccessfulSyncAt: record.LastSuccessfulSyncAt,
		LastSyncError:        record.LastSyncError,
		CoverageDescription:  "VATcomply currently provides VAT rates for EU member states. It does not cover every storefront market or determine which product-specific rate applies.",
	}
}

func (s *TaxRateSourceSnapshotService) UpdateSourceConfiguration(
	update TaxRateSourceConfigurationUpdate,
) (TaxRateSourceConfiguration, error) {
	if s == nil || s.repository == nil {
		return TaxRateSourceConfiguration{}, errors.New("tax rate source snapshot service is not configured")
	}
	if update.RefreshIntervalHours < 24 || update.RefreshIntervalHours > 8760 {
		return TaxRateSourceConfiguration{}, fmt.Errorf("%w: refresh interval must be from 24 to 8760 hours", ErrTaxRateSourceConfigInvalid)
	}
	if _, err := s.repository.UpdateSourceConfig(update.Enabled, update.RefreshIntervalHours); err != nil {
		return TaxRateSourceConfiguration{}, err
	}
	return s.GetSourceConfiguration()
}

func (s *TaxRateSourceSnapshotService) GetCurrentSnapshot() (CurrentTaxRateSourceSnapshot, error) {
	if s == nil || s.repository == nil {
		return CurrentTaxRateSourceSnapshot{}, errors.New("tax rate source snapshot service is not configured")
	}
	snapshot, err := s.repository.FindLatestSnapshot(vatcomplyTaxRateProviderCode)
	if err != nil {
		return CurrentTaxRateSourceSnapshot{}, err
	}
	if snapshot == nil {
		return CurrentTaxRateSourceSnapshot{Entries: []taxrate.TaxRateSourceSnapshotEntry{}}, nil
	}
	entries, err := s.repository.ListSnapshotEntries(snapshot.ID)
	if err != nil {
		return CurrentTaxRateSourceSnapshot{}, err
	}
	return CurrentTaxRateSourceSnapshot{Snapshot: snapshot, Entries: entries}, nil
}

func (s *TaxRateSourceSnapshotService) Sync(ctx context.Context) (*TaxRateSourceSnapshotSyncResult, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("tax rate source snapshot service is not configured")
	}
	if !s.syncMu.TryLock() {
		return nil, ErrTaxRateSourceSyncInProgress
	}
	defer s.syncMu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}

	config, err := s.repository.GetOrCreateSourceConfig()
	if err != nil {
		return nil, err
	}
	if config.ProviderCode != vatcomplyTaxRateProviderCode {
		return nil, fmt.Errorf("%w: unsupported provider %q", ErrTaxRateSourceConfigInvalid, config.ProviderCode)
	}
	if !config.Enabled {
		return nil, ErrTaxRateSourceDisabled
	}

	checkedAt := time.Now().UTC()
	if err := s.repository.RecordSyncStarted(checkedAt); err != nil {
		return nil, err
	}
	entries, countryCount, err := s.providerAdapter.FetchVATComplyTaxRateEntries(ctx, checkedAt)
	if err != nil {
		if statusErr := s.repository.RecordSyncFailure(time.Now().UTC(), err.Error()); statusErr != nil {
			return nil, fmt.Errorf("%w (record sync failure: %v)", err, statusErr)
		}
		return nil, err
	}

	contentHash, err := taxRateSourceSnapshotContentSHA256(entries)
	if err != nil {
		if statusErr := s.repository.RecordSyncFailure(time.Now().UTC(), err.Error()); statusErr != nil {
			return nil, fmt.Errorf("%w (record sync failure: %v)", err, statusErr)
		}
		return nil, err
	}
	snapshot := &taxrate.TaxRateSourceSnapshot{
		ProviderCode:   vatcomplyTaxRateProviderCode,
		SourceEndpoint: vatcomplyTaxRateAPIEndpoint,
		Version:        fmt.Sprintf("vatcomply-%s-%s", checkedAt.Format("20060102T150405.000000000Z"), contentHash[:12]),
		ContentSHA256:  contentHash,
		CapturedAt:     checkedAt,
		CountryCount:   countryCount,
		RateCount:      len(entries),
		CreatedAt:      checkedAt,
	}
	published, changed, err := s.repository.CreateSnapshotWhenContentChanges(snapshot, entries)
	if err != nil {
		if statusErr := s.repository.RecordSyncFailure(time.Now().UTC(), err.Error()); statusErr != nil {
			return nil, fmt.Errorf("save tax rate snapshot: %w (record sync failure: %v)", err, statusErr)
		}
		return nil, fmt.Errorf("save tax rate snapshot: %w", err)
	}
	completedAt := time.Now().UTC()
	if err := s.repository.RecordSyncSuccess(completedAt); err != nil {
		return nil, fmt.Errorf("record tax rate source sync success: %w", err)
	}
	return &TaxRateSourceSnapshotSyncResult{
		Snapshot:  published,
		Changed:   changed,
		CheckedAt: completedAt,
	}, nil
}

// SyncIfDue is used only by the background collector. The admin endpoint uses
// Sync to allow an explicit refresh regardless of the configured interval.
func (s *TaxRateSourceSnapshotService) SyncIfDue(ctx context.Context) (*TaxRateSourceSnapshotSyncResult, bool, error) {
	config, err := s.GetSourceConfiguration()
	if err != nil {
		return nil, false, err
	}
	if !config.Enabled {
		return nil, false, nil
	}
	now := time.Now().UTC()
	if config.LastSuccessfulSyncAt != nil && now.Sub(*config.LastSuccessfulSyncAt) < time.Duration(config.RefreshIntervalHours)*time.Hour {
		return nil, false, nil
	}
	if config.LastCheckedAt != nil && now.Sub(*config.LastCheckedAt) < taxRateSourceScheduledRetryWindow {
		return nil, false, nil
	}
	result, err := s.Sync(ctx)
	if errors.Is(err, ErrTaxRateSourceSyncInProgress) || errors.Is(err, ErrTaxRateSourceDisabled) {
		return nil, false, nil
	}
	return result, err == nil, err
}

func taxRateSourceSnapshotContentSHA256(entries []taxrate.TaxRateSourceSnapshotEntry) (string, error) {
	type contentEntry struct {
		CountryCode       string `json:"country_code"`
		SourceCountryCode string `json:"source_country_code"`
		CountryName       string `json:"country_name"`
		Currency          string `json:"currency"`
		RateType          string `json:"rate_type"`
		RateCategory      string `json:"rate_category"`
		RateDecimal       string `json:"rate_decimal"`
	}
	content := make([]contentEntry, 0, len(entries))
	for _, entry := range entries {
		content = append(content, contentEntry{
			CountryCode:       entry.CountryCode,
			SourceCountryCode: entry.SourceCountryCode,
			CountryName:       entry.CountryName,
			Currency:          entry.Currency,
			RateType:          entry.RateType,
			RateCategory:      entry.RateCategory,
			RateDecimal:       entry.RateDecimal,
		})
	}
	encoded, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("marshal normalized tax rate snapshot: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}
