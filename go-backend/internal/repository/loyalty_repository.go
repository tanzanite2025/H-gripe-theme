package repository

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/loyalty"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LoyaltyRepository struct {
	db *gorm.DB
}

var (
	ErrInvalidUserID      = errors.New("user ID is required")
	ErrInvalidPoints      = errors.New("points must not be zero")
	ErrInsufficientPoints = errors.New("insufficient points")
)

func NewLoyaltyRepository(db *gorm.DB) *LoyaltyRepository {
	return &LoyaltyRepository{db: db}
}

// WithTx 使用指定的事务创建新的 repository 实例
func (r *LoyaltyRepository) WithTx(tx *gorm.DB) *LoyaltyRepository {
	return &LoyaltyRepository{db: tx}
}

// GetDB 获取底层 GORM DB 实例
// LoyaltyTransaction 相关方法

// CreateTransaction 创建积分交易
func (r *LoyaltyRepository) CreateTransaction(t *loyalty.LoyaltyTransaction) error {
	return r.db.Create(t).Error
}

// FindTransactionByID 根据ID查找交易
func (r *LoyaltyRepository) FindTransactionByID(id uint) (*loyalty.LoyaltyTransaction, error) {
	var t loyalty.LoyaltyTransaction
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindTransactionsByUserID 查找用户的积分交易记录
func (r *LoyaltyRepository) FindTransactionsByUserID(userID uint, page, pageSize int) ([]loyalty.LoyaltyTransaction, int64, error) {
	var transactions []loyalty.LoyaltyTransaction
	var total int64

	query := r.db.Model(&loyalty.LoyaltyTransaction{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&transactions).Error

	return transactions, total, err
}

// GetUserPointsBalance 获取用户积分余额
func (r *LoyaltyRepository) GetUserPointsBalance(userID uint) (int, error) {
	var userLoyalty loyalty.UserLoyalty
	err := r.db.Where("user_id = ?", userID).First(&userLoyalty).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return userLoyalty.AvailablePoints, nil
}

func (r *LoyaltyRepository) FindOrCreateUserLoyaltyForUpdate(userID uint) (*loyalty.UserLoyalty, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return findOrCreateUserLoyaltyForUpdate(r.db, userID)
}

func (r *LoyaltyRepository) SumTransactionPointsByUser(userID uint, transactionType, source string, start, end time.Time) (int, error) {
	var sumPoints int
	err := r.db.Model(&loyalty.LoyaltyTransaction{}).
		Where("user_id = ? AND type = ? AND source = ? AND created_at BETWEEN ? AND ?",
			userID, transactionType, source, start, end).
		Select("COALESCE(SUM(points), 0)").
		Scan(&sumPoints).Error
	return sumPoints, err
}

func (r *LoyaltyRepository) CountTransactionsByUserAndSource(userID uint, transactionType, source string) (int64, error) {
	var count int64
	err := r.db.Model(&loyalty.LoyaltyTransaction{}).
		Where("user_id = ? AND type = ? AND source = ?", userID, transactionType, source).
		Count(&count).Error
	return count, err
}

// AdjustUserPoints atomically updates a user's points summary and creates the matching ledger entry.
func (r *LoyaltyRepository) AdjustUserPoints(userID uint, points int, transactionType, source string, sourceID uint, description string) (*loyalty.LoyaltyTransaction, error) {
	var transaction *loyalty.LoyaltyTransaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		transaction, err = r.WithTx(tx).AdjustUserPointsInCurrentTx(userID, points, transactionType, source, sourceID, description)
		return err
	})
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (r *LoyaltyRepository) AdjustUserPointsInCurrentTx(userID uint, points int, transactionType, source string, sourceID uint, description string) (*loyalty.LoyaltyTransaction, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if points == 0 {
		return nil, ErrInvalidPoints
	}

	if transactionType == "" {
		transactionType = "adjust"
	}

	if source == "" {
		source = transactionType
	}

	userLoyalty, err := findOrCreateUserLoyaltyForUpdate(r.db, userID)
	if err != nil {
		return nil, err
	}

	if points < 0 && userLoyalty.AvailablePoints+points < 0 {
		return nil, ErrInsufficientPoints
	}

	applyPointsDelta(userLoyalty, points, transactionType)

	if err := r.db.Save(userLoyalty).Error; err != nil {
		return nil, fmt.Errorf("failed to update user loyalty: %w", err)
	}

	transaction := &loyalty.LoyaltyTransaction{
		UserID:      userID,
		Type:        transactionType,
		Points:      points,
		Balance:     userLoyalty.AvailablePoints,
		Source:      source,
		SourceID:    sourceID,
		Description: description,
	}

	if err := r.db.Create(transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to create loyalty transaction: %w", err)
	}

	return transaction, nil
}

func findOrCreateUserLoyaltyForUpdate(tx *gorm.DB, userID uint) (*loyalty.UserLoyalty, error) {
	var userLoyalty loyalty.UserLoyalty
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&userLoyalty).Error
	if err == nil {
		return &userLoyalty, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoNothing: true,
	}).Create(&loyalty.UserLoyalty{UserID: userID}).Error; err != nil {
		return nil, fmt.Errorf("failed to initialize user loyalty: %w", err)
	}

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&userLoyalty).Error; err != nil {
		return nil, err
	}

	return &userLoyalty, nil
}

func applyPointsDelta(userLoyalty *loyalty.UserLoyalty, points int, transactionType string) {
	if points > 0 {
		userLoyalty.AvailablePoints += points
		if transactionType == "refund" {
			userLoyalty.UsedPoints -= points
			if userLoyalty.UsedPoints < 0 {
				userLoyalty.UsedPoints = 0
			}
			return
		}
		userLoyalty.TotalPoints += points
		return
	}

	userLoyalty.AvailablePoints += points
	userLoyalty.UsedPoints += -points
}
