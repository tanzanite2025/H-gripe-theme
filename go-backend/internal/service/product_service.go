package service

import (
	"fmt"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/repository"
	"time"
)

type ProductService struct {
	productRepo *repository.ProductRepository
	cache       *cache.RedisCache
	cacheTTL    time.Duration
}

func NewProductService(productRepo *repository.ProductRepository, cache *cache.RedisCache, cacheTTL int) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		cache:       cache,
		cacheTTL:    time.Duration(cacheTTL) * time.Second,
	}
}

// GetByID 根据ID获取产品
func (s *ProductService) GetByID(id uint) (*product.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	// 尝试从缓存获取
	var p product.Product
	if err := s.cache.Get(cacheKey, &p); err == nil {
		return &p, nil
	}

	// 从数据库获取
	result, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 增加浏览次数
	_ = s.productRepo.IncrementViewCount(id)

	// 写入缓存
	_ = s.cache.Set(cacheKey, result, s.cacheTTL)

	return result, nil
}

// GetBySlug 根据slug获取产品
func (s *ProductService) GetBySlug(slug, locale string) (*product.Product, error) {
	cacheKey := fmt.Sprintf("product:slug:%s:%s", slug, locale)

	// 尝试从缓存获取
	var p product.Product
	if err := s.cache.Get(cacheKey, &p); err == nil {
		return &p, nil
	}

	// 从数据库获取
	result, err := s.productRepo.FindBySlug(slug, locale)
	if err != nil {
		return nil, err
	}

	// 增加浏览次数
	_ = s.productRepo.IncrementViewCount(result.ID)

	// 写入缓存
	_ = s.cache.Set(cacheKey, result, s.cacheTTL)

	return result, nil
}

// List 获取产品列表
func (s *ProductService) List(locale, status string, featured bool, page, pageSize int) ([]product.Product, int64, error) {
	offset := (page - 1) * pageSize
	return s.productRepo.List(locale, status, featured, offset, pageSize)
}

func (s *ProductService) SearchPublic(locale, status, keyword string, page, pageSize int) ([]product.Product, int64, error) {
	offset := (page - 1) * pageSize
	return s.productRepo.SearchPublic(locale, status, keyword, offset, pageSize)
}

// Create 创建产品
func (s *ProductService) Create(p *product.Product) error {
	return s.productRepo.Create(p)
}

// Update 更新产品
func (s *ProductService) Update(p *product.Product) error {
	if err := s.productRepo.Update(p); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("product:%d", p.ID)
	_ = s.cache.Delete(cacheKey)

	slugCacheKey := fmt.Sprintf("product:slug:%s:%s", p.Slug, p.Locale)
	_ = s.cache.Delete(slugCacheKey)

	return nil
}

// Delete 删除产品
func (s *ProductService) Delete(id uint) error {
	p, err := s.productRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("product:%d", id)
	_ = s.cache.Delete(cacheKey)

	slugCacheKey := fmt.Sprintf("product:slug:%s:%s", p.Slug, p.Locale)
	_ = s.cache.Delete(slugCacheKey)

	return nil
}

// UpdateStock 更新库存
func (s *ProductService) UpdateStock(id uint, quantity int) error {
	if err := s.productRepo.UpdateStock(id, quantity); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("product:%d", id)
	_ = s.cache.Delete(cacheKey)

	return nil
}

// GetAttributeByID 根据ID获取属性
func (s *ProductService) GetAttributeByID(id uint) (*product.ProductAttribute, error) {
	return s.productRepo.FindAttributeByID(id)
}

// GetAttributeBySlug 根据Slug获取属性
func (s *ProductService) GetAttributeBySlug(slug string) (*product.ProductAttribute, error) {
	return s.productRepo.FindAttributeBySlug(slug)
}

// ListAttributes 获取属性分页列表
func (s *ProductService) ListAttributes(page, pageSize int) ([]product.ProductAttribute, int64, error) {
	return s.productRepo.FindAllAttributes(page, pageSize)
}

// CreateAttribute 创建属性
func (s *ProductService) CreateAttribute(attr *product.ProductAttribute) error {
	return s.productRepo.CreateAttribute(attr)
}

// UpdateAttribute 更新属性
func (s *ProductService) UpdateAttribute(attr *product.ProductAttribute) error {
	return s.productRepo.UpdateAttribute(attr)
}

// DeleteAttribute 删除属性
func (s *ProductService) DeleteAttribute(id uint) error {
	return s.productRepo.DeleteAttribute(id)
}

// GetFilterableAttributes 获取前台可过滤属性列表
func (s *ProductService) GetFilterableAttributes() ([]product.ProductAttribute, error) {
	return s.productRepo.FindFilterableAttributes()
}

// GetAttributeValueByID 获取属性值
func (s *ProductService) GetAttributeValueByID(id uint) (*product.AttributeValue, error) {
	return s.productRepo.FindAttributeValueByID(id)
}

// CreateAttributeValue 创建属性值
func (s *ProductService) CreateAttributeValue(val *product.AttributeValue) error {
	return s.productRepo.CreateAttributeValue(val)
}

// UpdateAttributeValue 更新属性值
func (s *ProductService) UpdateAttributeValue(val *product.AttributeValue) error {
	return s.productRepo.UpdateAttributeValue(val)
}

// DeleteAttributeValue 删除属性值
func (s *ProductService) DeleteAttributeValue(id uint) error {
	return s.productRepo.DeleteAttributeValue(id)
}

// GetValuesByAttributeID 获取属性对应的所有属性值
func (s *ProductService) GetValuesByAttributeID(attrID uint) ([]product.AttributeValue, error) {
	return s.productRepo.FindValuesByAttributeID(attrID)
}
