package service

import (
	"encoding/json"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
)

func customerServiceMessageMetadataWithoutDuplicateAttachmentURLs(value string) string {
	var metadata map[string]interface{}
	if json.Unmarshal([]byte(value), &metadata) != nil || metadata == nil {
		return value
	}
	for _, key := range []string{"url", "thumbnail", "image_url", "video_url", "attachment_url"} {
		delete(metadata, key)
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return ""
	}
	return string(payload)
}

func customerServiceMessageAttachmentBindingActorIdentity(conversation *ticket.Ticket, actor CustomerServiceRealtimeActor) (media.MediaUploadActorIdentity, error) {
	switch actor.Kind {
	case "customer":
		if actor.UserID != nil && conversation.CustomerUserID != nil && *actor.UserID == *conversation.CustomerUserID {
			return media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorCustomer, AccountID: *actor.UserID}, nil
		}
		if actor.UserID == nil && conversation.CustomerUserID == nil && conversation.VisitorSessionHash != "" {
			return media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor, VisitorSessionHash: conversation.VisitorSessionHash}, nil
		}
	case "agent":
		if actor.UserID != nil && *actor.UserID > 0 {
			return media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: *actor.UserID}, nil
		}
	}
	return media.MediaUploadActorIdentity{}, repository.ErrCustomerServiceAttachmentBindingUnavailable
}
