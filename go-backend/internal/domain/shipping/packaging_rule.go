package shipping

import "time"

// PackagingRule 包装规格规则
type PackagingRule struct {
	ID          uint                 `gorm:"primarykey" json:"id"`
	RuleName    string               `gorm:"type:varchar(100);not null" json:"rule_name"`
	Description string               `gorm:"type:text" json:"description"`
	BoxWeight   float64              `gorm:"type:decimal(10,3);default:0;not null" json:"box_weight"`
	BoxLength   float64              `gorm:"type:decimal(10,2);default:0;not null" json:"box_length"`
	BoxWidth    float64              `gorm:"type:decimal(10,2);default:0;not null" json:"box_width"`
	BoxHeight   float64              `gorm:"type:decimal(10,2);default:0;not null" json:"box_height"`
	MaxWeight   float64              `gorm:"type:decimal(10,3);default:0;not null" json:"max_weight"`
	IsActive    bool                 `gorm:"default:true;index" json:"is_active"`
	Applies     []PackagingRuleApply `gorm:"foreignKey:RuleID" json:"applies"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// TableName 指定表名
func (PackagingRule) TableName() string {
	return "shipping_packaging_rules"
}

// PackagingRuleApply 包装规则应用的产品
type PackagingRuleApply struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	RuleID    uint           `gorm:"not null;index;uniqueIndex:idx_shipping_packaging_rule_apply_rule_product" json:"rule_id"`
	ProductID uint           `gorm:"not null;index;uniqueIndex:idx_shipping_packaging_rule_apply_rule_product;uniqueIndex:idx_shipping_packaging_rule_apply_product" json:"product_id"`
	Rule      *PackagingRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// TableName 指定表名
func (PackagingRuleApply) TableName() string {
	return "shipping_packaging_rule_applies"
}
