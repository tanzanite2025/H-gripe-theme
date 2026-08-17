package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
)

const (
	MediaImageDimensionStateAll               = "all"
	MediaImageDimensionStateAttention         = "attention"
	MediaImageDimensionStateMissingDimensions = "missing_dimensions"
	MediaImageDimensionStateMissingVariants   = "missing_variants"
	MediaImageDimensionStateReady             = "ready"
)

type MediaImageDimensionListInput struct {
	Page     int
	PageSize int
	Search   string
	State    string
}

type MediaImageDimensionFinding struct {
	Asset          mediadomain.MediaAsset `json:"asset"`
	State          string                 `json:"state"`
	MissingPresets []string               `json:"missing_presets,omitempty"`
}

type MediaImageDimensionSummary struct {
	Total             int64 `json:"total"`
	Attention         int64 `json:"attention"`
	Ready             int64 `json:"ready"`
	MissingDimensions int64 `json:"missing_dimensions"`
	MissingVariants   int64 `json:"missing_variants"`
}

type MediaImageDimensionListResult struct {
	Items   []MediaImageDimensionFinding      `json:"items"`
	Summary MediaImageDimensionSummary        `json:"summary"`
	Presets []MediaDerivativePresetDefinition `json:"presets"`
	Total   int64                             `json:"-"`
}

type MediaImageDimensionEngine struct {
	media *MediaService
}

func NewMediaImageDimensionEngine(media *MediaService) *MediaImageDimensionEngine {
	return &MediaImageDimensionEngine{media: media}
}

func (e *MediaImageDimensionEngine) List(
	input MediaImageDimensionListInput,
) (*MediaImageDimensionListResult, error) {
	if e == nil || e.media == nil {
		return nil, errors.New("media image dimension engine is unavailable")
	}
	return e.media.ListImageDimensionFindings(input)
}

func (e *MediaImageDimensionEngine) Reconcile(
	ctx context.Context,
	assetID uint,
) (*MediaImageDimensionFinding, error) {
	if e == nil || e.media == nil {
		return nil, ErrMediaStorageUnavailable
	}
	return e.media.ReconcileImageDimensions(ctx, assetID)
}

// ListImageDimensionFindings is the media-domain image dimensions engine. It
// intentionally evaluates stored assets only, without joining page-level
// Lighthouse findings into this separate operational workflow.
func (s *MediaService) ListImageDimensionFindings(
	input MediaImageDimensionListInput,
) (*MediaImageDimensionListResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("media image dimension engine is unavailable")
	}

	state, err := normalizeMediaImageDimensionState(input.State)
	if err != nil {
		return nil, err
	}
	presets, err := s.activeMediaDerivativePresetDefinitions()
	if err != nil {
		return nil, err
	}
	requiredDerivatives := mediaDerivativePresetRequirements(presets)
	assets, total, err := s.repo.ListImageDimensionAssets(repository.MediaImageDimensionFilter{
		Page:                input.Page,
		PageSize:            input.PageSize,
		Search:              strings.TrimSpace(input.Search),
		State:               state,
		RequiredDerivatives: requiredDerivatives,
	})
	if err != nil {
		return nil, err
	}

	items := make([]MediaImageDimensionFinding, 0, len(assets))
	for index := range assets {
		s.hydrateAssetAccessURL(&assets[index])
		items = append(items, mediaImageDimensionFinding(assets[index], presets))
	}

	summary, err := s.imageDimensionSummary(requiredDerivatives)
	if err != nil {
		return nil, err
	}
	return &MediaImageDimensionListResult{
		Items:   items,
		Summary: summary,
		Presets: presets,
		Total:   total,
	}, nil
}

// ReconcileImageDimensions reads the original stored image, records its real
// dimensions, and creates any missing standard derivative sizes in one action.
func (s *MediaService) ReconcileImageDimensions(
	ctx context.Context,
	assetID uint,
) (*MediaImageDimensionFinding, error) {
	if s == nil || s.repo == nil || s.storage == nil {
		return nil, ErrMediaStorageUnavailable
	}
	asset, err := s.GetAsset(assetID)
	if err != nil {
		return nil, err
	}
	if asset.MediaType != "image" {
		return nil, ErrUnsupportedMediaType
	}

	if err := s.syncImageAssetDimensions(ctx, asset); err != nil {
		return nil, err
	}

	if _, err := s.ensureAssetDerivatives(ctx, asset); err != nil {
		return nil, err
	}
	presets, err := s.activeMediaDerivativePresetDefinitions()
	if err != nil {
		return nil, err
	}
	s.hydrateAssetAccessURL(asset)
	finding := mediaImageDimensionFinding(*asset, presets)
	return &finding, nil
}

