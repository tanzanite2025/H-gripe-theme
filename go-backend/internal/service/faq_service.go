package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
)

type FAQService struct {
	faqRepo                          *repository.FAQRepository
	storefrontHTMLCacheInvalidator   *StorefrontHTMLCacheInvalidator
	storefrontContentReleaseNotifier *StorefrontContentReleaseNotifier
	storage                          storage.StorageService
}

func NewFAQService(faqRepo *repository.FAQRepository, storageSvc storage.StorageService) *FAQService {
	return &FAQService{
		faqRepo: faqRepo,
		storage: storageSvc,
	}
}

func (s *FAQService) SetStorefrontHTMLCacheInvalidator(invalidator *StorefrontHTMLCacheInvalidator) {
	s.storefrontHTMLCacheInvalidator = invalidator
}

func (s *FAQService) SetStorefrontContentReleaseNotifier(notifier *StorefrontContentReleaseNotifier) {
	s.storefrontContentReleaseNotifier = notifier
}

func (s *FAQService) notifyStorefrontContentChange(reason string) {
	if s.storefrontHTMLCacheInvalidator != nil {
		s.storefrontHTMLCacheInvalidator.PurgeAllAsync(reason)
	}
	if s.storefrontContentReleaseNotifier != nil {
		s.storefrontContentReleaseNotifier.TriggerAsync(reason)
	}
}

func hasPublishedFAQStatus(items ...*faq.FAQ) bool {
	for _, item := range items {
		if item != nil && normalizeFAQStatus(item.Status, "") == "published" {
			return true
		}
	}
	return false
}

func (s *FAQService) UploadAnswerImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("FAQ image storage is not configured")
	}
	return s.storage.Upload(ctx, file)
}

// GetByID 根据ID获取FAQ
func (s *FAQService) GetByID(id uint) (*faq.FAQ, error) {
	return s.faqRepo.FindByID(id)
}

// List 获取FAQ列表
func (s *FAQService) List(locale, pageID, category, status string, page, pageSize int) ([]faq.FAQ, int64, error) {
	offset := (page - 1) * pageSize
	if locale != "" {
		locale = normalizeLocale(locale)
	}
	items, total, err := s.faqRepo.List(locale, pageID, category, status, offset, pageSize)
	return sanitizeFAQSliceForPublic(items), total, err
}

func (s *FAQService) ListAdmin(locale, pageID, category, status, search string, page, pageSize int) ([]faq.FAQ, int64, error) {
	offset := (page - 1) * pageSize
	if locale != "" {
		locale = normalizeLocale(locale)
	}
	return s.faqRepo.ListAdmin(locale, pageID, category, status, search, offset, pageSize)
}

// GetCategories 获取所有分类
func (s *FAQService) GetCategories(locale string) ([]string, error) {
	if locale != "" {
		locale = normalizeLocale(locale)
	}
	return s.faqRepo.GetCategories(locale)
}

// Create 创建FAQ
func (s *FAQService) Create(f *faq.FAQ) error {
	if err := s.normalizeFAQContent(f); err != nil {
		return err
	}
	locale, err := requireSupportedLocale(f.Locale)
	if err != nil {
		return err
	}
	f.Locale = locale
	if err := s.validateFAQPlacement(f.PageID, f.Category, f.Locale); err != nil {
		return err
	}
	if err := s.faqRepo.Create(f); err != nil {
		return err
	}
	if hasPublishedFAQStatus(f) {
		s.notifyStorefrontContentChange("admin faq create")
	}
	return nil
}

// Update 更新FAQ
func (s *FAQService) Update(f *faq.FAQ) error {
	previousFAQ, err := s.faqRepo.FindByID(f.ID)
	if err != nil {
		return err
	}

	if err := s.normalizeFAQContent(f); err != nil {
		return err
	}
	locale, err := requireSupportedLocale(f.Locale)
	if err != nil {
		return err
	}
	f.Locale = locale
	if err := s.validateFAQPlacement(f.PageID, f.Category, f.Locale); err != nil {
		return err
	}
	if err := s.faqRepo.Update(f); err != nil {
		return err
	}
	if hasPublishedFAQStatus(previousFAQ, f) {
		s.notifyStorefrontContentChange("admin faq update")
	}
	return nil
}

