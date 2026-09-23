package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"

	"github.com/stretchr/testify/require"
)

type recordingDisputeEmailSender struct {
	to      []string
	subject string
	body    string
	err     error
}

func (s *recordingDisputeEmailSender) SendEmail(to []string, subject, body string) error {
	s.to, s.subject, s.body = append([]string(nil), to...), subject, body
	return s.err
}
func (s *recordingDisputeEmailSender) SendHTMLEmail([]string, string, string, interface{}) error {
	return nil
}
func (s *recordingDisputeEmailSender) SendPasswordReset(string, interface{}) error        { return nil }
func (s *recordingDisputeEmailSender) SendWelcomeEmail(string, interface{}) error         { return nil }

func TestOrderDisputeContactEmailOutboxHandlerSendsPayload(t *testing.T) {
	sender := &recordingDisputeEmailSender{}
	handler := NewOrderDisputeContactEmailOutboxHandler(sender)
	requestedAt := time.Now().UTC()
	payload, err := json.Marshal(outbox.OrderDisputeContactEmailPayload{
		OrderID:           9,
		Provider:          "stripe",
		DisputeID:         4,
		ProviderDisputeID: "dp_4",
		RecipientEmail:    "buyer@example.com",
		Subject:           "Regarding your order",
		Body:              "Please reply with more details.",
		RequestedAt:       requestedAt,
	})
	require.NoError(t, err)

	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderDisputeContactEmail,
		Payload:   payload,
	}))
	require.Equal(t, []string{"buyer@example.com"}, sender.to)
	require.Equal(t, "Regarding your order", sender.subject)
	require.Equal(t, "Please reply with more details.", sender.body)
}

func TestOrderDisputeContactEmailOutboxHandlerValidatesPayloadAndPropagatesFailure(t *testing.T) {
	handler := NewOrderDisputeContactEmailOutboxHandler(&recordingDisputeEmailSender{err: errors.New("smtp unavailable")})
	require.Error(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderDisputeContactEmail,
		Payload:   []byte(`{"order_id":1}`),
	}))

	badRecipient, err := json.Marshal(outbox.OrderDisputeContactEmailPayload{
		OrderID:        1,
		Provider:       "stripe",
		DisputeID:      2,
		RecipientEmail: "not-an-email",
		Subject:        "subject",
		Body:           "body",
		RequestedAt:    time.Now().UTC(),
	})
	require.NoError(t, err)
	require.Error(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderDisputeContactEmail,
		Payload:   badRecipient,
	}))
}
