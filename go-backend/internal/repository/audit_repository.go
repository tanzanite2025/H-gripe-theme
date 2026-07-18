package repository

import (
	"tanzanite/internal/domain/audit"
	"time"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// CreateAuditLog 创建审计日志
func (r *AuditRepository) CreateAuditLog(log *audit.AuditLog) error {
	return r.db.Create(log).Error
}

// FindAuditLogByID 根据ID查找审计日志
func (r *AuditRepository) FindAuditLogByID(id uint) (*audit.AuditLog, error) {
	var log audit.AuditLog
	err := r.db.First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindAuditLogsByUserID 查找用户的审计日志
func (r *AuditRepository) FindAuditLogsByUserID(userID uint, page, pageSize int) ([]audit.AuditLog, int64, error) {
	var logs []audit.AuditLog
	var total int64

	query := r.db.Model(&audit.AuditLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

// FindAuditLogsByEntity 查找实体的审计日志
func (r *AuditRepository) FindAuditLogsByEntity(entityType string, entityID uint, page, pageSize int) ([]audit.AuditLog, int64, error) {
	var logs []audit.AuditLog
	var total int64

	query := r.db.Model(&audit.AuditLog{}).
		Where("resource = ? AND resource_id = ?", entityType, entityID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

// FindAllAuditLogs 查找所有审计日志（管理员）
func (r *AuditRepository) FindAllAuditLogs(page, pageSize int, action, resource string) ([]audit.AuditLog, int64, error) {
	var logs []audit.AuditLog
	var total int64

	query := r.db.Model(&audit.AuditLog{})

	if action != "" {
		query = query.Where("action = ?", action)
	}

	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

// FindAuditLogsByDateRange 根据日期范围查找审计日志
func (r *AuditRepository) FindAuditLogsByDateRange(startDate, endDate time.Time, page, pageSize int) ([]audit.AuditLog, int64, error) {
	var logs []audit.AuditLog
	var total int64

	query := r.db.Model(&audit.AuditLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

// FindAuditLogsByIP 根据IP地址查找审计日志
func (r *AuditRepository) FindAuditLogsByIP(ipAddress string, page, pageSize int) ([]audit.AuditLog, int64, error) {
	var logs []audit.AuditLog
	var total int64

	query := r.db.Model(&audit.AuditLog{}).Where("ip_address = ?", ipAddress)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

// SearchAuditLogs 搜索审计日志
func (r *AuditRepository) SearchAuditLogs(keyword string, page, pageSize int) ([]audit.AuditLog, int64, error) {
	var logs []audit.AuditLog
	var total int64

	query := r.db.Model(&audit.AuditLog{}).
		Where(
			"username ILIKE ? OR action ILIKE ? OR resource ILIKE ? OR path ILIKE ? OR user_agent ILIKE ? OR error_message ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%",
		)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

// DeleteOldAuditLogs 删除旧的审计日志
func (r *AuditRepository) DeleteOldAuditLogs(beforeDate time.Time) error {
	return r.db.Where("created_at < ?", beforeDate).Delete(&audit.AuditLog{}).Error
}

// GetAuditStats 获取审计统计
func (r *AuditRepository) GetAuditStats(startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := r.db.Model(&audit.AuditLog{})
	if !startDate.IsZero() && !endDate.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	}

	// 总日志数
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	var todayCount int64
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if err := query.Where("created_at >= ?", startOfToday).Count(&todayCount).Error; err != nil {
		return nil, err
	}
	stats["today_count"] = todayCount

	var successCount int64
	if err := query.Where("status = ?", "success").Count(&successCount).Error; err != nil {
		return nil, err
	}
	stats["success_count"] = successCount

	var failedCount int64
	if err := query.Where("status = ?", "failed").Count(&failedCount).Error; err != nil {
		return nil, err
	}
	stats["failed_count"] = failedCount

	// 按操作类型统计
	var actionStats []struct {
		Action string
		Count  int64
	}
	if err := query.Select("action, count(*) as count").
		Group("action").Scan(&actionStats).Error; err != nil {
		return nil, err
	}
	stats["action_stats"] = actionStats

	// 按实体类型统计
	var entityStats []struct {
		EntityType string
		Count      int64
	}
	if err := query.Select("resource as entity_type, count(*) as count").
		Group("resource").Scan(&entityStats).Error; err != nil {
		return nil, err
	}
	stats["entity_stats"] = entityStats

	// 按用户统计（Top 10）
	var userStats []struct {
		UserID uint
		Count  int64
	}
	if err := query.Select("user_id, count(*) as count").
		Group("user_id").Order("count DESC").Limit(10).
		Scan(&userStats).Error; err != nil {
		return nil, err
	}
	stats["top_users"] = userStats

	return stats, nil
}

// GetRecentActivities 获取最近活动
func (r *AuditRepository) GetRecentActivities(limit int) ([]audit.AuditLog, error) {
	var logs []audit.AuditLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
