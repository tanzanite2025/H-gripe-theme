package service

import (
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/pkg/safehtml"
	"errors"
	"time"
)

type PostCreateInput struct {
	Title              string
	Slug               string
	Content            string
	Excerpt            string
	Status             string
	AuthorID           uint
	Locale             string
	FeaturedImg        string
	Tags               string
	TranslationGroupID *uint
	CategoryIDs        []uint
}

type PostUpdateInput struct {
	Title                    *string
	Slug                     *string
	Content                  *string
	Excerpt                  *string
	Status                   *string
	Locale                   *string
	FeaturedImg              *string
	Tags                     *string
	TranslationGroupID       *uint
	UpdateTranslationGroupID bool
	CategoryIDs              []uint
	UpdateCategoryIDs        bool
}

func (s *PostService) ListAdmin(page, pageSize int, status, locale, search, authorID string) ([]post.Post, int64, error) {
	return s.postRepo.FindAllWithFilters(page, pageSize, status, locale, search, authorID)
}

func (s *PostService) normalizePostCategoryIDs(categoryIDs []uint, locale string) ([]uint, error) {
	categoryIDs = uniqueCategoryIDs(categoryIDs)
	if err := s.categoryService.ValidatePostCategories(categoryIDs, locale); err != nil {
		return nil, err
	}
	return categoryIDs, nil
}

func (s *PostService) GetAdminPost(id uint) (*post.Post, error) {
	foundPost, err := s.findPost(id)
	if err != nil {
		return nil, err
	}

	if foundPost.TranslationGroupID != nil {
		translations, err := s.postRepo.FindByTranslationGroup(*foundPost.TranslationGroupID)
		if err != nil {
			return nil, err
		}
		foundPost.Translations = sanitizePostSliceHTML(translations)
	}

	return sanitizePostHTML(foundPost), nil
}

func (s *PostService) GetStats() (map[string]interface{}, error) {
	return s.postRepo.GetStats()
}

