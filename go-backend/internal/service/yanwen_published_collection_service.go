package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

// YanwenPublishedCollectionReference is the small identity projection that
// generic shipping management may consume when selecting a service route.
type YanwenPublishedCollectionReference struct {
	ID          uint   `json:"id"`
	ProductCode string `json:"product_code"`
	DisplayName string `json:"display_name"`
	Countries   string `json:"countries"`
	Enabled     bool   `json:"enabled"`
}

// YanwenPublishedCollectionService owns the operator-curated Yanwen service
// collection and its relationship to the official product catalog.
type YanwenPublishedCollectionService struct {
	channels *repository.YanwenPublishedChannelRepository
	products *repository.YanwenProductCatalogRepository
}

func NewYanwenPublishedCollectionService(
	channels *repository.YanwenPublishedChannelRepository,
	products *repository.YanwenProductCatalogRepository,
) *YanwenPublishedCollectionService {
	return &YanwenPublishedCollectionService{channels: channels, products: products}
}

func (s *YanwenPublishedCollectionService) ListYanwenPublishedChannels(
	environment string,
	enabledOnly bool,
) ([]shipping.YanwenPublishedChannel, error) {
	if s == nil || s.channels == nil {
		return nil, errors.New("Yanwen published channel repository is not configured")
	}
	normalizedEnvironment, err := shipping.NormalizeYanwenPublishedChannelEnvironment(environment)
	if err != nil {
		return nil, err
	}
	return s.channels.FindYanwenPublishedChannelsByEnvironment(normalizedEnvironment, enabledOnly)
}

func (s *YanwenPublishedCollectionService) ListProductionYanwenCollectionReferences() (
	[]YanwenPublishedCollectionReference,
	error,
) {
	return s.listProductionYanwenCollectionReferences(true)
}

// ListProductionYanwenCollectionReferencesIncludingDisabled returns the same
// narrow projection for template validation. Disabled records are included so
// the shipping domain can report a disabled collection explicitly without
// reading Yanwen channel entities or repository data directly.
func (s *YanwenPublishedCollectionService) ListProductionYanwenCollectionReferencesIncludingDisabled() (
	[]YanwenPublishedCollectionReference,
	error,
) {
	return s.listProductionYanwenCollectionReferences(false)
}

func (s *YanwenPublishedCollectionService) listProductionYanwenCollectionReferences(
	enabledOnly bool,
) ([]YanwenPublishedCollectionReference, error) {
	channels, err := s.ListYanwenPublishedChannels(
		shipping.YanwenPublishedChannelEnvironmentProduction,
		enabledOnly,
	)
	if err != nil {
		return nil, err
	}
	references := make([]YanwenPublishedCollectionReference, 0, len(channels))
	for _, channel := range channels {
		references = append(references, newYanwenPublishedCollectionReference(channel))
	}
	return references, nil
}

func (s *YanwenPublishedCollectionService) GetYanwenPublishedChannel(id uint) (*shipping.YanwenPublishedChannel, error) {
	if s == nil || s.channels == nil {
		return nil, errors.New("Yanwen published channel repository is not configured")
	}
	return s.channels.FindYanwenPublishedChannelByID(id)
}

func (s *YanwenPublishedCollectionService) CreateYanwenPublishedChannel(channel *shipping.YanwenPublishedChannel) error {
	if channel == nil {
		return errors.New("Yanwen channel is required")
	}
	if err := channel.Validate(); err != nil {
		return err
	}
	if err := s.validateOfficialProduct(channel); err != nil {
		return err
	}
	if s == nil || s.channels == nil {
		return errors.New("Yanwen published channel repository is not configured")
	}
	return s.channels.CreateYanwenPublishedChannel(channel)
}

func (s *YanwenPublishedCollectionService) UpdateYanwenPublishedChannel(channel *shipping.YanwenPublishedChannel) error {
	if channel == nil {
		return errors.New("Yanwen channel is required")
	}
	if err := channel.Validate(); err != nil {
		return err
	}
	if err := s.validateOfficialProduct(channel); err != nil {
		return err
	}
	if s == nil || s.channels == nil {
		return errors.New("Yanwen published channel repository is not configured")
	}
	return s.channels.UpdateYanwenPublishedChannel(channel)
}

func (s *YanwenPublishedCollectionService) DeleteYanwenPublishedChannel(id uint) error {
	if s == nil || s.channels == nil {
		return errors.New("Yanwen published channel repository is not configured")
	}
	return s.channels.DeleteYanwenPublishedChannel(id)
}

func (s *YanwenPublishedCollectionService) validateOfficialProduct(
	channel *shipping.YanwenPublishedChannel,
) error {
	if s == nil || s.products == nil {
		return errors.New("Yanwen product catalog repository is not configured")
	}
	products, err := s.products.FindYanwenProductCatalogEntriesByEnvironment(channel.Environment)
	if err != nil {
		return fmt.Errorf("load Yanwen %s product catalog: %w", channel.Environment, err)
	}
	for _, product := range products {
		if strings.EqualFold(strings.TrimSpace(product.ProductID), channel.ProductCode) {
			return nil
		}
	}
	return fmt.Errorf(
		"Yanwen product %q is not present in the official %s catalog",
		channel.ProductCode,
		channel.Environment,
	)
}

func newYanwenPublishedCollectionReference(
	channel shipping.YanwenPublishedChannel,
) YanwenPublishedCollectionReference {
	return YanwenPublishedCollectionReference{
		ID:          channel.ID,
		ProductCode: strings.TrimSpace(channel.ProductCode),
		DisplayName: strings.TrimSpace(channel.DisplayName),
		Countries:   channel.Countries,
		Enabled:     channel.Enabled,
	}
}
