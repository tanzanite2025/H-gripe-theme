package product

import "time"

const (
	OptionValueRelationRequires  = "requires"
	OptionValueRelationConflicts = "conflicts"
)

// ProductOptionValueRelation expresses a small, explicit dependency between
// materialized product option values. It intentionally uses product-local
// option value IDs rather than labels or template IDs.
type ProductOptionValueRelation struct {
	ID                    uint      `gorm:"primarykey" json:"id"`
	ProductID             uint      `gorm:"not null;index;uniqueIndex:idx_product_option_value_relation_key" json:"product_id"`
	SourceOptionValueID   uint      `gorm:"not null;index;uniqueIndex:idx_product_option_value_relation_key" json:"source_option_value_id"`
	TargetOptionValueID   uint      `gorm:"not null;index;uniqueIndex:idx_product_option_value_relation_key" json:"target_option_value_id"`
	RelationType          string    `gorm:"type:varchar(16);not null;uniqueIndex:idx_product_option_value_relation_key" json:"relation_type"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (ProductOptionValueRelation) TableName() string { return "product_option_value_relations" }

func IsValidOptionValueRelationType(value string) bool {
	return value == OptionValueRelationRequires || value == OptionValueRelationConflicts
}
