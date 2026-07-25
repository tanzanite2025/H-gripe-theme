package post

import "time"

// PostListResponse 文章列表响应
type PostListResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	Status      string     `json:"status"`
	AuthorID    uint       `json:"author_id"`
	Locale      string     `json:"locale"`
	FeaturedImg string     `json:"featured_image"`
	ViewCount   int        `json:"view_count"`
	Tags        string     `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at"`
}

// ToListResponse 转换为列表响应
func (p *Post) ToListResponse() *PostListResponse {
	return &PostListResponse{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Excerpt:     p.Excerpt,
		Status:      p.Status,
		AuthorID:    p.AuthorID,
		Locale:      p.Locale,
		FeaturedImg: p.FeaturedImg,
		ViewCount:   p.ViewCount,
		Tags:        p.Tags,
		CreatedAt:   p.CreatedAt,
		PublishedAt: p.PublishedAt,
	}
}
