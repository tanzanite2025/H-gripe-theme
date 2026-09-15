package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	suppliercostdomain "commerce-platform/internal/domain/productsuppliercost"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	MaxProfitabilityBatchItems = 100
	MaxProfitabilityCodeItems  = 100
)

var (
	ErrProductProfitabilityInvalid                                 = errors.New("product profitability input is invalid")
	ErrProductProfitabilityBatchLarge                              = errors.New("product profitability batch is too large")
	ErrProductProfitabilitySupplierCostRecordRepositoryUnavailable = errors.New("product supplier cost record repository is unavailable")
)

type ProductProfitabilityService struct {
	repo                   *repository.ProductProfitCalculationRepository
	supplierCostRecordRepo *repository.ProductSupplierCostRecordRepository
}

type ProfitabilitySupplierCostDetailsInput struct {
	SupplierName         string
	SupplierContactName  string
	SupplierPhone        string
	SupplierEmail        string
	LeadTimeDays         int
	MinimumOrderQuantity int
}

type ProfitabilityItemInput struct {
	ProductCode string
	ProductName string

	SellingCurrency string
	CostCurrency    string

	ListPrice float64
	SalePrice *float64
	UnitCost  *float64
	// UnitCostKnown distinguishes an explicit zero cost from an omitted
	// cost. Unknown cost values are never persisted as a ready snapshot.
	UnitCostKnown bool

	InboundShippingUnitCost float64
	PackagingUnitCost       float64
	OtherUnitCost           float64

	SupplierCostDetails *ProfitabilitySupplierCostDetailsInput
}

