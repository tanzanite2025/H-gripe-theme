package service

import "errors"

func (s *PaymentService) ClaimStripeWebhookEvent(eventID, eventType, payload string) (bool, error) {
	if s.paymentRepo == nil {
		return false, errors.New("payment repository is unavailable")
	}
	return s.paymentRepo.ClaimStripeWebhookEvent(eventID, eventType, payload)
}

func (s *PaymentService) MarkStripeWebhookEventProcessed(eventID string) error {
	return s.paymentRepo.MarkStripeWebhookEventProcessed(eventID)
}

func (s *PaymentService) MarkStripeWebhookEventFailed(eventID string, processingErr error) error {
	return s.paymentRepo.MarkStripeWebhookEventFailed(eventID, processingErr)
}
