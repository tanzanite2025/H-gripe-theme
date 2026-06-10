package showcase

import (
	"time"

	"gorm.io/datatypes"
)

// Status 定义审核状态
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// Kind 定义图片类型
const (
	KindUser  = "user"
	KindBrand = "brand"
)

// Showcase 买家秀/骑行墙记录
type Showcase struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Kind      string    `gorm:"type:varchar(20);default:'user'" json:"kind"`
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	
	// UGC 属性
	Region    string    `gorm:"type:varchar(100)" json:"region"`
	Location  string    `gorm:"type:varchar(100)" json:"location"`
	Nickname  string    `gorm:"type:varchar(100)" json:"nickname"`
	BikeModel string    `gorm:"type:varchar(100)" json:"bike_model"`
	Notes     string    `gorm:"type:text" json:"notes"`

	// 推荐配件连接 (rim, wheel, hub, tire)
	ProductRefs datatypes.JSON `gorm:"type:json" json:"product_refs"`
	
	// 图集 (URL 数组)
	Images      datatypes.JSON `gorm:"type:json" json:"gallery_images"`

	// 状态机
	Status         string    `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	RejectedReason string    `gorm:"type:text" json:"rejected_reason"`
	ApprovedAt     *time.Time `json:"approved_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Comment 评论
type Comment struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ShowcaseID uint      `gorm:"index;not null" json:"showcase_id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	Author     string    `gorm:"type:varchar(100)" json:"author"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Location   string    `gorm:"type:varchar(100)" json:"location"`
	
	Status     string    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
