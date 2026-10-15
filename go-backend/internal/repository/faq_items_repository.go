package repository

import (
	"commerce-platform/internal/domain/faq"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// applyPublicFAQPageScope hides FAQ rows whose database-owned page is hidden
// or no longer backed by a current storefront route. Rows without a page
// record remain visible for backwards compatibility with legacy/global FAQs.
func applyPublicFAQPageScope(query *gorm.DB) *gorm.DB {
	return query.Where(`NOT EXISTS (
		SELECT 1
		FROM faq_pages AS public_faq_page
		WHERE public_faq_page.page_id = faqs.page_id
		  AND public_faq_page.locale = faqs.locale
		  AND public_faq_page.deleted_at IS NULL
		  AND (
			public_faq_page.status <> 'active'
			OR public_faq_page.route_status IN ('stale', 'missing')
		  )
	)`)
}

// Create 创建FAQ
func (r *FAQRepository) Create(f *faq.FAQ) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.validateAndCanonicalizeFAQAnswerLinksAndImageInOneAssetLockPlan(tx, f, &faq.FAQ{}); err != nil {
			return err
		}
		return tx.Create(f).Error
	})
}

// FindByID 根据ID查找FAQ
func (r *FAQRepository) FindByID(id uint) (*faq.FAQ, error) {
	var f faq.FAQ
	err := r.db.First(&f, id).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// Update 更新FAQ
func (r *FAQRepository) Update(f *faq.FAQ, mutations ...FAQMediaFieldMutation) error {
	if f.ID == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var previous faq.FAQ
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&previous, f.ID).Error; err != nil {
			return err
		}
		if len(mutations) > 0 {
			if !mutations[0].UpdateAnswer {
				f.Answer = previous.Answer
			}
			if !mutations[0].UpdateAnswerImage {
				f.AnswerImageURL, f.AnswerImageAlt = previous.AnswerImageURL, previous.AnswerImageAlt
				f.AnswerImageWidth, f.AnswerImageHeight = previous.AnswerImageWidth, previous.AnswerImageHeight
			}
		}
		if err := r.validateAndCanonicalizeFAQAnswerLinksAndImageInOneAssetLockPlan(tx, f, &previous); err != nil {
			return err
		}
		// Updates must never upsert a deleted FAQ.
		return tx.Model(&faq.FAQ{}).Where("id = ?", f.ID).Select("*").Omit("id", "created_at", "deleted_at").Updates(f).Error
	})
}

// Delete 删除FAQ（软删除）
func (r *FAQRepository) Delete(id uint) error {
	return r.db.Delete(&faq.FAQ{}, id).Error
}

// List 获取FAQ列表
func (r *FAQRepository) List(locale, pageID, status string, offset, limit int) ([]faq.FAQ, int64, error) {
	var faqs []faq.FAQ
	var total int64

	query := applyPublicFAQPageScope(r.db.Model(&faq.FAQ{}))

	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if pageID != "" {
		query = query.Where("page_id = ?", pageID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}}).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&faqs).Error
	return faqs, total, err
}

func (r *FAQRepository) ListAdmin(locale, pageID, status, search string, offset, limit int) ([]faq.FAQ, int64, error) {
	var faqs []faq.FAQ
	var total int64

	query := r.db.Model(&faq.FAQ{})

	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if pageID != "" {
		query = query.Where("page_id = ?", pageID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("question LIKE ? OR answer LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}}).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&faqs).Error
	return faqs, total, err
}

func (r *FAQRepository) ListForPage(locale, pageID, status string) ([]faq.FAQ, error) {
	var faqs []faq.FAQ
	query := r.db.Model(&faq.FAQ{}).
		Where("locale = ? AND page_id = ?", locale, pageID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}}).
		Order("created_at DESC").
		Find(&faqs).Error
	return faqs, err
}

// Search 搜索FAQ
func (r *FAQRepository) Search(keyword, locale string, offset, limit int) ([]faq.FAQ, int64, error) {
	var faqs []faq.FAQ
	var total int64

	query := applyPublicFAQPageScope(r.db.Model(&faq.FAQ{})).Where("status = ?", "published")

	if locale != "" {
		query = query.Where("locale = ?", locale)
	}

	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where("question LIKE ? OR answer LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}}).
		Order("view_count DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&faqs).Error

	return faqs, total, err
}

// UpdateOrder 更新排序
func (r *FAQRepository) UpdateOrder(id uint, order int) error {
	return r.db.Model(&faq.FAQ{}).Where("id = ?", id).Update("order", order).Error
}

// BatchUpdateOrder 批量更新排序
func (r *FAQRepository) BatchUpdateOrder(orders map[uint]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for id, order := range orders {
			if err := tx.Model(&faq.FAQ{}).Where("id = ?", id).Update("order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// IncrementViewCount 增加浏览次数
func (r *FAQRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&faq.FAQ{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// GetPopular 获取热门FAQ
func (r *FAQRepository) GetPopular(locale string, limit int) ([]faq.FAQ, error) {
	var faqs []faq.FAQ
	query := applyPublicFAQPageScope(r.db.Model(&faq.FAQ{})).Where("status = ?", "published")

	if locale != "" {
		query = query.Where("locale = ?", locale)
	}

	err := query.Order("view_count DESC, created_at DESC").
		Limit(limit).
		Find(&faqs).Error
	return faqs, err
}
