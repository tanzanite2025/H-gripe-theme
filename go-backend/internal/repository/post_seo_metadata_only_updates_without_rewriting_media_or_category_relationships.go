package repository

import (
	"commerce-platform/internal/domain/post"
	"gorm.io/gorm"
)

func (r *PostRepository) UpdatePostSEOMetadataFields(id uint, title, description, canonicalURL *string) error {
	updates := make(map[string]interface{})
	if title != nil {
		updates["meta_title"] = *title
	}
	if description != nil {
		updates["meta_desc"] = *description
	}
	if canonicalURL != nil {
		updates["canonical_url"] = *canonicalURL
	}
	if len(updates) == 0 {
		return nil
	}
	result := r.db.Model(&post.Post{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
