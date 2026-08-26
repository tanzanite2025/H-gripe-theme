package service

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"commerce-platform/internal/domain/currency"
	procurementdomain "commerce-platform/internal/domain/procurement"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrProductProcurementNotFound  = errors.New("product procurement record not found")
	ErrProductProcurementInvalid   = errors.New("product procurement record is invalid")
	ErrProductProcurementSKUExists = errors.New("SKU already has a procurement record")
	MaxProductProcurementCodeItems = 100
)

type ProductProcurementService struct {
	repo              *repository.ProductProcurementRepository
	profitabilityRepo *repository.ProductProfitCalculationRepository
	catalogRepo       *repository.ProductProcurementCatalogRepository
}

type ProductProcurementListInput struct {
	Page     int
	PageSize int
	Search   string
	ExactSKU string
}

type ProductProcurementDetailsInput struct {
	PurchasePrice           *float64
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

type ProductProcurementCreateInput struct {
	SKU string
	ProductProcurementDetailsInput
}

type ProductProcurementUpdateInput struct {
	ProductProcurementDetailsInput
}

type productProcurementSnapshotInput struct {
	ProductCode string
	ProductName string
	ProductProcurementDetailsInput
}

func NewProductProcurementServiceWithProfitability(
	repo *repository.ProductProcurementRepository,
	profitabilityRepo *repository.ProductProfitCalculationRepository,
) *ProductProcurementService {
	return &ProductProcurementService{
		repo:              repo,
		profitabilityRepo: profitabilityRepo,
	}
}

func (s *ProductProcurementService) ConfigureCatalogRepository(repo *repository.ProductProcurementCatalogRepository) {
	if s == nil {
		return
	}
	s.catalogRepo = repo
}

func (s *ProductProcurementService) ListProductOptions(input ProductProcurementListInput) ([]procurementdomain.ProductOption, int64, error) {
	if s == nil || s.catalogRepo == nil {
		return nil, 0, errors.New("product procurement catalog is unavailable")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 50 {
		input.PageSize = 20
	}
	if len(strings.TrimSpace(input.Search)) > 120 || len(strings.TrimSpace(input.ExactSKU)) > 160 {
		return nil, 0, ErrProductProcurementInvalid
	}

	return s.catalogRepo.ListOptions(repository.ProductProcurementCatalogFilter{
		Page:     input.Page,
		PageSize: input.PageSize,
		Search:   input.Search,
		SKU:      input.ExactSKU,
	})
}

func (s *ProductProcurementService) ListAdmin(input ProductProcurementListInput) ([]procurementdomain.ProductProcurement, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, errors.New("product procurement service is unavailable")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	return s.repo.List(input.Page, input.PageSize, repository.ProductProcurementFilter{
		Search: input.Search,
	})
}

func (s *ProductProcurementService) GetAdmin(id uint) (*procurementdomain.ProductProcurement, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product procurement service is unavailable")
	}
	record, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductProcurementNotFound
	}
	return record, err
}

func (s *ProductProcurementService) ListByProductCodes(codes []string) ([]procurementdomain.ProductProcurement, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product procurement service is unavailable")
	}
	normalized := normalizeProductCodes(codes)
	if len(normalized) > MaxProductProcurementCodeItems {
		return nil, ErrProductProcurementInvalid
	}
	return s.repo.FindByProductCodes(normalized)
}

