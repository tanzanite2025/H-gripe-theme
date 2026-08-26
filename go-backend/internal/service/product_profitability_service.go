package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	procurementdomain "commerce-platform/internal/domain/procurement"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	MaxProfitabilityBatchItems = 100
	MaxProfitabilityCodeItems  = 100
)

var (
	ErrProductProfitabilityInvalid                = errors.New("product profitability input is invalid")
	ErrProductProfitabilityBatchLarge             = errors.New("product profitability batch is too large")
	ErrProductProfitabilityProcurementUnavailable = errors.New("product procurement repository is unavailable")
)

type ProductProfitabilityService struct {
	repo            *repository.ProductProfitCalculationRepository
	procurementRepo *repository.ProductProcurementRepository
}

type ProfitabilityProcurementInput struct {
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

	ListPrice     float64
	SalePrice     *float64
	PurchasePrice *float64
	// PurchasePriceKnown distinguishes an explicit zero cost from an omitted
	// cost. Unknown cost values are never persisted as a ready snapshot.
	PurchasePriceKnown bool

	InboundShippingUnitCost float64
	PackagingUnitCost       float64
	OtherUnitCost           float64

	Procurement *ProfitabilityProcurementInput
}

type ProfitabilitySkippedItem struct {
	ProductCode string `json:"product_code"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}

type ProfitabilityBatchResult struct {
	Records []procurementdomain.ProductProfitCalculation `json:"records"`
	Skipped []ProfitabilitySkippedItem                   `json:"skipped"`
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

func NewProductProfitabilityServiceWithProcurement(
	repo *repository.ProductProfitCalculationRepository,
	procurementRepo *repository.ProductProcurementRepository,
) *ProductProfitabilityService {
	return &ProductProfitabilityService{
		repo:            repo,
		procurementRepo: procurementRepo,
	}
}

func (s *ProductProfitabilityService) Preview(items []ProfitabilityItemInput) ([]procurementdomain.ProfitCalculationResult, error) {
	if len(items) > MaxProfitabilityBatchItems {
		return nil, ErrProductProfitabilityBatchLarge
	}

	results := make([]procurementdomain.ProfitCalculationResult, 0, len(items))
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

func (s *ProductProfitabilityService) ListByCodes(codes []string) ([]procurementdomain.ProductProfitCalculation, error) {
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

	records := make([]procurementdomain.ProductProfitCalculation, 0, len(items))
	procurementRecords := make([]procurementdomain.ProductProcurement, 0, len(items))
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
		if normalized.Procurement != nil && result.Status == procurementdomain.ProfitStatusMissingPurchase {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Status:      result.Status,
				Reason:      "purchase_price is required when procurement data is supplied",
			})
			continue
		}
		if result.Status == procurementdomain.ProfitStatusMissingPurchase {
			skipped = append(skipped, ProfitabilitySkippedItem{
				ProductCode: normalized.ProductCode,
				Status:      result.Status,
				Reason:      "purchase price is not known",
			})
			clearCodes = append(clearCodes, normalized.ProductCode)
			continue
		}
		if result.Status == procurementdomain.ProfitStatusCurrencyMismatch ||
			result.Status == procurementdomain.ProfitStatusInvalidSelling ||
			result.Status == procurementdomain.ProfitStatusInvalidCost {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Status:      result.Status,
				Reason:      profitabilityStatusReason(result.Status),
			})
			continue
		}
		if result.PurchasePrice == nil {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Reason:      "purchase_price is required",
			})
			continue
		}

		costCurrency := normalized.CostCurrency
		if strings.TrimSpace(costCurrency) == "" {
			costCurrency = normalized.SellingCurrency
		}
		procurementInput := productProcurementSnapshotInput{
			ProductCode: normalized.ProductCode,
			ProductName: normalized.ProductName,
			ProductProcurementDetailsInput: ProductProcurementDetailsInput{
				PurchasePrice: result.PurchasePrice,
				Currency:      costCurrency,
			},
		}
		if normalized.Procurement != nil {
			procurementInput.SupplierName = normalized.Procurement.SupplierName
			procurementInput.SupplierContactName = normalized.Procurement.SupplierContactName
			procurementInput.SupplierPhone = normalized.Procurement.SupplierPhone
			procurementInput.SupplierEmail = normalized.Procurement.SupplierEmail
			procurementInput.LeadTimeDays = normalized.Procurement.LeadTimeDays
			procurementInput.MinimumOrderQuantity = normalized.Procurement.MinimumOrderQuantity
		}
		procurementInput.InboundShippingUnitCost = normalized.InboundShippingUnitCost
		procurementInput.PackagingUnitCost = normalized.PackagingUnitCost
		procurementInput.OtherUnitCost = normalized.OtherUnitCost
		procurementRecord, procurementErr := normalizeProductProcurementSnapshotInput(procurementInput)
		if procurementErr != nil {
			issues = append(issues, ProfitabilityItemIssue{
				Index:       index,
				ProductCode: normalized.ProductCode,
				Reason:      procurementErr.Error(),
			})
			continue
		}
		procurementRecords = append(procurementRecords, *procurementRecord)
		records = append(records, profitCalculationRecord(result))
	}

	if len(issues) > 0 {
		return ProfitabilityBatchResult{Skipped: skipped}, &ProfitabilityBatchValidationError{Items: issues}
	}
	if (len(procurementRecords) > 0 || len(clearCodes) > 0) && s.procurementRepo == nil {
		return ProfitabilityBatchResult{Skipped: skipped}, ErrProductProfitabilityProcurementUnavailable
	}
	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		if len(procurementRecords) > 0 {
			if err := s.procurementRepo.UpsertInTx(tx, procurementRecords); err != nil {
				return err
			}
		}
		if len(clearCodes) > 0 {
			if err := s.procurementRepo.DeleteByProductCodesInTx(tx, clearCodes); err != nil {
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
	return ProfitabilityBatchResult{Records: []procurementdomain.ProductProfitCalculation{}, Skipped: skipped}, nil
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
	if !input.PurchasePriceKnown {
		input.PurchasePrice = nil
	}
	if input.PurchasePriceKnown && input.PurchasePrice == nil {
		return input, errors.New("purchase_price is required when purchase_price_known is true")
	}
	return input, nil
}

func calculateProfitabilityItem(input ProfitabilityItemInput) (procurementdomain.ProfitCalculationResult, error) {
	return procurementdomain.CalculateProfit(procurementdomain.ProfitCalculationInput{
		ProductCode:             input.ProductCode,
		ProductName:             input.ProductName,
		SellingCurrency:         input.SellingCurrency,
		CostCurrency:            input.CostCurrency,
		ListPrice:               input.ListPrice,
		SalePrice:               input.SalePrice,
		PurchasePrice:           input.PurchasePrice,
		InboundShippingUnitCost: input.InboundShippingUnitCost,
		PackagingUnitCost:       input.PackagingUnitCost,
		OtherUnitCost:           input.OtherUnitCost,
	})
}

func profitCalculationRecord(result procurementdomain.ProfitCalculationResult) procurementdomain.ProductProfitCalculation {
	record := procurementdomain.ProductProfitCalculation{
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
	if result.PurchasePrice != nil {
		record.PurchasePrice = *result.PurchasePrice
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
	case procurementdomain.ProfitStatusCurrencyMismatch:
		return "selling currency and cost currency must match"
	case procurementdomain.ProfitStatusInvalidSelling:
		return "effective selling price must be positive"
	case procurementdomain.ProfitStatusInvalidCost:
		return "purchase price and additional costs must be valid and non-negative"
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

func profitabilityRecordCodes(records []procurementdomain.ProductProfitCalculation) []string {
	codes := make([]string, 0, len(records))
	for _, record := range records {
		codes = append(codes, record.ProductCode)
	}
	return codes
}

func IsProfitabilityNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
