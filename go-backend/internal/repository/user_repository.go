package repository

import (
	"tanzanite/internal/domain/user"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(u *user.User) error {
	return r.db.Create(u).Error
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(id uint) (*user.User, error) {
	var u user.User
	err := r.db.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(email string) (*user.User, error) {
	var u user.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*user.User, error) {
	var u user.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Update 更新用户
func (r *UserRepository) Update(u *user.User) error {
	return r.db.Save(u).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&user.User{}, id).Error
}

// List 获取用户列表
func (r *UserRepository) List(offset, limit int) ([]user.User, int64, error) {
	var users []user.User
	var total int64

	if err := r.db.Model(&user.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

// FindAllWithFilters 根据筛选条件获取用户列表
func (r *UserRepository) FindAllWithFilters(page, pageSize int, role, status, search string) ([]user.User, int64, error) {
	var users []user.User
	var total int64

	query := r.db.Model(&user.User{})

	// 应用筛选条件
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("email LIKE ? OR username LIKE ? OR first_name LIKE ? OR last_name LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error

	return users, total, err
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&user.User{}).Where("id = ?", id).Update("status", status).Error
}

// GetStats 获取用户统计
func (r *UserRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总用户数
	var total int64
	if err := r.db.Model(&user.User{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&user.User{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	for _, stat := range statusStats {
		stats[stat.Status] = stat.Count
	}

	// 按角色统计
	var roleStats []struct {
		Role  string
		Count int64
	}
	if err := r.db.Model(&user.User{}).Select("role, COUNT(*) as count").Group("role").Scan(&roleStats).Error; err != nil {
		return nil, err
	}

	roleMap := make(map[string]int64)
	for _, stat := range roleStats {
		roleMap[stat.Role] = stat.Count
	}
	stats["by_role"] = roleMap

	return stats, nil
}

// Count 获取用户总数
func (r *UserRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&user.User{}).Count(&count).Error
	return count, err
}

// CountByDateRange 获取日期范围内的用户数
func (r *UserRepository) CountByDateRange(startDate, endDate time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&user.User{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&count).Error
	return count, err
}

// FindRecent 获取最近注册的用户
func (r *UserRepository) FindRecent(limit int) ([]user.User, error) {
	var users []user.User
	err := r.db.Order("created_at DESC").Limit(limit).Find(&users).Error
	return users, err
}
