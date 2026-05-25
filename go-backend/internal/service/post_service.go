package service

import (
	"fmt"
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

// GetByID 根据ID获取文章
func (s *PostService) GetByID(id uint) (*post.Post, error) {
	cacheKey := fmt.Sprintf("post:%d", id)

	// 尝试从缓存获取
	var p post.Post
	if err := s.cache.Get(cacheKey, &p); err == nil {
		return &p, nil
	}

	// 从数据库获取
	result, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 增加浏览次数
	_ = s.postRepo.IncrementViewCount(id)

	// 写入缓存
	_ = s.cache.Set(cacheKey, result, s.cacheTTL)

	return result, nil
}

// GetBySlug 根据slug获取文章
func (s *PostService) GetBySlug(slug, locale string) (*post.Post, error) {
	cacheKey := fmt.Sprintf("post:slug:%s:%s", slug, locale)

	// 尝试从缓存获取
	var p post.Post
	if err := s.cache.Get(cacheKey, &p); err == nil {
		return &p, nil
	}

	// 从数据库获取
	result, err := s.postRepo.FindBySlug(slug, locale)
	if err != nil {
		return nil, err
	}

	// 增加浏览次数
	_ = s.postRepo.IncrementViewCount(result.ID)

	// 写入缓存
	_ = s.cache.Set(cacheKey, result, s.cacheTTL)

	return result, nil
}

// List 获取文章列表
func (s *PostService) List(locale, status string, page, pageSize int) ([]post.Post, int64, error) {
	offset := (page - 1) * pageSize
	return s.postRepo.List(locale, status, offset, pageSize)
}

// Create 创建文章
func (s *PostService) Create(p *post.Post) error {
	return s.postRepo.Create(p)
}

// Update 更新文章
func (s *PostService) Update(p *post.Post) error {
	if err := s.postRepo.Update(p); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("post:%d", p.ID)
	_ = s.cache.Delete(cacheKey)

	slugCacheKey := fmt.Sprintf("post:slug:%s:%s", p.Slug, p.Locale)
	_ = s.cache.Delete(slugCacheKey)

	return nil
}

// Delete 删除文章
func (s *PostService) Delete(id uint) error {
	p, err := s.postRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.postRepo.Delete(id); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("post:%d", id)
	_ = s.cache.Delete(cacheKey)

	slugCacheKey := fmt.Sprintf("post:slug:%s:%s", p.Slug, p.Locale)
	_ = s.cache.Delete(slugCacheKey)

	return nil
}

// GetTranslations 获取文章的所有翻译版本
func (s *PostService) GetTranslations(postID uint) ([]post.Post, error) {
	// 先获取文章的翻译组ID
	groupID, err := s.postRepo.GetTranslationGroupID(postID)
	if err != nil {
		return nil, err
	}

	// 如果没有翻译组ID，返回空列表
	if groupID == nil {
		return []post.Post{}, nil
	}

	// 获取翻译组中的所有文章
	translations, err := s.postRepo.FindByTranslationGroup(*groupID)
	if err != nil {
		return nil, err
	}

	return translations, nil
}

// GetTranslationsByGroup 根据翻译组ID获取所有翻译版本
func (s *PostService) GetTranslationsByGroup(groupID uint) ([]post.Post, error) {
	return s.postRepo.FindByTranslationGroup(groupID)
}

// GetPublishedPosts 获取所有已发布的文章
func (s *PostService) GetPublishedPosts() ([]post.Post, error) {
	return s.postRepo.FindPublished()
}

// GetPublishedPostsByLocale 获取指定语言的已发布文章
func (s *PostService) GetPublishedPostsByLocale(locale string) ([]post.Post, error) {
	return s.postRepo.FindPublishedByLocale(locale)
}
