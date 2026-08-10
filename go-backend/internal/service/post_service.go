package service

import (
	"errors"
	"tanzanite/internal/domain/post"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/pkg/safehtml"
	"tanzanite/internal/repository"
	"time"
)

type PostService struct {
	postRepo                       *repository.PostRepository
	cache                          *cache.RedisCache
	cacheTTL                       time.Duration
	storefrontHTMLCacheInvalidator *StorefrontHTMLCacheInvalidator
}

func NewPostService(postRepo *repository.PostRepository, cache *cache.RedisCache, cacheTTL int) *PostService {
	return &PostService{
		postRepo: postRepo,
		cache:    cache,
		cacheTTL: time.Duration(cacheTTL) * time.Second,
	}
}

func (s *PostService) SetStorefrontHTMLCacheInvalidator(invalidator *StorefrontHTMLCacheInvalidator) {
	s.storefrontHTMLCacheInvalidator = invalidator
}

var (
	ErrPostNotFound   = errors.New("post not found")
	ErrPostSlugExists = errors.New("post slug already exists")
)

func (s *PostService) GetByID(id uint) (*post.Post, error) {
	cacheKey := postIDCacheKey(id)

	var cachedPost post.Post
	if s.cache != nil && s.cache.Get(cacheKey, &cachedPost) == nil {
		return &cachedPost, nil
	}

	result, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	result = sanitizePostHTML(result)

	_ = s.postRepo.IncrementViewCount(id)

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, result, s.cacheTTL)
	}

	return result, nil
}

func (s *PostService) GetBySlug(slug, locale string) (*post.Post, error) {
	cacheKey := postSlugCacheKey(slug, locale)

	var cachedPost post.Post
	if s.cache != nil && s.cache.Get(cacheKey, &cachedPost) == nil {
		return &cachedPost, nil
	}

	result, err := s.postRepo.FindBySlug(slug, locale)
	if err != nil {
		return nil, err
	}
	result = sanitizePostHTML(result)

	_ = s.postRepo.IncrementViewCount(result.ID)

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, result, s.cacheTTL)
	}

	return result, nil
}

func (s *PostService) GetPublicByID(id uint) (*post.Post, error) {
	result, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if result.Status != "published" {
		return nil, ErrPostNotFound
	}
	result = sanitizePostHTML(result)
	_ = s.postRepo.IncrementViewCount(id)
	return result, nil
}

func (s *PostService) GetPublicBySlug(slug, locale string) (*post.Post, error) {
	result, _, err := s.GetPublicBySlugWithRoutes(slug, locale)
	return result, err
}

func (s *PostService) List(locale, status string, page, pageSize int) ([]post.Post, int64, error) {
	offset := (page - 1) * pageSize
	posts, total, err := s.postRepo.List(locale, status, offset, pageSize)
	return sanitizePostSliceHTML(posts), total, err
}

func (s *PostService) ListPublic(locale string, page, pageSize int) ([]post.Post, int64, error) {
	return s.List(locale, "published", page, pageSize)
}

func (s *PostService) Create(p *post.Post) error {
	if err := s.postRepo.Create(p); err != nil {
		return err
	}
	s.invalidateStorefrontHTMLCache("post create")
	return nil
}

func (s *PostService) Update(p *post.Post) error {
	previousPost, err := s.findPost(p.ID)
	if err != nil {
		return err
	}

	if err := s.postRepo.Update(p); err != nil {
		return err
	}

	s.clearPostCache(previousPost)
	s.clearPostCache(p)
	s.invalidateStorefrontHTMLCache("post update")

	return nil
}

func (s *PostService) GetPublishedPosts() ([]post.Post, error) {
	posts, err := s.postRepo.FindPublished()
	return sanitizePostSliceHTML(posts), err
}

func (s *PostService) GetPublishedPostsByLocale(locale string) ([]post.Post, error) {
	posts, err := s.postRepo.FindPublishedByLocale(locale)
	return sanitizePostSliceHTML(posts), err
}

func sanitizePostHTML(p *post.Post) *post.Post {
	if p == nil {
		return nil
	}
	if sanitized, err := safehtml.Sanitize(p.Content); err == nil {
		p.Content = sanitized
	}
	return p
}

func sanitizePostSliceHTML(posts []post.Post) []post.Post {
	for index := range posts {
		if sanitized, err := safehtml.Sanitize(posts[index].Content); err == nil {
			posts[index].Content = sanitized
		}
	}
	return posts
}

func (s *PostService) findPost(id uint) (*post.Post, error) {
	foundPost, err := s.postRepo.FindByID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return foundPost, nil
}

func (s *PostService) ensureSlugAvailable(slug, locale string, currentPostID uint) error {
	existingPost, err := s.postRepo.FindBySlug(slug, locale)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil
		}
		return err
	}

	if existingPost != nil && existingPost.ID != currentPostID {
		return ErrPostSlugExists
	}

	return nil
}
