package service

import (
	"errors"
	"strings"
	"tanzanite/internal/domain/audit"
	"tanzanite/internal/repository"
	"time"
)

var (
	ErrInvalidAuditDate     = errors.New("invalid audit date")
	ErrAuditKeywordRequired = errors.New("audit search keyword is required")
)

type AuditService struct {
	auditRepo *repository.AuditRepository
}

type AuditListInput struct {
	Page      int
	PageSize  int
	Action    string
	Resource  string
	UserID    uint
	IPAddress string
	StartDate string
	EndDate   string
}

func NewAuditService(auditRepo *repository.AuditRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

func (s *AuditService) ListAuditLogs(input AuditListInput) ([]audit.AuditLog, int64, error) {
	if input.UserID > 0 {
		return s.auditRepo.FindAuditLogsByUserID(input.UserID, input.Page, input.PageSize)
	}
	if input.IPAddress != "" {
		return s.auditRepo.FindAuditLogsByIP(input.IPAddress, input.Page, input.PageSize)
	}
	if input.StartDate != "" && input.EndDate != "" {
		start, end, err := parseAuditDateRange(input.StartDate, input.EndDate)
		if err != nil {
			return nil, 0, err
		}
		return s.auditRepo.FindAuditLogsByDateRange(start, end, input.Page, input.PageSize)
	}
	return s.auditRepo.FindAllAuditLogs(input.Page, input.PageSize, input.Action, input.Resource)
}

func (s *AuditService) GetAuditLog(id uint) (*audit.AuditLog, error) {
	return s.auditRepo.FindAuditLogByID(id)
}

func (s *AuditService) GetAuditStats(startDate, endDate string) (map[string]interface{}, error) {
	start, err := parseOptionalAuditDate(startDate)
	if err != nil {
		return nil, err
	}
	end, err := parseOptionalAuditDate(endDate)
	if err != nil {
		return nil, err
	}
	return s.auditRepo.GetAuditStats(start, end)
}

func (s *AuditService) GetRecentActivities(limit int) ([]audit.AuditLog, error) {
	return s.auditRepo.GetRecentActivities(limit)
}

func (s *AuditService) SearchAuditLogs(keyword string, page, pageSize int) ([]audit.AuditLog, int64, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, 0, ErrAuditKeywordRequired
	}
	return s.auditRepo.SearchAuditLogs(keyword, page, pageSize)
}

func (s *AuditService) GetUserAuditLogs(userID uint, page, pageSize int) ([]audit.AuditLog, int64, error) {
	return s.auditRepo.FindAuditLogsByUserID(userID, page, pageSize)
}

func (s *AuditService) DeleteOldLogs(beforeDate string) error {
	parsedDate, err := parseAuditDate(beforeDate)
	if err != nil {
		return err
	}
	return s.auditRepo.DeleteOldAuditLogs(parsedDate)
}

func parseAuditDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := parseAuditDate(startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseAuditDate(endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func parseOptionalAuditDate(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, nil
	}
	return parseAuditDate(value)
}

func parseAuditDate(value string) (time.Time, error) {
	parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, ErrInvalidAuditDate
	}
	return parsedDate, nil
}
