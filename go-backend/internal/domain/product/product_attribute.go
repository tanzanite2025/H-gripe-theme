package product

import "time"

// ProductAttribute 商品属性
type ProductAttribute struct {
	ID           uint             `gorm:"primarykey" json:"id"`
	Name         string           `gorm:"type:varchar(120);not null" json:"name"`
	Slug         string           `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	Type         string           `gorm:"type:varchar(32);default:'select';not null" json:"type"` // select, text
	SortOrder    int              `gorm:"default:0;not null" json:"sort_order"`
	IsFilterable bool             `gorm:"default:true;not null" json:"is_filterable"`
	AffectsSKU   bool             `gorm:"default:true;not null" json:"affects_sku"`
	AffectsStock bool             `gorm:"default:false;not null" json:"affects_stock"`
	IsEnabled    bool             `gorm:"default:true;not null" json:"is_enabled"`
	Meta         string           `gorm:"type:text" json:"meta"`
	Values       []AttributeValue `gorm:"foreignKey:AttributeID" json:"values"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// TableName 指定表名
func (ProductAttribute) TableName() string {
	return "product_attributes"
}

// AttributeValue 属性值
type AttributeValue struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	AttributeID uint      `gorm:"not null;index" json:"attribute_id"`
	Name        string    `gorm:"type:varchar(120);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(120);not null" json:"slug"`
	Value       string    `gorm:"type:varchar(255)" json:"value"`
	SortOrder   int       `gorm:"default:0;not null" json:"sort_order"`
	IsEnabled   bool      `gorm:"default:true;not null" json:"is_enabled"`
	Meta        string    `gorm:"type:text" json:"meta"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AttributeValue) TableName() string {
	return "product_attribute_values"
}
