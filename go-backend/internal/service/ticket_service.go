package service

import (
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
	"encoding/json"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"strconv"
	"time"
)

const customerServiceTicketCategory = "customer_service"

type TicketService struct {
	ticketRepo                    *repository.TicketRepository
	userRepo                      *repository.UserRepository
	faqRepo                       *repository.FAQRepository
	customerServiceRealtimeOutbox *repository.OutboxRepository
}

// ConfigureCustomerServiceRealtimeOutbox makes public/staff message writes
// transactionally emit a durable realtime event. Application dependencies and
// service tests that exercise customer-service messages must configure it.
func (s *TicketService) ConfigureCustomerServiceRealtimeOutbox(repo *repository.OutboxRepository) {
	if s == nil {
		return
	}
	s.customerServiceRealtimeOutbox = repo
}

func NewTicketService(ticketRepo *repository.TicketRepository, userRepo *repository.UserRepository, faqRepos ...*repository.FAQRepository) *TicketService {
	service := &TicketService{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
	if len(faqRepos) > 0 {
		service.faqRepo = faqRepos[0]
	}
	return service
}

func (s *TicketService) createTicket(t *ticket.Ticket) error {
	t.Status = "open"
	t.Priority = "medium"
	if s.customerServiceRealtimeOutbox == nil || t.Category != customerServiceTicketCategory {
		return s.ticketRepo.CreateTicket(t)
	}
	return s.ticketRepo.WithinTx(func(ticketRepo *repository.TicketRepository, tx *gorm.DB) error {
		if err := ticketRepo.CreateTicket(t); err != nil {
			return err
		}
		conversationID := ""
		if t.ConversationID != nil {
			conversationID = *t.ConversationID
		}
		event := NewCustomerServiceRealtimeEventWithIDAndAudience(
			CustomerServiceConversationCreatedEventID(t.ID),
			CustomerServiceEventConversationCreated,
			t.ID,
			conversationID,
			CustomerServiceRealtimeActor{Kind: "system"},
			CustomerServiceRealtimeAudienceBoth,
			CustomerServiceConversationCreatedPayload{Status: t.Status, AssignedTo: t.AssignedTo},
		)
		payload, err := json.Marshal(event.Payload)
		if err != nil {
			return err
		}
		return s.customerServiceRealtimeOutbox.WithTx(tx).CreateEvent(&outbox.Event{
			EventKey: event.EventID, EventType: outbox.EventTypeCustomerServiceRealtime,
			AggregateType: outbox.AggregateTypeCustomerServiceConversation,
			AggregateID:   strconv.FormatUint(uint64(t.ID), 10), Payload: datatypes.JSON(payload), AvailableAt: time.Now().UTC(),
		})
	})
}

func (s *TicketService) assignTicket(id, assignedTo uint) error {
	return s.ticketRepo.AssignTicket(id, assignedTo)
}

func (s *TicketService) updateTicketStatus(id uint, status string) error {
	return s.ticketRepo.UpdateTicketStatus(id, status)
}
