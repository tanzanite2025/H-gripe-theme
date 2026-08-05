package ticket

import (
	"errors"
	"testing"

	"tanzanite/internal/pkg/ugc"

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
