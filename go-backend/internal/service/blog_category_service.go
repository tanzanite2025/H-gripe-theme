package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var blogCategorySlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)

var (
	ErrBlogCategoryNotFound   = errors.New("blog category not found")
	ErrBlogCategoryInvalid    = errors.New("blog category invalid")
	ErrBlogCategorySlugExists = errors.New("blog category slug already exists")
	ErrBlogCategoryInUse      = errors.New("blog category is still used by posts")
)

type BlogCategoryService struct {
	repo *repository.BlogCategoryRepository
}

type BlogCategoryInput struct {
	Name        string
	Slug        string
	Description string
	Locale      string
	SortOrder   int
}

func NewBlogCategoryService(repo *repository.BlogCategoryRepository) *BlogCategoryService {
	return &BlogCategoryService{repo: repo}
}

func (s *BlogCategoryService) List(locale string) ([]post.Category, error) {
	if s == nil || s.repo == nil {
		return []post.Category{}, nil
	}

	normalizedLocale := strings.TrimSpace(locale)
	if normalizedLocale != "" {
		var err error
		normalizedLocale, err = requireSupportedLocale(normalizedLocale)
		if err != nil {
			return nil, err
		}
	}
	return s.repo.List(normalizedLocale)
}

func (s *BlogCategoryService) Get(id uint) (*post.Category, error) {
	if s == nil || s.repo == nil {
		return nil, ErrBlogCategoryNotFound
	}
	category, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBlogCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (s *BlogCategoryService) Create(input BlogCategoryInput) (*post.Category, error) {
	if s == nil || s.repo == nil {
		return nil, ErrBlogCategoryNotFound
	}
	category, err := normalizeBlogCategoryInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.ExistsBySlugLocale(category.Slug, category.Locale, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrBlogCategorySlugExists
	}
	if err := s.repo.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *BlogCategoryService) Update(id uint, input BlogCategoryInput) (*post.Category, error) {
	if s == nil || s.repo == nil {
		return nil, ErrBlogCategoryNotFound
	}
	if _, err := s.Get(id); err != nil {
		return nil, err
	}
	category, err := normalizeBlogCategoryInput(input)
	if err != nil {
		return nil, err
	}
	category.ID = id
	exists, err := s.repo.ExistsBySlugLocale(category.Slug, category.Locale, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrBlogCategorySlugExists
	}
	if err := s.repo.Update(category); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *BlogCategoryService) Delete(id uint) error {
	if s == nil || s.repo == nil {
		return ErrBlogCategoryNotFound
	}
	if _, err := s.Get(id); err != nil {
		return err
	}
	count, err := s.repo.CountPosts(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrBlogCategoryInUse
	}
	return s.repo.Delete(id)
}

func (s *BlogCategoryService) ValidatePostCategories(categoryIDs []uint, locale string) error {
	if len(categoryIDs) == 0 {
		return nil
	}
	if s == nil || s.repo == nil {
		return ErrBlogCategoryNotFound
	}

	normalizedLocale, err := requireSupportedLocale(locale)
	if err != nil {
		return err
	}

	uniqueIDs := make(map[uint]struct{}, len(categoryIDs))
	normalizedIDs := make([]uint, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		if categoryID == 0 {
			continue
		}
		if _, exists := uniqueIDs[categoryID]; exists {
			continue
		}
		uniqueIDs[categoryID] = struct{}{}
		normalizedIDs = append(normalizedIDs, categoryID)
	}
	categories, err := s.repo.FindByIDs(normalizedIDs)
	if err != nil {
		return err
	}
	if len(categories) != len(normalizedIDs) {
		return ErrBlogCategoryNotFound
	}
	for _, category := range categories {
		if strings.TrimSpace(category.Locale) != normalizedLocale {
			return fmt.Errorf("%w: category %d belongs to locale %s", ErrBlogCategoryInvalid, category.ID, category.Locale)
		}
	}
	return nil
}

func (s *BlogCategoryService) ReplacePostCategories(postID uint, categoryIDs []uint) error {
	if s == nil || s.repo == nil {
		if len(categoryIDs) == 0 {
			return nil
		}
		return ErrBlogCategoryNotFound
	}
	categoryIDs = uniqueCategoryIDs(categoryIDs)
	return s.repo.ReplacePostCategories(postID, categoryIDs)
}

func normalizeBlogCategoryInput(input BlogCategoryInput) (*post.Category, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	locale, err := requireSupportedLocale(input.Locale)
	if err != nil {
		return nil, err
	}
	if name == "" || slug == "" || len([]rune(name)) > 120 || len([]rune(slug)) > 120 || !blogCategorySlugPattern.MatchString(slug) {
		return nil, fmt.Errorf("%w: name and a lowercase slug are required", ErrBlogCategoryInvalid)
	}
	return &post.Category{
		Name:        name,
		Slug:        slug,
		Description: strings.TrimSpace(input.Description),
		Locale:      locale,
		SortOrder:   input.SortOrder,
	}, nil
}

func uniqueCategoryIDs(categoryIDs []uint) []uint {
	seen := make(map[uint]struct{}, len(categoryIDs))
	result := make([]uint, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		if categoryID == 0 {
			continue
		}
		if _, exists := seen[categoryID]; exists {
			continue
		}
		seen[categoryID] = struct{}{}
		result = append(result, categoryID)
	}
	return result
}
