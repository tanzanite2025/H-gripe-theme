package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
)

const MediaUsageSummaryCandidateLimit = 200
const mediaUsageSummaryReferenceLimit = 20000
const mediaUsageSummaryVisibleReferencePreviewLimit = 20

var ErrMediaUsageSummaryAlreadyRunning = errors.New("media usage summary is already running")

// AllowedReferenceTypes is supplied by the authenticated HTTP boundary, never
// decoded from request data. Unknown business sources remain restricted.
type MediaUsageSummaryAccessScope struct {
	AllowedReferenceTypes map[string]bool
}

type MediaUsageSummaryReference struct {
	Category         media.ReferenceCategory `json:"category"`
	ResourceType     string                  `json:"resource_type"`
	ResourceID       string                  `json:"resource_id"`
	ParentResourceID string                  `json:"parent_resource_id"`
	Label            string                  `json:"label"`
	Field            string                  `json:"field"`
	UsageState       string                  `json:"usage_state"`
}

type MediaUsageSummaryRow struct {
	AssetID                 string                           `json:"asset_id"`
	Title                   string                           `json:"title"`
	LibraryPartition        media.MediaAssetLibraryPartition `json:"library_partition"`
	NoRecognizedReferences  bool                             `json:"no_recognized_references"`
	HasRestrictedReferences bool                             `json:"has_restricted_references"`
	PendingBusinessDraft    bool                             `json:"pending_business_draft"`
	VisibleReferenceTotal   int                              `json:"visible_reference_total"`
	CrossBusiness           bool                             `json:"cross_business"`
	Retained                bool                             `json:"retained"`
	Anomalous               bool                             `json:"anomalous"`
	Categories              []media.ReferenceCategory        `json:"categories"`
	References              []MediaUsageSummaryReference     `json:"references"`
	ReferencePreviewLimited bool                             `json:"reference_preview_limited"`
}

type MediaUsageSummaryReport struct {
	Coverage          string                 `json:"coverage"`
	SnapshotStartedAt time.Time              `json:"snapshot_started_at"`
	CompletedAt       time.Time              `json:"completed_at"`
	CandidateLimit    int                    `json:"candidate_limit"`
	Counts            map[string]int         `json:"counts"`
	Rows              []MediaUsageSummaryRow `json:"rows"`
}

func (s *MediaService) RebuildReadOnlyMediaUsageSummary(ctx context.Context, input MediaAssetListInput, scope MediaUsageSummaryAccessScope) (MediaUsageSummaryReport, error) {
	filter, err := normalizeMediaAssetListBusinessPartitionAndAccessScope(input)
	if err != nil {
		return MediaUsageSummaryReport{}, err
	}
	if s.usageSummaryRunning == nil || !s.usageSummaryRunning.CompareAndSwap(false, true) {
		return MediaUsageSummaryReport{}, ErrMediaUsageSummaryAlreadyRunning
	}
	defer s.usageSummaryRunning.Store(false)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	report := MediaUsageSummaryReport{Coverage: "recognized_business_sources_only", SnapshotStartedAt: time.Now().UTC(),
		CandidateLimit: MediaUsageSummaryCandidateLimit, Rows: make([]MediaUsageSummaryRow, 0),
		Counts: map[string]int{"all": 0, "no_recognized_references": 0, "cross_business": 0, "retained": 0, "anomalous": 0, "restricted": 0}}
	err = s.repo.WalkReadOnlyMediaUsageSnapshot(ctx, filter, MediaUsageSummaryCandidateLimit, mediaUsageSummaryReferenceLimit, func(asset media.MediaAsset, references []media.AssetReference) {
		row := summarizeMediaUsageWithPermissionScopedReferenceStates(asset, references, scope)
		report.Rows = append(report.Rows, row)
		report.Counts["all"]++
		for key, matches := range map[string]bool{"no_recognized_references": row.NoRecognizedReferences, "cross_business": row.CrossBusiness,
			"retained": row.Retained, "anomalous": row.Anomalous, "restricted": row.HasRestrictedReferences} {
			if matches {
				report.Counts[key]++
			}
		}
	})
	if err != nil {
		// A partially built report is never delivered, including on commit failure.
		return MediaUsageSummaryReport{}, err
	}
	report.CompletedAt = time.Now().UTC()
	return report, nil
}

