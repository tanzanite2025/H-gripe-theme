package service

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"commerce-platform/internal/domain/currency"
	suppliercostdomain "commerce-platform/internal/domain/productsuppliercost"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrProductSupplierCostRecordNotFound  = errors.New("product supplier cost record not found")
	ErrProductSupplierCostRecordInvalid   = errors.New("product supplier cost record is invalid")
	ErrProductSupplierCostRecordSKUExists = errors.New("SKU already has a supplier cost record")
	MaxProductSupplierCostRecordCodeItems = 100
)

type ProductSupplierCostRecordService struct {
	repo              *repository.ProductSupplierCostRecordRepository
	profitabilityRepo *repository.ProductProfitCalculationRepository
	catalogRepo       *repository.ProductSupplierCostCatalogRepository
}

type ProductSupplierCostRecordListInput struct {
	Page     int
	PageSize int
	Search   string
	ExactSKU string
}

type ProductSupplierCostRecordDetailsInput struct {
	UnitCost                *float64
	Currency                string
	SupplierName            string
	SupplierContactName     string
	SupplierPhone           string
	SupplierEmail           string
	LeadTimeDays            int
	MinimumOrderQuantity    int
	InboundShippingUnitCost float64
	PackagingUnitCost       float64
	OtherUnitCost           float64
}

type ProductSupplierCostRecordCreateInput struct {
	SKU string
	ProductSupplierCostRecordDetailsInput
}

type ProductSupplierCostRecordUpdateInput struct {
	ProductSupplierCostRecordDetailsInput
}

type productSupplierCostRecordSnapshotInput struct {
	ProductCode string
	ProductName string
	ProductSupplierCostRecordDetailsInput
}

func NewProductSupplierCostRecordServiceWithProfitability(
	repo *repository.ProductSupplierCostRecordRepository,
	profitabilityRepo *repository.ProductProfitCalculationRepository,
) *ProductSupplierCostRecordService {
	return &ProductSupplierCostRecordService{
		repo:              repo,
		profitabilityRepo: profitabilityRepo,
	}
}

func (s *ProductSupplierCostRecordService) ConfigureCatalogRepository(repo *repository.ProductSupplierCostCatalogRepository) {
	if s == nil {
		return
	}
	s.catalogRepo = repo
}

func (s *ProductSupplierCostRecordService) ListProductOptions(input ProductSupplierCostRecordListInput) ([]suppliercostdomain.ProductOption, int64, error) {
	if s == nil || s.catalogRepo == nil {
		return nil, 0, errors.New("product supplier cost catalog is unavailable")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 50 {
		input.PageSize = 20
	}
	if len(strings.TrimSpace(input.Search)) > 120 || len(strings.TrimSpace(input.ExactSKU)) > 160 {
		return nil, 0, ErrProductSupplierCostRecordInvalid
	}

	return s.catalogRepo.ListOptions(repository.ProductSupplierCostCatalogFilter{
		Page:     input.Page,
		PageSize: input.PageSize,
		Search:   input.Search,
		SKU:      input.ExactSKU,
	})
}

func (s *ProductSupplierCostRecordService) ListAdmin(input ProductSupplierCostRecordListInput) ([]suppliercostdomain.ProductSupplierCostRecord, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, errors.New("product supplier cost record service is unavailable")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	return s.repo.List(input.Page, input.PageSize, repository.ProductSupplierCostRecordFilter{
		Search: input.Search,
	})
}

func (s *ProductSupplierCostRecordService) GetAdmin(id uint) (*suppliercostdomain.ProductSupplierCostRecord, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product supplier cost record service is unavailable")
	}
	record, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductSupplierCostRecordNotFound
	}
	return record, err
}

func (s *ProductSupplierCostRecordService) ListByProductCodes(codes []string) ([]suppliercostdomain.ProductSupplierCostRecord, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product supplier cost record service is unavailable")
	}
	normalized := normalizeProductCodes(codes)
	if len(normalized) > MaxProductSupplierCostRecordCodeItems {
		return nil, ErrProductSupplierCostRecordInvalid
	}
	return s.repo.FindByProductCodes(normalized)
}

