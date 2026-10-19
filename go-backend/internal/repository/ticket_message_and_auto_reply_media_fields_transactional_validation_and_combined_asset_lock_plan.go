package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
)

var ErrTicketStructuredMediaReferenceInvalid = errors.New("customer-service message media reference is invalid")
var ErrAutoReplyRuleMediaReferenceInvalid = errors.New("automatic-reply media reference is invalid")
var ErrAutoReplyRuleChangedBeforeDelivery = errors.New("automatic-reply rule changed before delivery")

func (r *TicketRepository) ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver PublicBusinessMediaReferenceObjectKeyResolver) {
	if r != nil {
		r.businessMediaReferenceResolver = resolver
	}
}

// A rule's previous fields must come from its locked authority row. Messages
// always pass nil: copying a rule or another message creates a new placement.
func (r *TicketRepository) validateAndCanonicalizeTicketMediaFieldsInOneAssetLockPlan(messageType string, metadata, attachments *string, previous *ticket.AutoReplyRule, managedPlan *resolvedCustomerServiceMessageAttachmentPlan, legacyAttachmentActors ...media.MediaUploadActorIdentity) error {
	document, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(messageType, *metadata)
	if err != nil {
		return fmt.Errorf("%w: invalid structured metadata", ErrTicketStructuredMediaReferenceInvalid)
	}
	previousFields := make(map[string]string)
	var previousAttachments []string
	if previous != nil && previous.MessageType == messageType {
		previousDocument, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(previous.MessageType, previous.Metadata)
		if err != nil {
			return fmt.Errorf("%w: invalid retained structured metadata", ErrTicketStructuredMediaReferenceInvalid)
		}
		for _, field := range previousDocument.Fields {
			previousFields[field.Path] = field.URL
		}
		if strings.TrimSpace(previous.Attachments) != "" {
			_ = json.Unmarshal([]byte(previous.Attachments), &previousAttachments)
		}
	}
	var bindings []publicBusinessMediaReferenceBinding
	for index := range document.Fields {
		field := &document.Fields[index]
		if previous != nil && previousFields[field.Path] == field.URL {
			continue
		}
		bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &field.URL, ExpectedMediaType: field.ExpectedMediaType})
	}
	var attachmentValues []string
	if strings.TrimSpace(*attachments) != "" {
		if err := json.Unmarshal([]byte(*attachments), &attachmentValues); err != nil {
			return fmt.Errorf("%w: invalid attachment list", ErrTicketStructuredMediaReferenceInvalid)
		}
	}
	originalAttachments := append([]string(nil), attachmentValues...)
	legacyAttachmentBindings := make([]publicBusinessMediaReferenceBinding, 0)
	legacyAttachmentAssetIDs := make(map[string]uint)
	legacyActor := media.MediaUploadActorIdentity{}
	if len(legacyAttachmentActors) > 0 {
		legacyActor = legacyAttachmentActors[0]
	}
	for index := range attachmentValues {
		if previous != nil && index < len(previousAttachments) && previousAttachments[index] == attachmentValues[index] {
			continue
		}
		expectedType := ""
		if messageType == "image" || messageType == "video" {
			expectedType = messageType
		}
		binding := publicBusinessMediaReferenceBinding{URL: &attachmentValues[index], ExpectedMediaType: expectedType}
		owned, err := r.resolveVerifiedUploaderLegacyPublicConversationAttachment(attachmentValues[index], legacyActor)
		if err != nil {
			return fmt.Errorf("%w: legacy conversation attachment ownership is unavailable", ErrTicketStructuredMediaReferenceInvalid)
		}
		if owned != nil {
			legacyAttachmentAssetIDs[attachmentValues[index]] = owned.ID
			legacyAttachmentBindings = append(legacyAttachmentBindings, binding)
		} else {
			bindings = append(bindings, binding)
		}
	}
	if messageType == "image" || messageType == "video" {
		publicBindings := make([]publicBusinessMediaReferenceBinding, 0, len(bindings))
		for _, binding := range bindings {
			if legacyAttachmentAssetIDs[*binding.URL] != 0 {
				legacyAttachmentBindings = append(legacyAttachmentBindings, binding)
			} else {
				publicBindings = append(publicBindings, binding)
			}
		}
		bindings = publicBindings
	}
	var references []string
	for _, binding := range bindings {
		references = append(references, *binding.URL)
	}
	var publicAttachmentIDs, privateAttachmentIDs []uint
	for _, id := range legacyAttachmentAssetIDs {
		privateAttachmentIDs = append(privateAttachmentIDs, id)
	}
	if managedPlan != nil {
		for _, request := range managedPlan.requests {
			if request.sessionID != "" {
				privateAttachmentIDs = append(privateAttachmentIDs, request.assetID)
			} else {
				publicAttachmentIDs = append(publicAttachmentIDs, request.assetID)
			}
		}
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentitiesWithAdditionalPrivateAttachmentLocks(r.db, r.businessMediaReferenceResolver, references, publicAttachmentIDs, privateAttachmentIDs)
	if err != nil {
		return mapPublicMediaIdentityErrorToTicketMediaError(err)
	}
	for _, binding := range bindings {
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding); err != nil {
			return mapPublicMediaIdentityErrorToTicketMediaError(err)
		}
	}
	for _, binding := range legacyAttachmentBindings {
		asset, err := r.resolveVerifiedUploaderLegacyPublicConversationAttachment(*binding.URL, legacyActor)
		if err != nil || asset == nil || asset.ID != legacyAttachmentAssetIDs[*binding.URL] || (binding.ExpectedMediaType != "" && asset.MediaType != binding.ExpectedMediaType) {
			return fmt.Errorf("%w: legacy conversation attachment is unavailable", ErrTicketStructuredMediaReferenceInvalid)
		}
		*binding.URL = asset.URL
	}
	for _, id := range publicAttachmentIDs {
		reference, assetID := "", &id
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(publicBusinessMediaReferenceBinding{URL: &reference, AssetID: &assetID}); err != nil {
			return mapPublicMediaIdentityErrorToTicketMediaError(err)
		}
	}
	*metadata, err = document.RenderMetadataWithCanonicalMediaFields()
	if err != nil {
		return err
	}
	for index := range attachmentValues {
		if attachmentValues[index] != originalAttachments[index] {
			encoded, err := json.Marshal(attachmentValues)
			if err != nil {
				return err
			}
			*attachments = string(encoded)
			break
		}
	}
	return nil
}

