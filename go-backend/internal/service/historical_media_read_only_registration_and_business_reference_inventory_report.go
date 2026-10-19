package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type HistoricalMediaReadOnlyInventoryOptions struct {
	BatchSize           int `json:"batch_size"`
	MaxRecordsPerSource int `json:"max_records_per_source"`
	MaxIssues           int `json:"max_issues"`
}

type HistoricalMediaReadOnlyInventoryIssue struct {
	Code                  string `json:"code"`
	Table                 string `json:"table"`
	RecordID              uint   `json:"record_id"`
	Field                 string `json:"field"`
	AssetID               uint   `json:"asset_id,omitempty"`
	RelatedRecordID       uint   `json:"related_record_id,omitempty"`
	ReferenceSHA256       string `json:"reference_sha256,omitempty"`
	Origin                string `json:"origin,omitempty"`
	RetainedDeletedRecord bool   `json:"retained_deleted_record,omitempty"`
}

type HistoricalMediaReadOnlyInventorySourceCoverage struct {
	Table         string `json:"table"`
	Field         string `json:"field"`
	RowsScanned   int    `json:"rows_scanned"`
	LastRecordID  uint   `json:"last_record_id"`
	Complete      bool   `json:"complete"`
	Unavailable   bool   `json:"unavailable,omitempty"`
	RemainingRows int64  `json:"remaining_rows,omitempty"`
}

type HistoricalMediaReadOnlyInventoryReport struct {
	GeneratedAt                  time.Time                                        `json:"generated_at"`
	DatabaseTarget               string                                           `json:"database_target,omitempty"`
	IdentityConfiguration        map[string]string                                `json:"identity_configuration"`
	Options                      HistoricalMediaReadOnlyInventoryOptions          `json:"options"`
	SnapshotMode                 string                                           `json:"snapshot_mode"`
	Coverage                     string                                           `json:"coverage"`
	PhysicalObjectsChecked       bool                                             `json:"physical_objects_checked"`
	UnreferencedObjectsEvaluated bool                                             `json:"unreferenced_objects_evaluated"`
	ExcludedSources              []string                                         `json:"excluded_sources"`
	RegistrationsScanned         int                                              `json:"registrations_scanned"`
	ReferencesScanned            int                                              `json:"references_scanned"`
	RegistrationIndexComplete    bool                                             `json:"registration_index_complete"`
	RequestedSourcesComplete     bool                                             `json:"requested_sources_complete"`
	Sources                      []HistoricalMediaReadOnlyInventorySourceCoverage `json:"sources"`
	TotalIssues                  int                                              `json:"total_issues"`
	IssueCounts                  map[string]int                                   `json:"issue_counts"`
	IssuesTruncated              bool                                             `json:"issues_truncated"`
	Issues                       []HistoricalMediaReadOnlyInventoryIssue          `json:"issues"`
}

type HistoricalMediaReadOnlyInventoryService struct {
	db                  *gorm.DB
	sharedStorageConfig *storage.Config
	logoStorageConfig   *storage.Config
	storefrontURL       string
}

func NewHistoricalMediaReadOnlyInventoryService(db *gorm.DB, shared, logo *storage.Config, storefrontURL string) *HistoricalMediaReadOnlyInventoryService {
	return &HistoricalMediaReadOnlyInventoryService{db: db, sharedStorageConfig: shared, logoStorageConfig: logo, storefrontURL: storefrontURL}
}

type historicalMediaReadOnlyRegistrationIdentity struct {
	Table      string
	ID         uint
	AssetID    uint
	URL        string
	StorageKey string
	DeletedAt  gorm.DeletedAt
	Asset      *media.MediaAsset
}

type historicalMediaReadOnlyRegistrationIndex struct {
	entries  []historicalMediaReadOnlyRegistrationIdentity
	urls     map[string][]int
	keys     map[string][]int
	assets   map[uint]*media.MediaAsset
	complete bool
}

