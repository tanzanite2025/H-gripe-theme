package service

import (
	"errors"
	"fmt"
	"net/url"
	stdpath "path"
	"strings"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

type StorefrontRedirectRuleService struct {
	rules   *repository.StorefrontRedirectRuleRepository
	catalog *repository.StorefrontRouteCatalogRepository
}

func NewStorefrontRedirectRuleService(
	rules *repository.StorefrontRedirectRuleRepository,
	catalog *repository.StorefrontRouteCatalogRepository,
) *StorefrontRedirectRuleService {
	return &StorefrontRedirectRuleService{
		rules:   rules,
		catalog: catalog,
	}
}

func (s *StorefrontRedirectRuleService) List() ([]urlmanagementdomain.StorefrontRedirectRule, error) {
	if s == nil || s.rules == nil {
		return nil, errors.New("storefront redirect rule service is unavailable")
	}
	return s.rules.List()
}

func (s *StorefrontRedirectRuleService) ListPublished() ([]urlmanagementdomain.StorefrontPublishedRedirect, error) {
	if s == nil || s.rules == nil {
		return nil, errors.New("storefront redirect rule service is unavailable")
	}
	return s.rules.ListPublished()
}

func (s *StorefrontRedirectRuleService) Create(
	input urlmanagementdomain.StorefrontRedirectRuleInput,
	createdByID uint,
) (*urlmanagementdomain.StorefrontRedirectRule, error) {
	if s == nil || s.rules == nil {
		return nil, errors.New("storefront redirect rule service is unavailable")
	}

	sourcePath, targetPath, statusCode, reason, err := s.normalizeInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.validatePaths(sourcePath, targetPath); err != nil {
		return nil, err
	}

	if existing, err := s.rules.FindBySourcePath(sourcePath); err == nil && existing != nil {
		return nil, fmt.Errorf("a redirect rule already exists for %s", sourcePath)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	rule := &urlmanagementdomain.StorefrontRedirectRule{
		SourcePath:  sourcePath,
		TargetPath:  targetPath,
		StatusCode:  statusCode,
		State:       urlmanagementdomain.RedirectRuleStateDraft,
		Reason:      reason,
		CreatedByID: createdByID,
	}
	if err := s.rules.Create(rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *StorefrontRedirectRuleService) Publish(id, publishedByID uint) (*urlmanagementdomain.StorefrontRedirectRule, error) {
	if s == nil || s.rules == nil {
		return nil, errors.New("storefront redirect rule service is unavailable")
	}
	rule, err := s.rules.FindByID(id)
	if err != nil {
		return nil, err
	}
	if rule.State == urlmanagementdomain.RedirectRuleStateDisabled {
		return nil, errors.New("disabled redirect rules cannot be published")
	}
	if err := s.validatePaths(rule.SourcePath, rule.TargetPath); err != nil {
		return nil, err
	}
	if err := s.rules.Publish(id, publishedByID, time.Now().UTC()); err != nil {
		return nil, err
	}
	return s.rules.FindByID(id)
}

func (s *StorefrontRedirectRuleService) Disable(id uint) (*urlmanagementdomain.StorefrontRedirectRule, error) {
	if s == nil || s.rules == nil {
		return nil, errors.New("storefront redirect rule service is unavailable")
	}
	if _, err := s.rules.FindByID(id); err != nil {
		return nil, err
	}
	if err := s.rules.Disable(id, time.Now().UTC()); err != nil {
		return nil, err
	}
	return s.rules.FindByID(id)
}

func (s *StorefrontRedirectRuleService) normalizeInput(
	input urlmanagementdomain.StorefrontRedirectRuleInput,
) (string, string, int, string, error) {
	sourcePath, err := normalizeStorefrontRedirectPath(input.SourcePath)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("source path: %w", err)
	}
	targetPath, err := normalizeStorefrontRedirectPath(input.TargetPath)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("target path: %w", err)
	}
	if sourcePath == targetPath {
		return "", "", 0, "", errors.New("source and target paths must differ")
	}
	if input.StatusCode != 301 && input.StatusCode != 308 {
		return "", "", 0, "", errors.New("redirect status code must be 301 or 308")
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return "", "", 0, "", errors.New("redirect reason is required")
	}
	return sourcePath, targetPath, input.StatusCode, reason, nil
}

func (s *StorefrontRedirectRuleService) validatePaths(sourcePath, targetPath string) error {
	if s == nil || s.catalog == nil {
		return errors.New("storefront route catalog is unavailable")
	}

	source, err := s.catalog.FindCurrentByPath(sourcePath)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if source != nil {
		if source.IsAlias || source.EntryStatus == seodomain.RouteEntryStatusAlias {
			return fmt.Errorf("%s is a system-owned alias and must be changed in the storefront route registry", sourcePath)
		}
		return fmt.Errorf("%s is still an active storefront route", sourcePath)
	}

	target, err := s.catalog.FindCanonicalByPath(targetPath)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%s is not an active canonical storefront route", targetPath)
		}
		return err
	}
	if target == nil {
		return fmt.Errorf("%s is not an active canonical storefront route", targetPath)
	}
	return nil
}

func normalizeStorefrontRedirectPath(value string) (string, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "", errors.New("path is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil {
		return "", errors.New("path is invalid")
	}
	if parsed.IsAbs() || parsed.Host != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("use an exact internal path without host, query, or fragment")
	}
	if !strings.HasPrefix(parsed.Path, "/") {
		return "", errors.New("path must begin with /")
	}
	cleaned := stdpath.Clean(parsed.Path)
	if cleaned == "." || cleaned == "/" {
		return "", errors.New("root path cannot be redirected")
	}
	return strings.TrimRight(cleaned, "/"), nil
}
