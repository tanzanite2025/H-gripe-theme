package review

import (
	"time"

	reviewdomain "commerce-platform/internal/domain/review"
)

// ProductReviewPublicListItem contains the fields used by the public product
// review list. Business ownership and moderation records stay in the domain.
type ProductReviewPublicListItem struct {
	ID           uint       `json:"id"`
	ProductID    uint       `json:"product_id"`
	Rating       int        `json:"rating"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Images       string     `json:"images"`
	Pros         string     `json:"pros"`
	Cons         string     `json:"cons"`
	Verified     bool       `json:"verified"`
	Featured     bool       `json:"featured"`
	HelpfulCount int        `json:"helpful_count"`
	ReplyContent string     `json:"reply_content"`
	RepliedAt    *time.Time `json:"replied_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

func projectProductReviewsToPublicListFields(items []reviewdomain.Review) []ProductReviewPublicListItem {
	publicItems := make([]ProductReviewPublicListItem, 0, len(items))
	for _, item := range items {
		publicItems = append(publicItems, ProductReviewPublicListItem{
			ID: item.ID, ProductID: item.ProductID, Rating: item.Rating, Title: item.Title, Content: item.Content,
			Images: item.Images, Pros: item.Pros, Cons: item.Cons, Verified: item.Verified, Featured: item.Featured,
			HelpfulCount: item.HelpfulCount, ReplyContent: item.ReplyContent, RepliedAt: item.RepliedAt, CreatedAt: item.CreatedAt,
		})
	}
	return publicItems
}
