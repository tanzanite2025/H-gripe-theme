package repository

import (
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/suggestionfeedback"

	"gorm.io/gorm"
)

type SuggestionFeedbackRepository struct {
	db *gorm.DB
}

func NewSuggestionFeedbackRepository(db *gorm.DB) *SuggestionFeedbackRepository {
	return &SuggestionFeedbackRepository{db: db}
}

func (r *SuggestionFeedbackRepository) Create(item *suggestionfeedback.SuggestionFeedback) error {
	return r.db.Create(item).Error
}

func (r *SuggestionFeedbackRepository) List(status, search string, page, pageSize int) ([]suggestionfeedback.SuggestionFeedback, int64, error) {
	var items []suggestionfeedback.SuggestionFeedback
	var total int64

	query := r.db.Model(&suggestionfeedback.SuggestionFeedback{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("full_name LIKE ? OR email LIKE ? OR message LIKE ? OR order_number LIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *SuggestionFeedbackRepository) GetUserMemberLevelName(userID uint) (string, error) {
	var userLoyalty loyalty.UserLoyalty
	if err := r.db.Where("user_id = ?", userID).First(&userLoyalty).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	if userLoyalty.MemberLevelID == 0 {
		return "", nil
	}

	var level loyalty.MemberLevel
	if err := r.db.First(&level, userLoyalty.MemberLevelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return level.Name, nil
}
