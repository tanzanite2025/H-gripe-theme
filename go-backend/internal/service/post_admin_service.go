package service

import (
	"errors"
	"tanzanite/internal/domain/post"
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
	MetaTitle          string
	MetaDesc           string
	MetaKeywords       string
	CanonicalURL       string
	TranslationGroupID *uint
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
	MetaTitle                *string
	MetaDesc                 *string
	MetaKeywords             *string
	CanonicalURL             *string
	TranslationGroupID       *uint
	UpdateTranslationGroupID bool
}

func (s *PostService) ListAdmin(page, pageSize int, status, locale, search, authorID string) ([]post.Post, int64, error) {
	return s.postRepo.FindAllWithFilters(page, pageSize, status, locale, search, authorID)
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
		foundPost.Translations = translations
	}

	return foundPost, nil
}

func (s *PostService) GetStats() (map[string]interface{}, error) {
	return s.postRepo.GetStats()
}

func (s *PostService) CreateAdminPost(input PostCreateInput) (*post.Post, error) {
	if err := s.ensureSlugAvailable(input.Slug, input.Locale, 0); err != nil {
		return nil, err
	}

	newPost := &post.Post{
		Title:              input.Title,
		Slug:               input.Slug,
		Content:            input.Content,
		Excerpt:            input.Excerpt,
		Status:             input.Status,
		AuthorID:           input.AuthorID,
		Locale:             input.Locale,
		FeaturedImg:        input.FeaturedImg,
		Tags:               input.Tags,
		MetaTitle:          input.MetaTitle,
		MetaDesc:           input.MetaDesc,
		MetaKeywords:       input.MetaKeywords,
		CanonicalURL:       input.CanonicalURL,
		TranslationGroupID: input.TranslationGroupID,
	}

	if input.Status == "published" {
		now := time.Now()
		newPost.PublishedAt = &now
	}

	if err := s.postRepo.Create(newPost); err != nil {
		return nil, err
	}

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
		nextLocale = *input.Locale
	}
	if nextSlug != existingPost.Slug || nextLocale != existingPost.Locale {
		if err := s.ensureSlugAvailable(nextSlug, nextLocale, existingPost.ID); err != nil {
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
		existingPost.Content = *input.Content
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
		existingPost.Locale = *input.Locale
	}
	if input.FeaturedImg != nil {
		existingPost.FeaturedImg = *input.FeaturedImg
	}
	if input.Tags != nil {
		existingPost.Tags = *input.Tags
	}
	if input.MetaTitle != nil {
		existingPost.MetaTitle = *input.MetaTitle
	}
	if input.MetaDesc != nil {
		existingPost.MetaDesc = *input.MetaDesc
	}
	if input.MetaKeywords != nil {
		existingPost.MetaKeywords = *input.MetaKeywords
	}
	if input.CanonicalURL != nil {
		existingPost.CanonicalURL = *input.CanonicalURL
	}
	if input.UpdateTranslationGroupID {
		existingPost.TranslationGroupID = input.TranslationGroupID
	}

	if err := s.postRepo.Update(existingPost); err != nil {
		return nil, err
	}

	s.clearPostCache(&previousPost)
	s.clearPostCache(existingPost)

	return existingPost, nil
}

func (s *PostService) Delete(id uint) error {
	existingPost, err := s.findPost(id)
	if err != nil {
		return err
	}

	if err := s.postRepo.Delete(id); err != nil {
		return err
	}

	s.clearPostCache(existingPost)

	return nil
}

func (s *PostService) UpdateStatus(id uint, status string) error {
	existingPost, err := s.findPost(id)
	if err != nil {
		return err
	}

	if err := s.postRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	s.clearPostCache(existingPost)

	return nil
}

func (s *PostService) BatchUpdateStatus(ids []uint, status string) (int, error) {
	updated := 0
	for _, id := range ids {
		if err := s.UpdateStatus(id, status); err != nil {
			if errors.Is(err, ErrPostNotFound) {
				continue
			}
			return updated, err
		}
		updated++
	}

	return updated, nil
}

func (s *PostService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			if errors.Is(err, ErrPostNotFound) {
				continue
			}
			return deleted, err
		}
		deleted++
	}

	return deleted, nil
}