func mapPublicMediaIdentityErrorToTicketMediaError(err error) error {
	if errors.Is(err, ErrProductMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrTicketStructuredMediaReferenceInvalid, err)
	}
	return err
}

func (r *TicketRepository) validateAndCanonicalizeAutoReplyRuleMediaWithinTransaction(rule, previous *ticket.AutoReplyRule) error {
	if r.businessMediaReferenceResolver == nil && (previous == nil || previous.MessageType != rule.MessageType || previous.Metadata != rule.Metadata || previous.Attachments != rule.Attachments) {
		document, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(rule.MessageType, rule.Metadata)
		var attachments []string
		_ = json.Unmarshal([]byte(rule.Attachments), &attachments)
		if err != nil || len(document.Fields) > 0 || len(attachments) > 0 {
			return fmt.Errorf("%w: configured media identity resolver is required", ErrAutoReplyRuleMediaReferenceInvalid)
		}
	}
	err := r.validateAndCanonicalizeTicketMediaFieldsInOneAssetLockPlan(rule.MessageType, &rule.Metadata, &rule.Attachments, previous, nil)
	if errors.Is(err, ErrTicketStructuredMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrAutoReplyRuleMediaReferenceInvalid, err)
	}
	return err
}

func (r *TicketRepository) validateSelectedAutoReplyRuleStillCurrentUnderRowLock(selected *ticket.AutoReplyRule) error {
	if selected == nil || selected.ID == 0 {
		return ErrAutoReplyRuleChangedBeforeDelivery
	}
	var current ticket.AutoReplyRule
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, selected.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAutoReplyRuleChangedBeforeDelivery
		}
		return err
	}
	if !current.IsActive || !current.UpdatedAt.Equal(selected.UpdatedAt) || current.MessageType != selected.MessageType || current.Metadata != selected.Metadata || current.Attachments != selected.Attachments || current.ReplyMessage != selected.ReplyMessage {
		return ErrAutoReplyRuleChangedBeforeDelivery
	}
	return nil
}

// The conversation command owns the parent lock and transaction. Its metadata,
// shared ID bindings and private upload sessions acquire one global asset order.
func (r *TicketRepository) CreateCustomerServiceMessageAndBindMediaWithinTransaction(message *ticket.TicketMessage, identity media.MediaUploadActorIdentity) error {
	var plan *resolvedCustomerServiceMessageAttachmentPlan
	var err error
	if len(message.RequestedMediaAttachmentBindings) > 0 {
		plan, err = r.resolveCustomerServiceMessageAttachmentRequestsWithinTransaction(identity, message.RequestedMediaAttachmentBindings)
		if err != nil {
			return err
		}
	}
	if err := r.validateAndCanonicalizeTicketMediaFieldsInOneAssetLockPlan(message.MessageType, &message.Metadata, &message.Attachments, nil, plan, identity); err != nil {
		if plan != nil && errors.Is(err, ErrTicketStructuredMediaReferenceInvalid) {
			return errors.Join(ErrCustomerServiceAttachmentBindingUnavailable, err)
		}
		return err
	}
	if err := r.db.Omit(clause.Associations).Create(message).Error; err != nil {
		return err
	}
	if plan != nil {
		return r.bindResolvedCustomerServiceMessageAttachmentsWithinTransaction(message.TicketID, message.ID, identity, plan)
	}
	return nil
}
