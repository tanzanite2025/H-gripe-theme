package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/domain/seo"
)

var ErrInvalidSEOSettings = errors.New("invalid SEO settings")

const (
	seoMetaTitleLimit       = 160
	seoMetaDescriptionLimit = 320
)

type SEOService struct {
	settings                       *SettingService
	storefrontHTMLCacheInvalidator SEOStorefrontCacheInvalidator
}

type SEOStorefrontCacheInvalidator interface {
	PurgeAllAsync(reason string)
}

func NewSEOService(settings *SettingService) *SEOService {
	return &SEOService{settings: settings}
}

func (s *SEOService) SetStorefrontHTMLCacheInvalidator(invalidator SEOStorefrontCacheInvalidator) {
	if s == nil {
		return
	}
	s.storefrontHTMLCacheInvalidator = invalidator
}

func (s *SEOService) getHome(locale string) (*seo.Settings, error) {
	if s == nil || s.settings == nil {
		return nil, errors.New("SEO service is not configured")
	}

	values, err := loadPublicManagedSettingValues(s.settings, seo.Group, locale)
	if err != nil {
		return nil, err
	}

	result := normalizeSEOValues(values)
	return &result, nil
}

func (s *SEOService) updateHome(request seo.UpdateRequest) (*seo.Settings, error) {
	if s == nil || s.settings == nil {
		return nil, errors.New("SEO service is not configured")
	}

	locale := normalizeManagedSettingLocale(request.Locale)
	current, err := s.getHome(locale)
	if err != nil {
		return nil, err
	}

	next := *current
	if request.MetaTitle != nil {
		next.MetaTitle = *request.MetaTitle
	}
	if request.MetaDescription != nil {
		next.MetaDescription = *request.MetaDescription
	}

	normalized, err := normalizeSEOSettings(next)
	if err != nil {
		return nil, err
	}

	values := map[string]string{
		seo.HomeKeys.MetaTitle:       normalized.MetaTitle,
		seo.HomeKeys.MetaDescription: normalized.MetaDescription,
	}
	descriptions := map[string]string{
		seo.HomeKeys.MetaTitle:       "Home meta title",
		seo.HomeKeys.MetaDescription: "Home meta description",
	}
	if err := s.settings.BatchSet(managedSettingRecords(seo.Group, locale, values, descriptions)); err != nil {
		return nil, err
	}
	if s.storefrontHTMLCacheInvalidator != nil {
		s.storefrontHTMLCacheInvalidator.PurgeAllAsync("SEO settings updated")
	}

	return &normalized, nil
}

func normalizeSEOValues(values map[string]string) seo.Settings {
	return seo.Settings{
		MetaTitle:       values[seo.HomeKeys.MetaTitle],
		MetaDescription: values[seo.HomeKeys.MetaDescription],
	}
}

func normalizeSEOSettings(values seo.Settings) (seo.Settings, error) {
	values.MetaTitle = managedSettingValue(values.MetaTitle)
	values.MetaDescription = managedSettingValue(values.MetaDescription)

	if len([]rune(values.MetaTitle)) > seoMetaTitleLimit {
		return seo.Settings{}, fmt.Errorf("%w: meta_title exceeds %d characters", ErrInvalidSEOSettings, seoMetaTitleLimit)
	}
	if len([]rune(values.MetaDescription)) > seoMetaDescriptionLimit {
		return seo.Settings{}, fmt.Errorf("%w: meta_description exceeds %d characters", ErrInvalidSEOSettings, seoMetaDescriptionLimit)
	}
	return values, nil
}
