package service

import (
	"errors"
	"tanzanite/internal/domain/post"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/repository"
	"time"
)

type PostService struct {
	postRepo *repository.PostRepository
	cache    *cache.RedisCache
	cacheTTL time.Duration
}

func NewPostService(postRepo *repository.PostRepository, cache *cache.RedisCache, cacheTTL int) *PostService {
	return &PostService{
		postRepo: postRepo,
		cache:    cache,
		cacheTTL: time.Duration(cacheTTL) * time.Second,
	}
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

	_ = s.postRepo.IncrementViewCount(result.ID)

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, result, s.cacheTTL)
	}

	return result, nil
}

func (s *PostService) List(locale, status string, page, pageSize int) ([]post.Post, int64, error) {
	offset := (page - 1) * pageSize
	return s.postRepo.List(locale, status, offset, pageSize)
}

func (s *PostService) Create(p *post.Post) error {
	return s.postRepo.Create(p)
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

	return nil
}

func (s *PostService) GetPublishedPosts() ([]post.Post, error) {
	return s.postRepo.FindPublished()
}

func (s *PostService) GetPublishedPostsByLocale(locale string) ([]post.Post, error) {
	return s.postRepo.FindPublishedByLocale(locale)
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
