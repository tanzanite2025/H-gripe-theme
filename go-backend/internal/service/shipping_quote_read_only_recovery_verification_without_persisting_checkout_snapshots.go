package service

import "commerce-platform/internal/repository"

// NewShippingServiceForReadOnlyRecoveryVerification uses the normal catalog,
// packaging and rate calculations without creating checkout quote identities or
// snapshots. Only the stopped recovery HTTP adapter should use this constructor;
// callers must also enforce a read-only SQL transaction and restrict routes.
func NewShippingServiceForReadOnlyRecoveryVerification(shippingRepository *repository.ShippingRepository, productRepository *repository.ProductRepository) *ShippingService {
	return &ShippingService{shippingRepo: shippingRepository, productRepo: productRepository, readOnlyRecoveryVerification: true}
}
