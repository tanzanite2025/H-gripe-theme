package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/faqcontent"
	"commerce-platform/internal/pkg/safehtml"
)

func (r *MediaRepository) faqReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	if !r.hasTable("faqs") {
		return references, nil
	}
	// The FAQ renderer removes img/video tags, but permits file links. Parse
	// that same sanitized subset, including historical encoded storage aliases.
	for afterID := uint(0); ; {
		var rows []struct {
			ID                               uint
			Question, Answer, AnswerImageURL string
			DeletedAt                        *time.Time
		}
		if err := r.db.Table("faqs").Select("id, question, answer, answer_image_url, deleted_at").Where("id > ? AND (COALESCE(answer, '') <> '' OR COALESCE(answer_image_url, '') <> '')", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			var fields []string
			if r.matchesExactRegisteredOrTrustedMediaReference(query, row.AnswerImageURL) {
				fields = append(fields, "answer_image_url")
			}
			answer, err := faqcontent.SanitizeAnswer(row.Answer)
			if err != nil {
				return nil, err
			}
			document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(answer)
			if err != nil {
				return nil, err
			}
			for _, attribute := range document.References {
				if r.matchesExactRegisteredOrTrustedMediaReference(query, *attribute.URL) {
					fields = append(fields, "answer")
					break
				}
			}
			if len(fields) == 0 {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryContent, "faq", row.ID, 0, namedMediaReferenceLabel("FAQ", row.ID, row.Question), strings.Join(fields, ", "))
			reference.UsageState = "active"
			if row.DeletedAt != nil {
				reference.UsageState = "retained_deleted_parent"
				reference.Label += "（已删除 FAQ 保留）"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}