func (s *ProductSupplierCostRecordService) Create(input ProductSupplierCostRecordCreateInput) (*suppliercostdomain.ProductSupplierCostRecord, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product supplier cost record service is unavailable")
	}

	sku := strings.TrimSpace(input.SKU)
	if sku == "" {
		return nil, fmt.Errorf("%w: sku is required", ErrProductSupplierCostRecordInvalid)
	}
	if len(sku) > 160 {
		return nil, fmt.Errorf("%w: sku is too long", ErrProductSupplierCostRecordInvalid)
	}
	productOption, err := s.findAvailableProductOption(sku)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeProductSupplierCostRecordFromInput(
		sku,
		productOption.ProductName,
		input.ProductSupplierCostRecordDetailsInput,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if _, err := repo.FindByProductCode(normalized.ProductCode); err == nil {
			return ErrProductSupplierCostRecordSKUExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := repo.Create(normalized); err != nil {
			return err
		}
		return s.syncProfitabilitySnapshotInTx(tx, normalized)
	}); err != nil {
		return nil, err
	}
	return s.GetAdmin(normalized.ID)
}

func (s *ProductSupplierCostRecordService) Update(id uint, input ProductSupplierCostRecordUpdateInput) (*suppliercostdomain.ProductSupplierCostRecord, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product supplier cost record service is unavailable")
	}
	existing, err := s.GetAdmin(id)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeProductSupplierCostRecordFromInput(
		existing.ProductCode,
		existing.ProductName,
		input.ProductSupplierCostRecordDetailsInput,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		normalized.ID = existing.ID
		if err := repo.Update(normalized); err != nil {
			return err
		}
		return s.syncProfitabilitySnapshotInTx(tx, normalized)
	}); err != nil {
		return nil, err
	}
	return s.GetAdmin(id)
}

func (s *ProductSupplierCostRecordService) Delete(id uint) error {
	if s == nil || s.repo == nil {
		return errors.New("product supplier cost record service is unavailable")
	}
	existing, err := s.GetAdmin(id)
	if err != nil {
		return err
	}
	return s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).Delete(id); err != nil {
			return err
		}
		if s.profitabilityRepo == nil {
			return nil
		}
		return s.profitabilityRepo.WithTx(tx).ReplaceCurrentSnapshotsInTx(
			tx,
			nil,
			[]string{existing.ProductCode},
		)
	})
}

func (s *ProductSupplierCostRecordService) findAvailableProductOption(sku string) (*suppliercostdomain.ProductOption, error) {
	if s.catalogRepo == nil {
		return nil, errors.New("product supplier cost catalog is unavailable")
	}

	option, err := s.catalogRepo.FindOptionBySKU(sku)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: sku does not exist", ErrProductSupplierCostRecordInvalid)
	}
	if err != nil {
		return nil, err
	}
	if !option.Available {
		return nil, fmt.Errorf("%w: sku is unavailable", ErrProductSupplierCostRecordInvalid)
	}
	return option, nil
}