func (s *ProductProcurementService) Create(input ProductProcurementCreateInput) (*procurementdomain.ProductProcurement, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product procurement service is unavailable")
	}

	sku := strings.TrimSpace(input.SKU)
	if sku == "" {
		return nil, fmt.Errorf("%w: sku is required", ErrProductProcurementInvalid)
	}
	if len(sku) > 160 {
		return nil, fmt.Errorf("%w: sku is too long", ErrProductProcurementInvalid)
	}
	productOption, err := s.findAvailableProductOption(sku)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeProductProcurementRecord(
		sku,
		productOption.ProductName,
		input.ProductProcurementDetailsInput,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if _, err := repo.FindByProductCode(normalized.ProductCode); err == nil {
			return ErrProductProcurementSKUExists
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

func (s *ProductProcurementService) Update(id uint, input ProductProcurementUpdateInput) (*procurementdomain.ProductProcurement, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product procurement service is unavailable")
	}
	existing, err := s.GetAdmin(id)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeProductProcurementRecord(
		existing.ProductCode,
		existing.ProductName,
		input.ProductProcurementDetailsInput,
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

func (s *ProductProcurementService) Delete(id uint) error {
	if s == nil || s.repo == nil {
		return errors.New("product procurement service is unavailable")
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

func (s *ProductProcurementService) findAvailableProductOption(sku string) (*procurementdomain.ProductOption, error) {
	if s.catalogRepo == nil {
		return nil, errors.New("product procurement catalog is unavailable")
	}

	option, err := s.catalogRepo.FindOptionBySKU(sku)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: sku does not exist", ErrProductProcurementInvalid)
	}
	if err != nil {
		return nil, err
	}
	if !option.Available {
		return nil, fmt.Errorf("%w: sku is unavailable", ErrProductProcurementInvalid)
	}
	return option, nil
}

func normalizeProductProcurementRecord(
	productCode string,
	productName string,
	input ProductProcurementDetailsInput,
) (*procurementdomain.ProductProcurement, error) {
	record := &procurementdomain.ProductProcurement{
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
		record.Currency = procurementdomain.DefaultCurrency
	}
	record.Currency = currency.NormalizeCode(record.Currency)
	if !currency.IsCatalogCode(record.Currency) {
		return nil, fmt.Errorf("%w: unsupported currency", ErrProductProcurementInvalid)
	}
	if record.ProductCode == "" {
		return nil, fmt.Errorf("%w: product_code is required", ErrProductProcurementInvalid)
	}
	if record.ProductName == "" {
		return nil, fmt.Errorf("%w: product_name is required", ErrProductProcurementInvalid)
	}
	if len(record.ProductCode) > 160 {
		return nil, fmt.Errorf("%w: product_code is too long", ErrProductProcurementInvalid)
	}
	if len(record.ProductName) > 255 {
		return nil, fmt.Errorf("%w: product_name is too long", ErrProductProcurementInvalid)
	}
	if input.PurchasePrice == nil {
		return nil, fmt.Errorf("%w: purchase_price is required", ErrProductProcurementInvalid)
	}
	record.PurchasePrice = *input.PurchasePrice

	costs := []struct {
		name  string
		value float64
	}{
		{name: "purchase_price", value: record.PurchasePrice},
		{name: "inbound_shipping_unit_cost", value: record.InboundShippingUnitCost},
		{name: "packaging_unit_cost", value: record.PackagingUnitCost},
		{name: "other_unit_cost", value: record.OtherUnitCost},
	}
	for _, cost := range costs {
		if math.IsNaN(cost.value) || math.IsInf(cost.value, 0) || cost.value < 0 {
			return nil, fmt.Errorf("%w: %s must be a finite non-negative amount", ErrProductProcurementInvalid, cost.name)
		}
	}
	if record.LeadTimeDays < 0 || record.LeadTimeDays > 3650 {
		return nil, fmt.Errorf("%w: lead_time_days must be between 0 and 3650", ErrProductProcurementInvalid)
	}
	if record.MinimumOrderQuantity == 0 {
		record.MinimumOrderQuantity = 1
	}
	if record.MinimumOrderQuantity < 1 || record.MinimumOrderQuantity > 1000000000 {
		return nil, fmt.Errorf("%w: minimum_order_quantity must be between 1 and 1000000000", ErrProductProcurementInvalid)
	}
	if len(record.SupplierEmail) > 190 {
		return nil, fmt.Errorf("%w: supplier_email is too long", ErrProductProcurementInvalid)
	}
	return record, nil
}

func normalizeProductProcurementSnapshotInput(input productProcurementSnapshotInput) (*procurementdomain.ProductProcurement, error) {
	return normalizeProductProcurementRecord(
		input.ProductCode,
		input.ProductName,
		input.ProductProcurementDetailsInput,
	)
}

func (s *ProductProcurementService) syncProfitabilitySnapshotInTx(tx *gorm.DB, record *procurementdomain.ProductProcurement) error {
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

	purchasePrice := record.PurchasePrice
	result, err := procurementdomain.CalculateProfit(procurementdomain.ProfitCalculationInput{
		ProductCode:             record.ProductCode,
		ProductName:             record.ProductName,
		SellingCurrency:         snapshot.Currency,
		CostCurrency:            record.Currency,
		ListPrice:               snapshot.ListPrice,
		SalePrice:               snapshot.SalePrice,
		PurchasePrice:           &purchasePrice,
		InboundShippingUnitCost: record.InboundShippingUnitCost,
		PackagingUnitCost:       record.PackagingUnitCost,
		OtherUnitCost:           record.OtherUnitCost,
	})
	if err != nil {
		return err
	}

	switch result.Status {
	case procurementdomain.ProfitStatusReady, procurementdomain.ProfitStatusWarning:
		return s.profitabilityRepo.WithTx(tx).ReplaceCurrentSnapshotsInTx(
			tx,
			[]procurementdomain.ProductProfitCalculation{profitCalculationRecord(result)},
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
