package service

import (
	"errors"
	"fmt"

	"commerce-platform/internal/repository"
)

var ErrCustomerServiceMessageMediaReferenceInvalid = repository.ErrTicketStructuredMediaReferenceInvalid

func (s *TicketService) ConfigureMediaService(assets *MediaService) {
	if s != nil && s.ticketRepo != nil {
		s.ticketRepo.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
	}
}

func mapAutoReplyMediaRepositoryWriteError(err error) error {
	if errors.Is(err, repository.ErrAutoReplyRuleMediaReferenceInvalid) || errors.Is(err, repository.ErrTicketStructuredMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrInvalidAutoReplyRule, err)
	}
	return err
}
