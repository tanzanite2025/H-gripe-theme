package repository

import (
	"commerce-platform/internal/domain/media"
	"fmt"
	"strings"
)

func (r *MediaRepository) registrationReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("product_registrations") {
		return []media.AssetReference{}, nil
	}

	type row struct {
		ID           uint
		SerialNumber string
	}
	var rows []row
	if err := r.db.Table("product_registrations").
		Select("id, serial_number").
		Where("deleted_at IS NULL AND purchase_proof IN ?", query.URLs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		label := fmt.Sprintf("产品注册 #%d", item.ID)
		if strings.TrimSpace(item.SerialNumber) != "" {
			label = fmt.Sprintf("产品注册 #%d：%s", item.ID, item.SerialNumber)
		}
		references = append(references, newMediaReference(
			media.ReferenceCategoryCustomer,
			"product_registration",
			item.ID,
			0,
			label,
			"purchase_proof",
		))
	}
	return references, nil
}

func (r *MediaRepository) warrantyClaimReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("warranty_claims") {
		return []media.AssetReference{}, nil
	}

	containsSQL, containsArgs := mediaReferenceContainsCondition([]string{"images"}, query.URLs)
	type row struct {
		ID       uint
		Images   string
		VideoURL string
	}
	var rows []row
	args := append([]interface{}{query.URLs}, containsArgs...)
	if err := r.db.Table("warranty_claims").
		Select("id, images, video_url").
		Where("deleted_at IS NULL AND (video_url IN ? OR "+containsSQL+")", args...).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		fields := make([]string, 0, 2)
		if containsMediaReferenceURL(query.URLs, item.VideoURL) {
			fields = append(fields, "video_url")
		}
		if containsMediaReferenceURLInText(query.URLs, item.Images) {
			fields = append(fields, "images")
		}
		references = append(references, newMediaReference(
			media.ReferenceCategoryCustomer,
			"warranty_claim",
			item.ID,
			0,
			fmt.Sprintf("保修申请 #%d", item.ID),
			strings.Join(fields, ", "),
		))
	}
	return references, nil
}

func (r *MediaRepository) suggestionFeedbackReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("suggestion_feedback") {
		return []media.AssetReference{}, nil
	}

	containsSQL, containsArgs := mediaReferenceContainsCondition([]string{"attachments"}, query.URLs)
	type row struct {
		ID uint
	}
	var rows []row
	if err := r.db.Table("suggestion_feedback").
		Select("id").
		Where("deleted_at IS NULL AND ("+containsSQL+")", containsArgs...).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		references = append(references, newMediaReference(
			media.ReferenceCategoryCustomer,
			"suggestion_feedback",
			item.ID,
			0,
			fmt.Sprintf("建议反馈 #%d", item.ID),
			"attachments",
		))
	}
	return references, nil
}

func (r *MediaRepository) ticketAutoReplyReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("ticket_auto_replies") {
		return []media.AssetReference{}, nil
	}

	containsSQL, containsArgs := mediaReferenceContainsCondition([]string{"attachments"}, query.URLs)
	type row struct {
		ID uint
	}
	var rows []row
	if err := r.db.Table("ticket_auto_replies").
		Select("id").
		Where(containsSQL, containsArgs...).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		references = append(references, newMediaReference(
			media.ReferenceCategorySupport,
			"ticket_auto_reply",
			item.ID,
			0,
			fmt.Sprintf("客服自动回复 #%d", item.ID),
			"attachments",
		))
	}
	return references, nil
}

func (r *MediaRepository) ticketMessageReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("ticket_messages") {
		return []media.AssetReference{}, nil
	}

	containsSQL, containsArgs := mediaReferenceContainsCondition([]string{"attachments"}, query.URLs)
	type row struct {
		ID       uint
		TicketID uint
	}
	var rows []row
	if err := r.db.Table("ticket_messages").
		Select("id, ticket_id").
		Where(containsSQL, containsArgs...).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0, len(rows))
	for _, item := range rows {
		references = append(references, newMediaReference(
			media.ReferenceCategorySupport,
			"ticket_message",
			item.ID,
			item.TicketID,
			fmt.Sprintf("工单 #%d 的消息 #%d", item.TicketID, item.ID),
			"attachments",
		))
	}
	return references, nil
}