type ProfitabilitySkippedItem struct {
	ProductCode string `json:"product_code"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}

type ProfitabilityBatchResult struct {
	Records []suppliercostdomain.ProductProfitCalculation `json:"records"`
	Skipped []ProfitabilitySkippedItem                    `json:"skipped"`
}

type ProfitabilityItemIssue struct {
	Index       int    `json:"index"`
	ProductCode string `json:"product_code"`
	Status      string `json:"status,omitempty"`
	Reason      string `json:"reason"`
}

type ProfitabilityBatchValidationError struct {
	Items []ProfitabilityItemIssue `json:"items"`
}

func (e *ProfitabilityBatchValidationError) Error() string {
	if e == nil || len(e.Items) == 0 {
		return ErrProductProfitabilityInvalid.Error()
	}
	return fmt.Sprintf("%s: %d item(s) invalid", ErrProductProfitabilityInvalid, len(e.Items))
}

func (e *ProfitabilityBatchValidationError) Unwrap() error {
	return ErrProductProfitabilityInvalid
}

func NewProductProfitabilityService(repo *repository.ProductProfitCalculationRepository) *ProductProfitabilityService {
	return &ProductProfitabilityService{repo: repo}
}

func NewProductProfitabilityServiceWithSupplierCostRecords(
	repo *repository.ProductProfitCalculationRepository,
	supplierCostRecordRepo *repository.ProductSupplierCostRecordRepository,
) *ProductProfitabilityService {
	return &ProductProfitabilityService{
		repo:                   repo,
		supplierCostRecordRepo: supplierCostRecordRepo,
	}
}

func (s *ProductProfitabilityService) Preview(items []ProfitabilityItemInput) ([]suppliercostdomain.ProfitCalculationResult, error) {
	if len(items) > MaxProfitabilityBatchItems {
		return nil, ErrProductProfitabilityBatchLarge
	}

	results := make([]suppliercostdomain.ProfitCalculationResult, 0, len(items))
	seenCodes := make(map[string]int, len(items))
	for index, item := range items {
		normalized, err := normalizeProfitabilityItem(item)
		if err != nil {
			return nil, fmt.Errorf("%w: item %d: %v", ErrProductProfitabilityInvalid, index+1, err)
		}
		if previousIndex, exists := seenCodes[normalized.ProductCode]; exists {
			return nil, fmt.Errorf(
				"%w: item %d: product_code duplicates item %d",
				ErrProductProfitabilityInvalid,
				index+1,
				previousIndex+1,
			)
		}
		seenCodes[normalized.ProductCode] = index
		result, err := calculateProfitabilityItem(normalized)
		if err != nil {
			return nil, fmt.Errorf("%w: item %d: %v", ErrProductProfitabilityInvalid, index+1, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *ProductProfitabilityService) ListByCodes(codes []string) ([]suppliercostdomain.ProductProfitCalculation, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("product profitability service is unavailable")
	}
	normalized := normalizeCodesForProfitability(codes)
	if len(normalized) > MaxProfitabilityCodeItems {
		return nil, ErrProductProfitabilityBatchLarge
	}
	return s.repo.FindByProductCodes(normalized)
}

func (s *ProductProfitabilityService) BulkUpsert(items []ProfitabilityItemInput) (ProfitabilityBatchResult, error) {
	if s == nil || s.repo == nil {
		return ProfitabilityBatchResult{}, errors.New("product profitability service is unavailable")
	}
	if len(items) > MaxProfitabilityBatchItems {
		return ProfitabilityBatchResult{}, ErrProductProfitabilityBatchLarge
	}

	records := make([]suppliercostdomain.ProductProfitCalculation, 0, len(items))
	supplierCostRecords := make([]suppliercostdomain.ProductSupplierCostRecord, 0, len(items))
	skipped := make([]ProfitabilitySkippedItem, 0)
	clearCodes := make([]string, 0)
	issues := make([]ProfitabilityItemIssue, 0)
	seenCodes := make(map[string]int, len(items))
	for index, item := range items {
		normalized, err := normalizeProfitabilityItem(item)
		if err != nil {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: strings.TrimSpace(item.ProductCode),
				Reason:      err.Error(),
			})
			continue
		}
		if previousIndex, exists := seenCodes[normalized.ProductCode]; exists {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Reason:      fmt.Sprintf("product_code duplicates item %d", previousIndex+1),
			})
			continue
		}
		seenCodes[normalized.ProductCode] = index

		result, err := calculateProfitabilityItem(normalized)
		if err != nil {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Reason:      err.Error(),
			})
			continue
		}
		if normalized.SupplierCostDetails != nil && result.Status == suppliercostdomain.ProfitStatusMissingUnitCost {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Status:      result.Status,
				Reason:      "unit_cost is required when supplier cost details are supplied",
			})
			continue
		}
		if result.Status == suppliercostdomain.ProfitStatusMissingUnitCost {
			skipped = append(skipped, ProfitabilitySkippedItem{
				ProductCode: normalized.ProductCode,
				Status:      result.Status,
				Reason:      "unit cost is not known",
			})
			clearCodes = append(clearCodes, normalized.ProductCode)
			continue
		}
		if result.Status == suppliercostdomain.ProfitStatusCurrencyMismatch ||
			result.Status == suppliercostdomain.ProfitStatusInvalidSelling ||
			result.Status == suppliercostdomain.ProfitStatusInvalidCost {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Status:      result.Status,
				Reason:      profitabilityStatusReason(result.Status),
			})
			continue
		}
		if result.UnitCost == nil {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Reason:      "unit_cost is required",
			})
			continue
		}

		costCurrency := normalized.CostCurrency
		if strings.TrimSpace(costCurrency) == "" {
			costCurrency = normalized.SellingCurrency
		}
		supplierCostRecordInput := productSupplierCostRecordSnapshotInput{
			ProductCode: normalized.ProductCode,
			ProductName: normalized.ProductName,
			ProductSupplierCostRecordDetailsInput: ProductSupplierCostRecordDetailsInput{
				UnitCost: result.UnitCost,
				Currency: costCurrency,
			},
		}
		if normalized.SupplierCostDetails != nil {
			supplierCostRecordInput.SupplierName = normalized.SupplierCostDetails.SupplierName
			supplierCostRecordInput.SupplierContactName = normalized.SupplierCostDetails.SupplierContactName
			supplierCostRecordInput.SupplierPhone = normalized.SupplierCostDetails.SupplierPhone
			supplierCostRecordInput.SupplierEmail = normalized.SupplierCostDetails.SupplierEmail
			supplierCostRecordInput.LeadTimeDays = normalized.SupplierCostDetails.LeadTimeDays
			supplierCostRecordInput.MinimumOrderQuantity = normalized.SupplierCostDetails.MinimumOrderQuantity
		}
		supplierCostRecordInput.InboundShippingUnitCost = normalized.InboundShippingUnitCost
		supplierCostRecordInput.PackagingUnitCost = normalized.PackagingUnitCost
		supplierCostRecordInput.OtherUnitCost = normalized.OtherUnitCost
		supplierCostRecord, supplierCostRecordErr := normalizeProductSupplierCostRecordSnapshotInput(supplierCostRecordInput)
		if supplierCostRecordErr != nil {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Reason:      supplierCostRecordErr.Error(),
			})
			continue
		}
		supplierCostRecords = append(supplierCostRecords, *supplierCostRecord)
		records = append(records, profitCalculationRecord(result))
	}

	if len(issues) > 0 {
		return ProfitabilityBatchResult{Skipped: skipped}, &ProfitabilityBatchValidationError{Items: issues}
	}
	if (len(supplierCostRecords) > 0 || len(clearCodes) > 0) && s.supplierCostRecordRepo == nil {
		return ProfitabilityBatchResult{Skipped: skipped}, ErrProductProfitabilitySupplierCostRecordRepositoryUnavailable
	}
	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		if len(supplierCostRecords) > 0 {
			if err := s.supplierCostRecordRepo.UpsertInTx(tx, supplierCostRecords); err != nil {
				return err
			}
		}
		if len(clearCodes) > 0 {
			if err := s.supplierCostRecordRepo.DeleteByProductCodesInTx(tx, clearCodes); err != nil {
				return err
			}
		}
		return s.repo.WithTx(tx).ReplaceCurrentSnapshotsInTx(tx, records, clearCodes)
	}); err != nil {
		return ProfitabilityBatchResult{Skipped: skipped}, err
	}
	if len(records) > 0 {
		saved, err := s.repo.FindByProductCodes(profitabilityRecordCodes(records))
		if err != nil {
			return ProfitabilityBatchResult{Skipped: skipped}, err
		}
		return ProfitabilityBatchResult{
			Records: saved,
			Skipped: skipped,
		}, nil
	}
	return ProfitabilityBatchResult{Records: []suppliercostdomain.ProductProfitCalculation{}, Skipped: skipped}, nil
}

func normalizeProfitabilityItem(input ProfitabilityItemInput) (ProfitabilityItemInput, error) {
	input.ProductCode = strings.TrimSpace(input.ProductCode)
	input.ProductName = strings.TrimSpace(input.ProductName)
	if input.ProductCode == "" {
		return input, errors.New("product_code is required")
	}
	if len(input.ProductCode) > 160 {
		return input, errors.New("product_code is too long")
	}
	if input.ProductName == "" {
		return input, errors.New("product_name is required")
	}
	if len(input.ProductName) > 255 {
		return input, errors.New("product_name is too long")
	}
	if !input.UnitCostKnown {
		input.UnitCost = nil
	}
	if input.UnitCostKnown && input.UnitCost == nil {
		return input, errors.New("unit_cost is required when unit_cost_known is true")
	}
	return input, nil
}

func calculateProfitabilityItem(input ProfitabilityItemInput) (suppliercostdomain.ProfitCalculationResult, error) {
	return suppliercostdomain.CalculateProfit(suppliercostdomain.ProfitCalculationInput{
		ProductCode:             input.ProductCode,
		ProductName:             input.ProductName,
		SellingCurrency:         input.SellingCurrency,
		CostCurrency:            input.CostCurrency,
		ListPrice:               input.ListPrice,
		SalePrice:               input.SalePrice,
		UnitCost:                input.UnitCost,
		InboundShippingUnitCost: input.InboundShippingUnitCost,
		PackagingUnitCost:       input.PackagingUnitCost,
		OtherUnitCost:           input.OtherUnitCost,
	})
}

func profitCalculationRecord(result suppliercostdomain.ProfitCalculationResult) suppliercostdomain.ProductProfitCalculation {
	record := suppliercostdomain.ProductProfitCalculation{
		ProductCode:             result.ProductCode,
		ProductName:             result.ProductName,
		Currency:                result.Currency,
		ListPrice:               result.ListPrice,
		SalePrice:               result.SalePrice,
		EffectiveSellingPrice:   result.EffectiveSellingPrice,
		InboundShippingUnitCost: result.InboundShippingUnitCost,
		PackagingUnitCost:       result.PackagingUnitCost,
		OtherUnitCost:           result.OtherUnitCost,
		CalculationStatus:       result.Status,
		FormulaVersion:          result.FormulaVersion,
		CalculatedAt:            time.Now().UTC(),
	}
	if result.UnitCost != nil {
		record.UnitCost = *result.UnitCost
	}
	if result.LandedCost != nil {
		record.LandedCost = *result.LandedCost
	}
	if result.GrossProfit != nil {
		record.GrossProfit = *result.GrossProfit
	}
	if result.GrossMarginBPS != nil {
		record.GrossMarginBPS = *result.GrossMarginBPS
	}
	warnings, err := json.Marshal(result.Warnings)
	if err != nil {
		warnings = []byte("[]")
	}
	record.WarningsData = datatypes.JSON(warnings)
	return record
}

func profitabilityStatusReason(status string) string {
	switch status {
	case suppliercostdomain.ProfitStatusCurrencyMismatch:
		return "selling currency and cost currency must match"
	case suppliercostdomain.ProfitStatusInvalidSelling:
		return "effective selling price must be positive"
	case suppliercostdomain.ProfitStatusInvalidCost:
		return "unit supplier cost and additional costs must be valid and non-negative"
	default:
		return "profitability calculation is not ready"
	}
}

func normalizeCodesForProfitability(codes []string) []string {
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

func profitabilityRecordCodes(records []suppliercostdomain.ProductProfitCalculation) []string {
	codes := make([]string, 0, len(records))
	for _, record := range records {
		codes = append(codes, record.ProductCode)
	}
	return codes
}

func IsProfitabilityNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
