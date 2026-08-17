package ticket

import (
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"commerce-platform/internal/pkg/ugc"

	"github.com/gin-gonic/gin"
)

func TestNormalizeTicketMessageTypeAllowsFAQ(t *testing.T) {
	if got := normalizeTicketMessageType(" FAQ "); got != "faq" {
		t.Fatalf("expected faq message type, got %q", got)
	}
}

func TestSanitizeTicketMessageAttachmentsAllowsMultipleForAuthenticatedUsers(t *testing.T) {
	handler := NewHandler(nil)

	got, err := handler.sanitizeTicketMessageAttachments([]string{
		"/uploads/a.jpg",
		"/uploads/b.png",
		"/uploads/c.webp",
		"/uploads/d.gif",
	}, ticketMessageAttachmentMaxCount)
	if err != nil {
		t.Fatalf("sanitizeTicketMessageAttachments() error = %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 attachments, got %d: %#v", len(got), got)
	}
}

func TestSanitizeTicketMessageAttachmentsRejectsGuestOverLimit(t *testing.T) {
	handler := NewHandler(nil)

	_, err := handler.sanitizeTicketMessageAttachments([]string{"/uploads/a.jpg", "/uploads/b.jpg"}, publicGuestMessageAttachmentMaxCount)
	if !errors.Is(err, ugc.ErrAttachmentTooMany) {
		t.Fatalf("sanitizeTicketMessageAttachments() error = %v, want ErrAttachmentTooMany", err)
	}
}

func TestPublicCustomerServiceMessageAttachmentLimitDependsOnLoginState(t *testing.T) {
	guest := &gin.Context{}
	if got := publicCustomerServiceMessageAttachmentLimit(guest); got != publicGuestMessageAttachmentMaxCount {
		t.Fatalf("guest attachment limit = %d, want %d", got, publicGuestMessageAttachmentMaxCount)
	}

	member := &gin.Context{}
	member.Set("user_id", uint(42))
	if got := publicCustomerServiceMessageAttachmentLimit(member); got != publicMemberMessageAttachmentMaxCount {
		t.Fatalf("member attachment limit = %d, want %d", got, publicMemberMessageAttachmentMaxCount)
	}
}

func TestPublicCustomerServiceMessageResponseCanonicalizesStoredMedia(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	item := ticket.TicketMessage{
		Attachments: `["http://media.internal:8080/uploads/chat/photo.webp","https://cdn.example.test/sticker.webp"]`,
		Metadata:    `{"thumbnail":"http://media.internal:8080/uploads/chat/thumb.webp","answer_image_url":"http://media.internal:8080/uploads/faq/answer.webp","url":"/support/faqs"}`,
	}

	response := publicCustomerServiceMessageResponse(item, "conversation-1", "", "", nil, resolver)
	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal public message response: %v", err)
	}
	if strings.Contains(string(payload), "media.internal") {
		t.Fatalf("public message response leaked internal media origin: %s", payload)
	}

	attachments, ok := response["attachments"].([]string)
	if !ok || len(attachments) != 2 {
		t.Fatalf("unexpected attachments: %#v", response["attachments"])
	}
	if attachments[0] != "https://shop.example.test/uploads/chat/photo.webp" {
		t.Fatalf("attachment[0] = %q", attachments[0])
	}
	if attachments[1] != "https://cdn.example.test/sticker.webp" {
		t.Fatalf("attachment[1] = %q", attachments[1])
	}

	metadata, ok := response["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected metadata type: %#v", response["metadata"])
	}
	if metadata["thumbnail"] != "https://shop.example.test/uploads/chat/thumb.webp" {
		t.Fatalf("metadata thumbnail = %#v", metadata["thumbnail"])
	}
	if metadata["answer_image_url"] != "https://shop.example.test/uploads/faq/answer.webp" {
		t.Fatalf("metadata answer_image_url = %#v", metadata["answer_image_url"])
	}
	if metadata["url"] != "/support/faqs" {
		t.Fatalf("metadata url was rewritten: %#v", metadata["url"])
	}
}
