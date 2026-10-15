package repository

import (
	"context"

	"commerce-platform/internal/domain/ticket"
)

func (r *TicketRepository) FindCustomerServiceAttachmentMessageByConversationAndID(ctx context.Context, ticketID, messageID uint, includeInternal bool) (*ticket.TicketMessage, error) {
	query := r.db.WithContext(ctx).Where("ticket_id = ? AND id = ?", ticketID, messageID)
	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}
	var message ticket.TicketMessage
	if err := query.First(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}