func normalizeProductSupplierCostRecordFromInput(
	productCode string,
	productName string,
	input ProductSupplierCostRecordDetailsInput,
) (*suppliercostdomain.ProductSupplierCostRecord, error) {
	record := &suppliercostdomain.ProductSupplierCostRecord{
		ProductCode:             strings.TrimSpace(productCode),
		ProductName:             strings.TrimSpace(productName),
		Currency:                strings.TrimSpace(input.Currency),
		SupplierName:            strings.TrimSpace(input.SupplierName),
		SupplierContactName:     strings.TrimSpace(input.SupplierContactName),
		SupplierPhone:           strings.TrimSpace(input.SupplierPhone),
		SupplierEmail:           strings.TrimSpace(input.SupplierEmail),
		LeadTimeDays:            input.LeadTimeDays,
		MinimumOrderQuantity:    input.MinimumOrderQuantity,
		InboundShippingUnitCost: input.InboundShippingUnitCost,
		PackagingUnitCost:       input.PackagingUnitCost,
		OtherUnitCost:           input.OtherUnitCost,
	}
	if record.Currency == "" {
		record.Currency = suppliercostdomain.DefaultCurrency
	}
	record.Currency = currency.NormalizeCode(record.Currency)
	if !currency.IsCatalogCode(record.Currency) {
		return nil, fmt.Errorf("%w: unsupported currency", ErrProductSupplierCostRecordInvalid)
	}
	if record.ProductCode == "" {
		return nil, fmt.Errorf("%w: product_code is required", ErrProductSupplierCostRecordInvalid)
	}
	if record.ProductName == "" {
		return nil, fmt.Errorf("%w: product_name is required", ErrProductSupplierCostRecordInvalid)
	}
	if len(record.ProductCode) > 160 {
		return nil, fmt.Errorf("%w: product_code is too long", ErrProductSupplierCostRecordInvalid)
	}
	if len(record.ProductName) > 255 {
		return nil, fmt.Errorf("%w: product_name is too long", ErrProductSupplierCostRecordInvalid)
	}
	if input.UnitCost == nil {
		return nil, fmt.Errorf("%w: unit_cost is required", ErrProductSupplierCostRecordInvalid)
	}
	record.UnitCost = *input.UnitCost

	costs := []struct {
		name  string
		value float64
	}{
		{name: "unit_cost", value: record.UnitCost},
		{name: "inbound_shipping_unit_cost", value: record.InboundShippingUnitCost},
		{name: "packaging_unit_cost", value: record.PackagingUnitCost},
		{name: "other_unit_cost", value: record.OtherUnitCost},
	}
	for _, cost := range costs {
		if math.IsNaN(cost.value) || math.IsInf(cost.value, 0) || cost.value < 0 {
			return nil, fmt.Errorf("%w: %s must be a finite non-negative amount", ErrProductSupplierCostRecordInvalid, cost.name)
		}
	}
	if record.LeadTimeDays < 0 || record.LeadTimeDays > 3650 {
		return nil, fmt.Errorf("%w: lead_time_days must be between 0 and 3650", ErrProductSupplierCostRecordInvalid)
	}
	if record.MinimumOrderQuantity == 0 {
		record.MinimumOrderQuantity = 1
	}
	if record.MinimumOrderQuantity < 1 || record.MinimumOrderQuantity > 1000000000 {
		return nil, fmt.Errorf("%w: minimum_order_quantity must be between 1 and 1000000000", ErrProductSupplierCostRecordInvalid)
	}
	if len(record.SupplierEmail) > 190 {
		return nil, fmt.Errorf("%w: supplier_email is too long", ErrProductSupplierCostRecordInvalid)
	}
	return record, nil
}

func normalizeProductSupplierCostRecordSnapshotInput(input productSupplierCostRecordSnapshotInput) (*suppliercostdomain.ProductSupplierCostRecord, error) {
	return normalizeProductSupplierCostRecordFromInput(
		input.ProductCode,
		input.ProductName,
		input.ProductSupplierCostRecordDetailsInput,
	)
}

func (s *ProductSupplierCostRecordService) syncProfitabilitySnapshotInTx(tx *gorm.DB, record *suppliercostdomain.ProductSupplierCostRecord) error {
	if s.profitabilityRepo == nil {
		return nil
	}

	snapshot, err := s.profitabilityRepo.WithTx(tx).FindByProductCode(record.ProductCode)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	unitCost := record.UnitCost
	result, err := suppliercostdomain.CalculateProfit(suppliercostdomain.ProfitCalculationInput{
		ProductCode:             record.ProductCode,
		ProductName:             record.ProductName,
		SellingCurrency:         snapshot.Currency,
		CostCurrency:            record.Currency,
		ListPrice:               snapshot.ListPrice,
		SalePrice:               snapshot.SalePrice,
		UnitCost:                &unitCost,
		InboundShippingUnitCost: record.InboundShippingUnitCost,
		PackagingUnitCost:       record.PackagingUnitCost,
		OtherUnitCost:           record.OtherUnitCost,
	})
	if err != nil {
		return err
	}

	switch result.Status {
	case suppliercostdomain.ProfitStatusReady, suppliercostdomain.ProfitStatusWarning:
		return s.profitabilityRepo.WithTx(tx).ReplaceCurrentSnapshotsInTx(
			tx,
			[]suppliercostdomain.ProductProfitCalculation{profitCalculationRecord(result)},
			nil,
		)
	default:
		return s.profitabilityRepo.WithTx(tx).ReplaceCurrentSnapshotsInTx(
			tx,
			nil,
			[]string{record.ProductCode},
		)
	}
}

func normalizeProductCodes(codes []string) []string {
	normalized := make([]string, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		normalized = append(normalized, code)
	}
	return normalized
}
