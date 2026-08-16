package admin

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/ticket"

	"github.com/stretchr/testify/require"
)

func TestAdminCustomerServiceConversationResponseUsesRecipientUnreadCount(t *testing.T) {
	response := adminCustomerServiceConversationResponse(ticket.Ticket{
		ID:                         42,
		CustomerServiceUnreadCount: 3,
		UpdatedAt:                  time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC),
		Messages: []ticket.TicketMessage{
			{Content: "Older customer message", IsStaff: false, IsRead: true},
			{Content: "Legacy unread flag must not control this", IsStaff: false, IsRead: false},
		},
	}, nil)

	require.Equal(t, 3, response["unread_count"])
}
