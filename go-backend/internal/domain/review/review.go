package review

import (
	"tanzanite/internal/domain/product"
	"tanzanite/internal/domain/user"
	"time"

	"gorm.io/gorm"
)

// Review 商品评价
type Review struct {
	ID           uint             `gorm:"primarykey" json:"id"`
	ProductID    uint             `gorm:"not null;index" json:"product_id"`
	UserID       uint             `gorm:"not null;index" json:"user_id"`
	OrderID      uint             `gorm:"index" json:"order_id"`
	Product      *product.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	User         *user.User       `gorm:"foreignKey:UserID" json:"-"`
	Rating       int              `gorm:"not null" json:"rating"` // 1-5星
	Title        string           `json:"title"`
	Content      string           `gorm:"type:text" json:"content"`
	Images       string           `gorm:"type:text" json:"images"`               // JSON数组
	Pros         string           `gorm:"type:text" json:"pros"`                 // 优点
	Cons         string           `gorm:"type:text" json:"cons"`                 // 缺点
	Status       string           `gorm:"index;default:'pending'" json:"status"` // pending, approved, rejected
	Featured     bool             `gorm:"default:false" json:"featured"`         // 是否精选
	Verified     bool             `gorm:"default:false" json:"verified"`         // 是否已购买验证
	HelpfulCount int              `gorm:"default:0" json:"helpful_count"`        // 有用数
	ReplyContent string           `gorm:"type:text" json:"reply_content"`        // 商家回复
	RepliedAt    *time.Time       `json:"replied_at"`
	RepliedBy    uint             `json:"replied_by"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Review) TableName() string {
	return "reviews"
}