func (s *HistoricalMediaReadOnlyInventoryService) AuditHistoricalRegistrationsAndBusinessReferences(ctx context.Context, options HistoricalMediaReadOnlyInventoryOptions) (HistoricalMediaReadOnlyInventoryReport, error) {
	report := HistoricalMediaReadOnlyInventoryReport{}
	if s == nil || s.db == nil || s.sharedStorageConfig == nil || s.logoStorageConfig == nil {
		return report, fmt.Errorf("historical media audit requires database and both storage identity configurations")
	}
	if options.BatchSize < 1 || options.BatchSize > 1000 || options.MaxRecordsPerSource < 1 || options.MaxRecordsPerSource > 1000000 || options.MaxIssues < 1 || options.MaxIssues > 100000 {
		return report, fmt.Errorf("invalid historical media audit limits")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	report = HistoricalMediaReadOnlyInventoryReport{
		GeneratedAt: time.Now().UTC(), Options: options, SnapshotMode: "repeatable_read_read_only",
		IdentityConfiguration: map[string]string{
			"shared_storage_type":    string(s.sharedStorageConfig.Type),
			"shared_public_base_url": historicalMediaInventoryPublicConfigurationURLWithoutCredentialsOrTokens(s.sharedStorageConfig.BaseURL),
			"logo_storage_type":      string(s.logoStorageConfig.Type),
			"logo_public_base_url":   historicalMediaInventoryPublicConfigurationURLWithoutCredentialsOrTokens(s.logoStorageConfig.BaseURL),
			"storefront_base_url":    historicalMediaInventoryPublicConfigurationURLWithoutCredentialsOrTokens(s.storefrontURL),
		},
		Coverage:                  "registered_originals_and_derivatives_and_recognized_public_business_fields_including_retained_rows",
		ExcludedSources:           []string{"community_and_private_attachment_payloads", "unrecognized_custom_fields_and_retired_unconfigured_origins", "physical_storage_and_unregistered_objects", "cleanup_queue_and_upload_receipts"},
		RegistrationIndexComplete: true, RequestedSourcesComplete: true, IssueCounts: make(map[string]int), Sources: make([]HistoricalMediaReadOnlyInventorySourceCoverage, 0), Issues: make([]HistoricalMediaReadOnlyInventoryIssue, 0),
	}
	if s.db.Dialector.Name() == "sqlite" {
		report.SnapshotMode = "sqlite_test_transaction"
	}
	// PostgreSQL/MySQL enforce read-only and one consistent snapshot. No migration,
	// business-row locks, cleanup events or storage client is involved.
	err := s.db.WithContext(ctx).Session(&gorm.Session{Logger: s.db.Logger.LogMode(logger.Silent)}).Transaction(func(tx *gorm.DB) error {
		index, err := s.loadHistoricalMediaRegistrationIdentityIndex(tx, &report)
		if err != nil {
			return err
		}
		if err := s.scanHistoricalMediaBusinessFieldsWithoutMutatingRecords(tx, index, &report); err != nil {
			return err
		}
		for _, source := range report.Sources {
			report.RequestedSourcesComplete = report.RequestedSourcesComplete && source.Complete && !source.Unavailable
		}
		return ctx.Err()
	}, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	return report, err
}

func historicalMediaInventoryPublicConfigurationURLWithoutCredentialsOrTokens(reference string) string {
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	parsed.User, parsed.RawQuery, parsed.Fragment, parsed.RawFragment, parsed.ForceQuery = nil, "", "", "", false
	return parsed.String()
}

func (report *HistoricalMediaReadOnlyInventoryReport) recordHistoricalMediaInventoryIssue(issue HistoricalMediaReadOnlyInventoryIssue, reference string) {
	if reference != "" {
		fingerprint := sha256.Sum256([]byte(reference))
		issue.ReferenceSHA256 = hex.EncodeToString(fingerprint[:])
		if parsed, err := url.Parse(reference); err == nil && parsed.Hostname() != "" && (parsed.Scheme == "http" || parsed.Scheme == "https") {
			issue.Origin = parsed.Scheme + "://" + parsed.Host
		}
	}
	report.TotalIssues++
	report.IssueCounts[issue.Code]++
	if len(report.Issues) < report.Options.MaxIssues {
		report.Issues = append(report.Issues, issue)
	} else {
		report.IssuesTruncated = true
	}
}

func normalizedHistoricalMediaRegistrationURL(reference string) string {
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil {
		return strings.TrimSpace(reference)
	}
	parsed.RawQuery, parsed.Fragment, parsed.RawFragment, parsed.ForceQuery = "", "", "", false
	return parsed.String()
}

func (s *HistoricalMediaReadOnlyInventoryService) resolveHistoricalMediaPublicReferenceObjectKey(reference string) (string, error) {
	key, err := storage.ResolveConfiguredPublicMediaObjectKeyWithoutStorageAccess(reference, s.sharedStorageConfig, s.storefrontURL)
	if err != nil || key != "" {
		return key, err
	}
	logoKey, err := storage.ResolveConfiguredPublicMediaObjectKeyWithoutStorageAccess(reference, s.logoStorageConfig, s.storefrontURL)
	if err == nil && storage.DedicatedPublicObjectNamespaceForStorageKey(logoKey) == "site-logo" {
		return logoKey, nil
	}
	return "", nil
}

func (s *HistoricalMediaReadOnlyInventoryService) loadHistoricalMediaRegistrationIdentityIndex(tx *gorm.DB, report *HistoricalMediaReadOnlyInventoryReport) (*historicalMediaReadOnlyRegistrationIndex, error) {
	index := &historicalMediaReadOnlyRegistrationIndex{urls: make(map[string][]int), keys: make(map[string][]int), assets: make(map[uint]*media.MediaAsset), complete: true}
	for _, table := range []string{"media_assets", "media_asset_derivatives"} {
		coverage := HistoricalMediaReadOnlyInventorySourceCoverage{Table: table, Field: "registration", Complete: true}
		if !tx.Migrator().HasTable(table) {
			coverage.Unavailable, coverage.Complete, index.complete = true, false, false
			report.recordHistoricalMediaInventoryIssue(HistoricalMediaReadOnlyInventoryIssue{Code: "registration_table_unavailable", Table: table, Field: "registration"}, "")
			report.Sources = append(report.Sources, coverage)
			continue
		}
		for {
			limit := min(report.Options.BatchSize, report.Options.MaxRecordsPerSource-coverage.RowsScanned)
			if limit == 0 {
				if err := tx.Table(table).Where("id > ?", coverage.LastRecordID).Count(&coverage.RemainingRows).Error; err != nil {
					return nil, err
				}
				coverage.Complete = coverage.RemainingRows == 0
				index.complete = index.complete && coverage.Complete
				break
			}
			var registrations []historicalMediaReadOnlyRegistrationIdentity
			if table == "media_assets" {
				var assets []media.MediaAsset
				if err := tx.Unscoped().Select("id", "url", "storage_key", "deleted_at", "media_type", "status", "visibility", "library_partition", "upload_origin", "uploader_actor_type", "classification_method", "customer_service_media_purpose").Where("id > ?", coverage.LastRecordID).Order("id ASC").Limit(limit).Find(&assets).Error; err != nil {
					return nil, fmt.Errorf("read historical media registrations: %w", err)
				}
				for position := range assets {
					asset := &assets[position]
					index.assets[asset.ID] = asset
					registrations = append(registrations, historicalMediaReadOnlyRegistrationIdentity{Table: table, ID: asset.ID, AssetID: asset.ID, URL: asset.URL, StorageKey: asset.StorageKey, DeletedAt: asset.DeletedAt, Asset: asset})
					if asset.LibraryPartition == "" || asset.LibraryPartition == media.MediaLibraryPartitionUnclassified || asset.UploadOrigin == "" || asset.UploadOrigin == media.MediaUploadOriginLegacyUnknown {
						report.recordHistoricalMediaInventoryIssue(HistoricalMediaReadOnlyInventoryIssue{Code: "historical_classification_or_provenance_unverified", Table: table, RecordID: asset.ID, Field: "library_partition,upload_origin", AssetID: asset.ID, RetainedDeletedRecord: asset.DeletedAt.Valid}, "")
					}
				}
			} else {
				var derivatives []media.MediaAssetDerivative
				if err := tx.Unscoped().Select("id", "media_asset_id", "url", "storage_key", "deleted_at").Where("id > ?", coverage.LastRecordID).Order("id ASC").Limit(limit).Find(&derivatives).Error; err != nil {
					return nil, err
				}
				for _, derivative := range derivatives {
					registrations = append(registrations, historicalMediaReadOnlyRegistrationIdentity{Table: table, ID: derivative.ID, AssetID: derivative.MediaAssetID, URL: derivative.URL, StorageKey: derivative.StorageKey, DeletedAt: derivative.DeletedAt})
					registrations[len(registrations)-1].Asset = index.assets[derivative.MediaAssetID]
					if index.assets[derivative.MediaAssetID] == nil && index.complete {
						report.recordHistoricalMediaInventoryIssue(HistoricalMediaReadOnlyInventoryIssue{Code: "derivative_parent_registration_missing", Table: table, RecordID: derivative.ID, Field: "media_asset_id", AssetID: derivative.MediaAssetID, RetainedDeletedRecord: derivative.DeletedAt.Valid}, "")
					}
				}
			}
			for _, registration := range registrations {
				s.addHistoricalMediaRegistrationIdentityAndDetectConflicts(index, report, registration)
				coverage.LastRecordID = registration.ID
			}
			coverage.RowsScanned += len(registrations)
			report.RegistrationsScanned += len(registrations)
			if len(registrations) < limit {
				break
			}
		}
		report.Sources = append(report.Sources, coverage)
	}
	report.RegistrationIndexComplete = index.complete
	return index, nil
}

func (s *HistoricalMediaReadOnlyInventoryService) addHistoricalMediaRegistrationIdentityAndDetectConflicts(index *historicalMediaReadOnlyRegistrationIndex, report *HistoricalMediaReadOnlyInventoryReport, registration historicalMediaReadOnlyRegistrationIdentity) {
	issue := HistoricalMediaReadOnlyInventoryIssue{Table: registration.Table, RecordID: registration.ID, Field: "storage_key,url", AssetID: registration.AssetID, RetainedDeletedRecord: registration.DeletedAt.Valid}
	key, valid := storage.NormalizeObjectKey(registration.StorageKey)
	if registration.StorageKey == "" {
		issue.Code = "registration_storage_key_missing"
		report.recordHistoricalMediaInventoryIssue(issue, registration.URL)
	} else if !valid {
		issue.Code = "registration_storage_key_invalid"
		report.recordHistoricalMediaInventoryIssue(issue, registration.StorageKey)
	}
	urlKey, err := s.resolveHistoricalMediaPublicReferenceObjectKey(registration.URL)
	if err == nil && urlKey != "" && key != "" && key != urlKey {
		issue.Code = "registration_url_and_storage_key_disagree"
		report.recordHistoricalMediaInventoryIssue(issue, registration.URL)
	}
	if err != nil && (registration.Asset == nil || registration.Asset.Visibility == "public") {
		issue.Code = "registration_public_url_invalid"
		report.recordHistoricalMediaInventoryIssue(issue, registration.URL)
	}
	if err == nil && urlKey == "" && strings.TrimSpace(registration.URL) != "" && (registration.Asset == nil || registration.Asset.Visibility == "public") {
		issue.Code = "registration_url_origin_not_verified_by_current_configuration"
		report.recordHistoricalMediaInventoryIssue(issue, registration.URL)
	}
	if key == "" && err == nil {
		key = urlKey
	}
	if storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" || storage.DedicatedPublicObjectNamespaceForStorageKey(urlKey) != "" {
		issue.Code = "shared_registration_in_dedicated_namespace_requires_storage_review"
		report.recordHistoricalMediaInventoryIssue(issue, registration.URL)
	}
	position := len(index.entries)
	index.entries = append(index.entries, registration)
	identities := map[int]bool{}
	address := normalizedHistoricalMediaRegistrationURL(registration.URL)
	for _, prior := range index.urls[address] {
		identities[prior] = true
	}
	if key != "" {
		for _, prior := range index.keys[key] {
			identities[prior] = true
		}
	}
	if len(identities) > 0 {
		issue.Code = "duplicate_registered_object_identity"
		firstPrior := position
		for prior := range identities {
			if prior < firstPrior {
				firstPrior = prior
			}
		}
		issue.RelatedRecordID = index.entries[firstPrior].ID
		report.recordHistoricalMediaInventoryIssue(issue, registration.URL)
	}
	if address != "" {
		index.urls[address] = append(index.urls[address], position)
	}
	if key != "" {
		index.keys[key] = append(index.keys[key], position)
	}
}
