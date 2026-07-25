package review

// ReviewSummary 评价统计
type ReviewSummary struct {
	ProductID     uint    `json:"product_id"`
	TotalReviews  int     `json:"total_reviews"`
	AverageRating float64 `json:"average_rating"`
	Rating5Count  int     `json:"rating_5_count"`
	Rating4Count  int     `json:"rating_4_count"`
	Rating3Count  int     `json:"rating_3_count"`
	Rating2Count  int     `json:"rating_2_count"`
	Rating1Count  int     `json:"rating_1_count"`
}