func (s *PostService) CreateAdminPost(input PostCreateInput) (*post.Post, error) {
	locale, err := requireSupportedLocale(input.Locale)
	if err != nil {
		return nil, err
	}
	content, err := safehtml.Sanitize(input.Content)
	if err != nil {
		return nil, err
	}

	if err := s.ensureSlugAvailable(input.Slug, locale, 0); err != nil {
		return nil, err
	}
	categoryIDs, err := s.normalizePostCategoryIDs(input.CategoryIDs, locale)
	if err != nil {
		return nil, err
	}

	newPost := &post.Post{
		Title:              input.Title,
		Slug:               input.Slug,
		Content:            content,
		Excerpt:            input.Excerpt,
		Status:             input.Status,
		AuthorID:           input.AuthorID,
		Locale:             locale,
		FeaturedImg:        input.FeaturedImg,
		Tags:               input.Tags,
		TranslationGroupID: input.TranslationGroupID,
	}

	if input.Status == "published" {
		now := time.Now()
		newPost.PublishedAt = &now
	}

	if err := s.postRepo.Create(newPost); err != nil {
		return nil, err
	}
	if err := s.categoryService.ReplacePostCategories(newPost.ID, categoryIDs); err != nil {
		_ = s.postRepo.Delete(newPost.ID)
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("admin post create")

	return newPost, nil
}

func (s *PostService) UpdateAdminPost(id uint, input PostUpdateInput) (*post.Post, error) {
	existingPost, err := s.findPost(id)
	if err != nil {
		return nil, err
	}

	previousPost := *existingPost
	nextSlug := existingPost.Slug
	nextLocale := existingPost.Locale
	if input.Slug != nil {
		nextSlug = *input.Slug
	}
	if input.Locale != nil {
		locale, err := requireSupportedLocale(*input.Locale)
		if err != nil {
			return nil, err
		}
		currentLocale, err := requireSupportedLocale(existingPost.Locale)
		if err != nil {
			return nil, err
		}
		if locale != currentLocale {
			return nil, ErrPostLocaleImmutable
		}
		nextLocale = locale
	}
	if nextSlug != existingPost.Slug || nextLocale != existingPost.Locale {
		if err := s.ensureSlugAvailable(nextSlug, nextLocale, existingPost.ID); err != nil {
			return nil, err
		}
	}
	categoryIDs := existingPostCategoryIDs(existingPost)
	if input.UpdateCategoryIDs {
		categoryIDs, err = s.normalizePostCategoryIDs(input.CategoryIDs, nextLocale)
		if err != nil {
			return nil, err
		}
	}

	if input.Title != nil {
		existingPost.Title = *input.Title
	}
	if input.Slug != nil {
		existingPost.Slug = *input.Slug
	}
	if input.Content != nil {
		content, err := safehtml.Sanitize(*input.Content)
		if err != nil {
			return nil, err
		}
		existingPost.Content = content
	}
	if input.Excerpt != nil {
		existingPost.Excerpt = *input.Excerpt
	}
	if input.Status != nil {
		if *input.Status == "published" && existingPost.Status != "published" && existingPost.PublishedAt == nil {
			now := time.Now()
			existingPost.PublishedAt = &now
		}
		existingPost.Status = *input.Status
	}
	if input.Locale != nil {
		existingPost.Locale = nextLocale
	}
	if input.FeaturedImg != nil {
		existingPost.FeaturedImg = *input.FeaturedImg
	}
	if input.Tags != nil {
		existingPost.Tags = *input.Tags
	}
	if input.UpdateTranslationGroupID {
		existingPost.TranslationGroupID = input.TranslationGroupID
	}

	if err := s.postRepo.Update(existingPost); err != nil {
		return nil, err
	}
	if input.UpdateCategoryIDs {
		if err := s.categoryService.ReplacePostCategories(existingPost.ID, categoryIDs); err != nil {
			return nil, err
		}
	}

	s.clearPostCache(&previousPost)
	s.clearPostCache(existingPost)
	s.invalidateStorefrontHTMLCache("admin post update")

	return existingPost, nil
}

func existingPostCategoryIDs(existingPost *post.Post) []uint {
	if existingPost == nil || len(existingPost.Categories) == 0 {
		return []uint{}
	}
	ids := make([]uint, 0, len(existingPost.Categories))
	for _, category := range existingPost.Categories {
		ids = append(ids, category.ID)
	}
	return ids
}

func (s *PostService) Delete(id uint) error {
	return s.deletePostByID(id, true)
}

func (s *PostService) deletePostByID(id uint, shouldInvalidateHTML bool) error {
	existingPost, err := s.findPost(id)
	if err != nil {
		return err
	}

	if err := s.postRepo.Delete(id); err != nil {
		return err
	}

	s.clearPostCache(existingPost)
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin post delete")
	}

	return nil
}

func (s *PostService) UpdateStatus(id uint, status string) error {
	return s.updatePostStatusByID(id, status, true)
}

func (s *PostService) updatePostStatusByID(id uint, status string, shouldInvalidateHTML bool) error {
	existingPost, err := s.findPost(id)
	if err != nil {
		return err
	}

	if err := s.postRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	s.clearPostCache(existingPost)
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin post status update")
	}

	return nil
}

func (s *PostService) BatchUpdateStatus(ids []uint, status string) (int, error) {
	updated := 0
	for _, id := range ids {
		if err := s.updatePostStatusByID(id, status, false); err != nil {
			if errors.Is(err, ErrPostNotFound) {
				continue
			}
			return updated, err
		}
		updated++
	}
	if updated > 0 {
		s.invalidateStorefrontHTMLCache("admin post batch status update")
	}

	return updated, nil
}

func (s *PostService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.deletePostByID(id, false); err != nil {
			if errors.Is(err, ErrPostNotFound) {
				continue
			}
			return deleted, err
		}
		deleted++
	}
	if deleted > 0 {
		s.invalidateStorefrontHTMLCache("admin post batch delete")
	}

	return deleted, nil
}