func (s *MediaService) syncImageAssetDimensions(ctx context.Context, asset *mediadomain.MediaAsset) error {
	if s == nil || s.repo == nil || asset == nil {
		return ErrMediaStorageUnavailable
	}
	openSource, err := s.openMediaDerivativeSource(ctx, asset)
	if err != nil {
		return err
	}
	release, err := acquireMediaDerivativeGenerationSlot(ctx)
	if err != nil {
		return err
	}
	decoded, decodeErr := decodeDerivativeSource(openSource)
	release()
	if decodeErr != nil {
		return decodeErr
	}
	bounds := decoded.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return fmt.Errorf("%w: source image dimensions are invalid", ErrMediaDerivativeGenerationFailed)
	}
	if asset.Width != width || asset.Height != height {
		if err := s.repo.UpdateAssetDimensions(asset.ID, width, height); err != nil {
			return err
		}
		asset.Width = width
		asset.Height = height
	}
	return nil
}

func (s *MediaService) imageDimensionSummary(requiredDerivatives []repository.MediaDerivativeRequirement) (MediaImageDimensionSummary, error) {
	total, err := s.repo.CountImageDimensionAssets(MediaImageDimensionStateAll, requiredDerivatives)
	if err != nil {
		return MediaImageDimensionSummary{}, err
	}
	ready, err := s.repo.CountImageDimensionAssets(MediaImageDimensionStateReady, requiredDerivatives)
	if err != nil {
		return MediaImageDimensionSummary{}, err
	}
	missingDimensions, err := s.repo.CountImageDimensionAssets(MediaImageDimensionStateMissingDimensions, requiredDerivatives)
	if err != nil {
		return MediaImageDimensionSummary{}, err
	}
	missingVariants, err := s.repo.CountImageDimensionAssets(MediaImageDimensionStateMissingVariants, requiredDerivatives)
	if err != nil {
		return MediaImageDimensionSummary{}, err
	}
	return MediaImageDimensionSummary{
		Total:             total,
		Attention:         total - ready,
		Ready:             ready,
		MissingDimensions: missingDimensions,
		MissingVariants:   missingVariants,
	}, nil
}

func mediaImageDimensionFinding(asset mediadomain.MediaAsset, presets []MediaDerivativePresetDefinition) MediaImageDimensionFinding {
	missing := missingMediaDerivativePresets(asset.Derivatives, presets)
	switch {
	case asset.Width <= 0 || asset.Height <= 0:
		if len(missing) > 0 {
			return MediaImageDimensionFinding{
				Asset:          asset,
				State:          "missing_dimensions_and_variants",
				MissingPresets: mediaDerivativePresetNamesFor(missing),
			}
		}
		return MediaImageDimensionFinding{
			Asset: asset,
			State: MediaImageDimensionStateMissingDimensions,
		}
	case len(missing) > 0:
		return MediaImageDimensionFinding{
			Asset:          asset,
			State:          MediaImageDimensionStateMissingVariants,
			MissingPresets: mediaDerivativePresetNamesFor(missing),
		}
	default:
		return MediaImageDimensionFinding{
			Asset: asset,
			State: MediaImageDimensionStateReady,
		}
	}
}

func mediaDerivativePresetNamesFor(presets []MediaDerivativePresetDefinition) []string {
	names := make([]string, 0, len(presets))
	for _, preset := range presets {
		names = append(names, preset.Name)
	}
	return names
}

func normalizeMediaImageDimensionState(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "", MediaImageDimensionStateAll:
		return MediaImageDimensionStateAll, nil
	case MediaImageDimensionStateAttention,
		MediaImageDimensionStateMissingDimensions,
		MediaImageDimensionStateMissingVariants,
		MediaImageDimensionStateReady:
		return strings.TrimSpace(value), nil
	default:
		return "", errors.New("unsupported image dimension state")
	}
}
