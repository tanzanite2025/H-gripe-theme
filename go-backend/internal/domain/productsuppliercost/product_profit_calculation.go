package productsuppliercost

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/datatypes"
)

// ProductProfitCalculation stores the current profitability snapshot for one
// local product-code reference. It intentionally has no catalog foreign key.
type ProductProfitCalculation struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	ProductCode string `gorm:"size:160;not null;uniqueIndex" json:"product_code"`
	ProductName string `gorm:"size:255;not null" json:"product_name"`
	Currency    string `gorm:"size:3;not null" json:"currency"`

	ListPriceMinor             int64  `gorm:"column:list_price_minor;not null" json:"list_price_minor"`
	SalePriceMinor             *int64 `gorm:"column:sale_price_minor" json:"sale_price_minor,omitempty"`
	EffectiveSellingPriceMinor int64  `gorm:"column:effective_selling_price_minor;not null" json:"effective_selling_price_minor"`

	UnitCostMinor                int64 `gorm:"column:purchase_price_minor;not null;default:0" json:"unit_cost_minor"`
	InboundShippingUnitCostMinor int64 `gorm:"column:inbound_shipping_unit_cost_minor;not null;default:0" json:"inbound_shipping_unit_cost_minor"`
	PackagingUnitCostMinor       int64 `gorm:"column:packaging_unit_cost_minor;not null;default:0" json:"packaging_unit_cost_minor"`
	OtherUnitCostMinor           int64 `gorm:"column:other_unit_cost_minor;not null;default:0" json:"other_unit_cost_minor"`

	LandedCostMinor  int64 `gorm:"column:landed_cost_minor;not null;default:0" json:"landed_cost_minor"`
	GrossProfitMinor int64 `gorm:"column:gross_profit_minor;not null;default:0" json:"gross_profit_minor"`
	GrossMarginBPS          int      `gorm:"not null" json:"gross_margin_bps"`

	CalculationStatus string         `gorm:"size:40;not null;default:'ready'" json:"calculation_status"`
	FormulaVersion    string         `gorm:"size:32;not null" json:"formula_version"`
	WarningsData      datatypes.JSON `gorm:"column:warnings;type:jsonb;not null;default:'[]'" json:"warnings"`
	CalculatedAt      time.Time      `gorm:"not null" json:"calculated_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

func (ProductProfitCalculation) TableName() string {
	return "product_profit_calculations"
}

func (p ProductProfitCalculation) Warnings() []string {
	if len(p.WarningsData) == 0 {
		return []string{}
	}
	var warnings []string
	if err := json.Unmarshal(p.WarningsData, &warnings); err != nil || warnings == nil {
		return []string{}
	}
	return warnings
}

func NormalizeProductProfitCalculation(record *ProductProfitCalculation) {
	if record == nil {
		return
	}
	record.ProductCode = strings.TrimSpace(record.ProductCode)
	record.ProductName = strings.TrimSpace(record.ProductName)
	record.Currency = strings.ToUpper(strings.TrimSpace(record.Currency))
	record.CalculationStatus = strings.TrimSpace(record.CalculationStatus)
	record.FormulaVersion = strings.TrimSpace(record.FormulaVersion)
	if len(record.WarningsData) == 0 {
		record.WarningsData = datatypes.JSON([]byte("[]"))
	}
}
