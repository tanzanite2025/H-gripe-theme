package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

var ErrShipmentRecordUnavailable = errors.New("shipment record service is unavailable")

type ShipmentRecordService struct {
	repo *repository.ShipmentRecordRepository
}

type ShipmentRecordListInput struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
}

type ShipmentRecordUpdateInput struct {
	ShippingNote   string
	ShippingImages []string
	ProductCodes   []string
	WarrantyMonths int
	WarrantyStart  time.Time
}

func NewShipmentRecordService(repo *repository.ShipmentRecordRepository) *ShipmentRecordService {
	return &ShipmentRecordService{repo: repo}
}

func (s *ShipmentRecordService) ListAdmin(input ShipmentRecordListInput) ([]shippingdomain.ShipmentRecord, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, ErrShipmentRecordUnavailable
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}
	return s.repo.FindAll(input.Page, input.PageSize, repository.ShipmentRecordFilter{
		Keyword: input.Keyword,
		Status:  input.Status,
	})
}

func (s *ShipmentRecordService) GetAdmin(id uint) (*shippingdomain.ShipmentRecord, error) {
	if s == nil || s.repo == nil {
		return nil, ErrShipmentRecordUnavailable
	}
	return s.repo.FindByOrderID(id)
}

func (s *ShipmentRecordService) GetStats() (map[string]interface{}, error) {
	if s == nil || s.repo == nil {
		return nil, ErrShipmentRecordUnavailable
	}
	return s.repo.GetStats()
}

func (s *ShipmentRecordService) GetForUser(orderNumber string, userID uint) (*shippingdomain.ShipmentRecord, error) {
	if s == nil || s.repo == nil {
		return nil, ErrShipmentRecordUnavailable
	}
	return s.repo.FindByOrderNumberForUser(orderNumber, userID)
}

func (s *ShipmentRecordService) Update(id uint, input ShipmentRecordUpdateInput) (*shippingdomain.ShipmentRecord, error) {
	if s == nil || s.repo == nil {
		return nil, ErrShipmentRecordUnavailable
	}
	record, err := s.repo.FindByOrderID(id)
	if err != nil {
		return nil, err
	}

	if input.WarrantyMonths <= 0 {
		input.WarrantyMonths = record.WarrantyMonths
	}
	if input.WarrantyMonths <= 0 {
		input.WarrantyMonths = 12
	}
	if input.WarrantyMonths > 120 {
		return nil, errors.New("warranty months must be between 1 and 120")
	}
	if input.WarrantyStart.IsZero() {
		input.WarrantyStart = record.WarrantyStartAt
	}
	if input.WarrantyStart.IsZero() {
		input.WarrantyStart = record.ShippedAt
	}

	images, err := normalizeShipmentImages(input.ShippingImages)
	if err != nil {
		return nil, err
	}
	if input.ShippingImages == nil {
		images, err = decodeShipmentStringList(record.ShippingImages)
		if err != nil {
			return nil, err
		}
	}

	productCodes, err := normalizeShipmentProductCodes(input.ProductCodes)
	if err != nil {
		return nil, err
	}
	if input.ProductCodes == nil {
		productCodes, err = decodeShipmentStringList(record.ProductCodes)
		if err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.UpsertDetailsForOrder(
		id,
		input.ShippingNote,
		images,
		productCodes,
		input.WarrantyMonths,
		input.WarrantyStart,
		input.WarrantyStart.AddDate(0, input.WarrantyMonths, 0),
	)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *ShipmentRecordService) AddImages(id uint, imageURLs []string) (*shippingdomain.ShipmentRecord, error) {
	if s == nil || s.repo == nil {
		return nil, ErrShipmentRecordUnavailable
	}
	record, err := s.repo.FindByOrderID(id)
	if err != nil {
		return nil, err
	}

	current, err := decodeShipmentStringList(record.ShippingImages)
	if err != nil {
		return nil, err
	}
	additional, err := normalizeShipmentImages(imageURLs)
	if err != nil {
		return nil, err
	}
	if len(current)+len(additional) > 20 {
		return nil, errors.New("shipping images cannot exceed 20 files")
	}

	seen := make(map[string]struct{}, len(current)+len(additional))
	merged := make([]string, 0, len(current)+len(additional))
	for _, image := range append(current, additional...) {
		if _, ok := seen[image]; ok {
			continue
		}
		seen[image] = struct{}{}
		merged = append(merged, image)
	}
	productCodes, err := decodeShipmentStringList(record.ProductCodes)
	if err != nil {
		return nil, err
	}

	updated, err := s.Update(id, ShipmentRecordUpdateInput{
		ShippingNote:   record.ShippingNote,
		ShippingImages: merged,
		ProductCodes:   productCodes,
		WarrantyMonths: record.WarrantyMonths,
		WarrantyStart:  record.WarrantyStartAt,
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func normalizeShipmentImages(images []string) ([]string, error) {
	result := make([]string, 0, len(images))
	seen := make(map[string]struct{}, len(images))
	for _, image := range images {
		image = strings.TrimSpace(image)
		if image == "" {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(image), "http://") && !strings.HasPrefix(strings.ToLower(image), "https://") {
			return nil, errors.New("shipping image must be an http or https URL")
		}
		if _, ok := seen[image]; ok {
			continue
		}
		seen[image] = struct{}{}
		result = append(result, image)
	}
	if len(result) > 20 {
		return nil, errors.New("shipping images cannot exceed 20 files")
	}
	return result, nil
}

func normalizeShipmentProductCodes(codes []string) ([]string, error) {
	result := make([]string, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if len(code) > 160 {
			return nil, errors.New("product code cannot exceed 160 characters")
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	if len(result) > 100 {
		return nil, errors.New("product codes cannot exceed 100 entries")
	}
	return result, nil
}

func decodeShipmentStringList(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}
	var images []string
	if err := json.Unmarshal(raw, &images); err != nil {
		return nil, err
	}
	return images, nil
}

func ShipmentRecordWarrantyStatus(record *shippingdomain.ShipmentRecord, now time.Time) (string, int) {
	if record == nil {
		return "expired", 0
	}
	if now.IsZero() {
		now = time.Now()
	}
	if record.Status == "cancelled" || !record.WarrantyExpires.After(now) {
		days := int(now.Sub(record.WarrantyExpires).Hours() / 24)
		if days < 0 {
			days = 0
		}
		return "expired", days
	}
	return "valid", int(record.WarrantyExpires.Sub(now).Hours() / 24)
}
