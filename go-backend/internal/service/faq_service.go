package service

import (
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/repository"
)

type FAQService struct {
	faqRepo *repository.FAQRepository
}

type FAQAdminUpdateInput struct {
	Question string
	Answer   string
	PageID   string
	Category string
	Locale   string
	Status   string
	Order    int
}

func NewFAQService(faqRepo *repository.FAQRepository) *FAQService {
	return &FAQService{
		faqRepo: faqRepo,
	}
}

// GetByID 根据ID获取FAQ
func (s *FAQService) GetByID(id uint) (*faq.FAQ, error) {
	return s.faqRepo.FindByID(id)
}

// List 获取FAQ列表
func (s *FAQService) List(locale, pageID, category, status string, page, pageSize int) ([]faq.FAQ, int64, error) {
	offset := (page - 1) * pageSize
	return s.faqRepo.List(locale, pageID, category, status, offset, pageSize)
}

func (s *FAQService) ListAdmin(locale, pageID, category, status, search string, page, pageSize int) ([]faq.FAQ, int64, error) {
	if search != "" {
		return s.Search(search, locale, page, pageSize)
	}
	return s.List(locale, pageID, category, status, page, pageSize)
}

// GetCategories 获取所有分类
func (s *FAQService) GetCategories(locale string) ([]string, error) {
	return s.faqRepo.GetCategories(locale)
}

// Create 创建FAQ
func (s *FAQService) Create(f *faq.FAQ) error {
	return s.faqRepo.Create(f)
}

// Update 更新FAQ
func (s *FAQService) Update(f *faq.FAQ) error {
	return s.faqRepo.Update(f)
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
	if input.PageID != "" {
		existingFAQ.PageID = input.PageID
	}
	if input.Category != "" {
		existingFAQ.Category = input.Category
	}
	if input.Locale != "" {
		existingFAQ.Locale = input.Locale
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

// Delete 删除FAQ
func (s *FAQService) Delete(id uint) error {
	return s.faqRepo.Delete(id)
}

func (s *FAQService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.Delete(id); err == nil {
			deleted++
		}
	}
	return deleted, nil
}

// Search 搜索FAQ
func (s *FAQService) Search(keyword, locale string, page, pageSize int) ([]faq.FAQ, int64, error) {
	offset := (page - 1) * pageSize
	return s.faqRepo.Search(keyword, locale, offset, pageSize)
}

// UpdateOrder 更新排序
func (s *FAQService) UpdateOrder(id uint, order int) error {
	return s.faqRepo.UpdateOrder(id, order)
}

// BatchUpdateOrder 批量更新排序
func (s *FAQService) BatchUpdateOrder(orders map[uint]int) error {
	return s.faqRepo.BatchUpdateOrder(orders)
}

// IncrementViewCount 增加浏览次数
func (s *FAQService) IncrementViewCount(id uint) error {
	return s.faqRepo.IncrementViewCount(id)
}

// GetByCategory 获取分类下的FAQ
func (s *FAQService) GetByCategory(category, locale string) ([]faq.FAQ, error) {
	return s.faqRepo.GetByCategory(category, locale)
}

// GetPopular 获取热门FAQ
func (s *FAQService) GetPopular(locale string, limit int) ([]faq.FAQ, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.faqRepo.GetPopular(locale, limit)
}
