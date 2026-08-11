package service

import (
	"fmt"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/auth"
)

func (s *TicketService) validateAutoReplyAgentID(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	agentID, err := strconv.ParseUint(value, 10, 32)
	if err != nil || agentID == 0 {
		return fmt.Errorf("%w: agent_id must be an active customer-service user ID", ErrInvalidAutoReplyRule)
	}

	if s.userRepo == nil {
		return fmt.Errorf("%w: customer-service user repository is unavailable", ErrInvalidAutoReplyRule)
	}
	agent, err := s.userRepo.FindByID(uint(agentID))
	if err != nil || agent == nil ||
		strings.TrimSpace(agent.Status) != "active" ||
		!auth.IsCustomerServiceAgentRole(agent.Role) {
		return fmt.Errorf("%w: agent_id must reference an active customer-service user", ErrInvalidAutoReplyRule)
	}
	return nil
}