func summarizeMediaUsageWithPermissionScopedReferenceStates(asset media.MediaAsset, references []media.AssetReference, scope MediaUsageSummaryAccessScope) MediaUsageSummaryRow {
	title := strings.TrimSpace(asset.OriginalFilename)
	if title == "" {
		title = strings.TrimSpace(asset.Filename)
	}
	if title == "" {
		title = "媒体 #" + strconv.FormatUint(uint64(asset.ID), 10)
	}
	row := MediaUsageSummaryRow{AssetID: strconv.FormatUint(uint64(asset.ID), 10), Title: title, LibraryPartition: asset.LibraryPartition,
		NoRecognizedReferences: len(references) == 0 && asset.BusinessDraftUploadExpiresAt == nil,
		PendingBusinessDraft:   asset.BusinessDraftUploadExpiresAt != nil, Retained: asset.BusinessDraftUploadExpiresAt != nil,
		Categories: make([]media.ReferenceCategory, 0), References: make([]MediaUsageSummaryReference, 0)}
	categories := make(map[media.ReferenceCategory]bool)
	for _, reference := range references {
		if !scope.AllowedReferenceTypes[reference.ResourceType] {
			row.HasRestrictedReferences = true
			continue
		}
		row.VisibleReferenceTotal++
		if reference.Category != media.ReferenceCategoryRegistration && !categories[reference.Category] {
			categories[reference.Category] = true
			row.Categories = append(row.Categories, reference.Category)
		}
		row.Anomalous = row.Anomalous || reference.UsageState == "missing_parent" || reference.Category == media.ReferenceCategoryRegistration
		state := reference.UsageState
		if state == "" && (reference.ResourceType == "media_upload_session" || reference.Field == "pending_upload") {
			state = "pending_upload"
		}
		switch state {
		case "retained_deleted_parent", "retained_deleted_placement", "retained_unpublished_parent", "retained_upload_receipt", "disabled", "draft", "archived", "pending_upload", "missing_parent":
			row.Retained = true
		}
		if isImmutableOrderHistoryMediaUsageReference(reference.ResourceType) {
			row.Retained = true
			state = "immutable_history"
		}
		if len(row.References) < mediaUsageSummaryVisibleReferencePreviewLimit {
			preview := MediaUsageSummaryReference{Category: reference.Category, ResourceType: reference.ResourceType,
				ResourceID: strconv.FormatUint(uint64(reference.ResourceID), 10), ParentResourceID: strconv.FormatUint(uint64(reference.ParentResourceID), 10),
				Label: reference.Label, Field: reference.Field, UsageState: state}
			if reference.Category == media.ReferenceCategoryRegistration {
				// The conflicting registration may belong to a private conversation
				// outside this actor's scope. Report the conflict without its identity.
				preview.ResourceID, preview.ParentResourceID = "0", "0"
				preview.Label = "其他媒体登记占用同一存储对象，需核查登记"
			}
			row.References = append(row.References, preview)
		}
	}
	row.CrossBusiness = len(categories) > 1
	row.ReferencePreviewLimited = row.VisibleReferenceTotal > len(row.References)
	return row
}

func isImmutableOrderHistoryMediaUsageReference(resourceType string) bool {
	switch resourceType {
	case "order_evidence_attachments", "order_evidence_items", "order_evidence_snapshots", "order_evidence_submission_snapshots",
		"order_evidence_export_snapshots", "order_policy_disclosures", "paypal_disputes":
		return true
	default:
		return false
	}
}