func (s *FAQService) UpdateAdminFAQ(id uint, input FAQAdminUpdateInput) (*faq.FAQ, error) {
	existingFAQ, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if input.Question != "" {
		existingFAQ.Question = input.Question
	}
	if input.Answer != "" {
		existingFAQ.Answer = input.Answer
	}
	if input.AnswerImageSet {
		existingFAQ.AnswerImageURL = input.AnswerImageURL
		existingFAQ.AnswerImageAlt = input.AnswerImageAlt
		existingFAQ.AnswerImageWidth = input.AnswerImageWidth
		existingFAQ.AnswerImageHeight = input.AnswerImageHeight
	}
	if input.PageID != "" {
		existingFAQ.PageID = input.PageID
	}
	if input.Category != "" {
		existingFAQ.Category = input.Category
	}
	if input.Locale != "" {
		locale, err := validateFAQLocaleUpdate(existingFAQ.Locale, input.Locale)
		if err != nil {
			return nil, err
		}
		existingFAQ.Locale = locale
	}
	if input.Status != "" {
		existingFAQ.Status = input.Status
	}
	existingFAQ.Order = input.Order

	if err := s.Update(existingFAQ); err != nil {
		return nil, err
	}

	return existingFAQ, nil
}

func validateFAQLocaleUpdate(existingLocale, requestedLocale string) (string, error) {
	currentLocale, err := requireSupportedLocale(existingLocale)
	if err != nil {
		return "", err
	}

	nextLocale, err := requireSupportedLocale(requestedLocale)
	if err != nil {
		return "", err
	}

	if nextLocale != currentLocale {
		return "", fmt.Errorf("%w: %s -> %s", ErrFAQLocaleImmutable, currentLocale, nextLocale)
	}

	return currentLocale, nil
}

// Delete 删除FAQ
func (s *FAQService) Delete(id uint) error {
	return s.deleteFAQByID(id, true)
}

func (s *FAQService) deleteFAQByID(id uint, shouldInvalidateHTML bool) error {
	item, err := s.faqRepo.FindByID(id)
	if err != nil {
		return err
	}
	if err := s.faqRepo.Delete(id); err != nil {
		return err
	}
	if shouldInvalidateHTML && hasPublishedFAQStatus(item) {
		s.notifyStorefrontContentChange("admin faq delete")
	}
	return nil
}

func (s *FAQService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	shouldNotify := false
	for _, id := range ids {
		item, err := s.faqRepo.FindByID(id)
		if err != nil {
			continue
		}
		if err := s.faqRepo.Delete(id); err == nil {
			deleted++
			shouldNotify = shouldNotify || hasPublishedFAQStatus(item)
		}
	}
	if deleted > 0 && shouldNotify {
		s.notifyStorefrontContentChange("admin faq batch delete")
	}
	return deleted, nil
}

// Search 搜索FAQ
func (s *FAQService) Search(keyword, locale string, page, pageSize int) ([]faq.FAQ, int64, error) {
	offset := (page - 1) * pageSize
	if locale != "" {
		locale = normalizeLocale(locale)
	}
	items, total, err := s.faqRepo.Search(keyword, locale, offset, pageSize)
	return sanitizeFAQSliceForPublic(items), total, err
}

// UpdateOrder 更新排序
func (s *FAQService) UpdateOrder(id uint, order int) error {
	item, err := s.faqRepo.FindByID(id)
	if err != nil {
		return err
	}
	if err := s.faqRepo.UpdateOrder(id, order); err != nil {
		return err
	}
	if hasPublishedFAQStatus(item) {
		s.notifyStorefrontContentChange("admin faq order update")
	}
	return nil
}

// BatchUpdateOrder 批量更新排序
func (s *FAQService) BatchUpdateOrder(orders map[uint]int) error {
	shouldNotify := false
	for id := range orders {
		item, err := s.faqRepo.FindByID(id)
		if err == nil && hasPublishedFAQStatus(item) {
			shouldNotify = true
			break
		}
	}
	if err := s.faqRepo.BatchUpdateOrder(orders); err != nil {
		return err
	}
	if len(orders) > 0 && shouldNotify {
		s.notifyStorefrontContentChange("admin faq batch order update")
	}
	return nil
}

// IncrementViewCount 增加浏览次数
func (s *FAQService) IncrementViewCount(id uint) error {
	return s.faqRepo.IncrementViewCount(id)
}

// GetByCategory 获取分类下的FAQ
func (s *FAQService) GetByCategory(category, locale string) ([]faq.FAQ, error) {
	if locale != "" {
		locale = normalizeLocale(locale)
	}
	items, err := s.faqRepo.GetByCategory(category, locale)
	return sanitizeFAQSliceForPublic(items), err
}

// GetPopular 获取热门FAQ
func (s *FAQService) GetPopular(locale string, limit int) ([]faq.FAQ, error) {
	if limit <= 0 {
		limit = 10
	}
	if locale != "" {
		locale = normalizeLocale(locale)
	}
	items, err := s.faqRepo.GetPopular(locale, limit)
	return sanitizeFAQSliceForPublic(items), err
}
