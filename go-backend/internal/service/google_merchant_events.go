package service

import "fmt"

func (s *GoogleMerchantService) enqueueMerchantOfferRevalidate(offerID uint, reason string) error {
	if s == nil || s.merchantEvents == nil || offerID == 0 {
		return nil
	}
	return s.merchantEvents.EnqueueOfferRevalidate(offerID, reason)
}

func requireMerchantOfferEvent(err error, offerID uint, reason string) error {
	if err == nil {
		return nil
	}
	if offerID == 0 {
		return err
	}
	return fmt.Errorf("Merchant offer %d saved but revalidation event %q was not queued: %w", offerID, reason, err)
}
