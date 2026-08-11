package repository

import (
	"fmt"
	"strings"
	"commerce-platform/internal/domain/media"
)

func (r *MediaRepository) faqReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("faqs") {
		return []media.AssetReference{}, nil
	}

	type row struct {
		ID       uint
		Question string
	}
	var rows []row
	if err := r.db.Table("faqs").
		Select("id, question").
		Where("deleted_at IS NULL AND answer_image_url IN ?", query.URLs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		label := fmt.Sprintf("FAQ #%d", item.ID)
		if strings.TrimSpace(item.Question) != "" {
			label = fmt.Sprintf("FAQ #%d：%s", item.ID, truncateMediaReferenceLabel(item.Question))
		}
		references = append(references, newMediaReference(
			media.ReferenceCategoryContent,
			"faq",
			item.ID,
			0,
			label,
			"answer_image_url",
		))
	}
	return references, nil
}

func (r *MediaRepository) postReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("posts") {
		return []media.AssetReference{}, nil
	}

	containsSQL, containsArgs := mediaReferenceContainsCondition([]string{"content"}, query.URLs)
	type row struct {
		ID          uint
		Title       string
		FeaturedImg string
		Content     string
	}
	var rows []row
	args := append([]interface{}{query.URLs}, containsArgs...)
	if err := r.db.Table("posts").
		Select("id, title, featured_img, content").
		Where("deleted_at IS NULL AND (featured_img IN ? OR "+containsSQL+")", args...).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		fields := make([]string, 0, 2)
		if containsMediaReferenceURL(query.URLs, item.FeaturedImg) {
			fields = append(fields, "featured_img")
		}
		if containsMediaReferenceURLInText(query.URLs, item.Content) {
			fields = append(fields, "content")
		}
		references = append(references, newMediaReference(
			media.ReferenceCategoryContent,
			"post",
			item.ID,
			0,
			namedMediaReferenceLabel("文章", item.ID, item.Title),
			strings.Join(fields, ", "),
		))
	}
	return references, nil
}

func (r *MediaRepository) showcaseReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("showcases") {
		return []media.AssetReference{}, nil
	}

	containsSQL, containsArgs := mediaReferenceContainsCondition([]string{"images"}, query.URLs)
	type row struct {
		ID    uint
		Title string
	}
	var rows []row
	if err := r.db.Table("showcases").
		Select("id, title").
		Where(containsSQL, containsArgs...).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		references = append(references, newMediaReference(
			media.ReferenceCategoryCommunity,
			"showcase",
			item.ID,
			0,
			namedMediaReferenceLabel("买家秀", item.ID, item.Title),
			"images",
		))
	}
	return references, nil
}

func (r *MediaRepository) reviewReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("reviews") {
		return []media.AssetReference{}, nil
	}

	containsSQL, containsArgs := mediaReferenceContainsCondition([]string{"images"}, query.URLs)
	type row struct {
		ID        uint
		ProductID uint
	}
	var rows []row
	if err := r.db.Table("reviews").
		Select("id, product_id").
		Where("deleted_at IS NULL AND ("+containsSQL+")", containsArgs...).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		references = append(references, newMediaReference(
			media.ReferenceCategoryCommunity,
			"review",
			item.ID,
			item.ProductID,
			fmt.Sprintf("商品 #%d 的评价 #%d", item.ProductID, item.ID),
			"images",
		))
	}
	return references, nil
}
