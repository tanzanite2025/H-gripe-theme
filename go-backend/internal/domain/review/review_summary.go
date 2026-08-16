package review

// ReviewSummary 评价统计
type ReviewSummary struct {
	ProductID     uint    `gorm:"primaryKey;column:product_id" json:"product_id"`
	TotalReviews  int     `gorm:"column:total_reviews;not null;default:0" json:"total_reviews"`
	AverageRating float64 `gorm:"column:average_rating;not null;default:0" json:"average_rating"`
	Rating5Count  int     `gorm:"column:rating_5_count;not null;default:0" json:"rating_5_count"`
	Rating4Count  int     `gorm:"column:rating_4_count;not null;default:0" json:"rating_4_count"`
	Rating3Count  int     `gorm:"column:rating_3_count;not null;default:0" json:"rating_3_count"`
	Rating2Count  int     `gorm:"column:rating_2_count;not null;default:0" json:"rating_2_count"`
	Rating1Count  int     `gorm:"column:rating_1_count;not null;default:0" json:"rating_1_count"`
}

func (ReviewSummary) TableName() string {
	return "review_summaries"
}
