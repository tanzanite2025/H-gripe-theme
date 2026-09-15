package service

import (
	"errors"

	"commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/email"
	"commerce-platform/internal/pkg/ordernumber"
	"commerce-platform/internal/repository"
)

type OrderService struct {
	txManager                *repository.TxManager
	orderRepo                *repository.OrderRepository
	checkout                 *CheckoutService
	shipping                 *ShippingService
	payment                  *PaymentService
	emailSender              email.EmailService
	numberGenerator          *ordernumber.Generator
	productCache             ProductCacheInvalidator
	productCacheEvents       ProductCacheEventPublisher
	refundCancellationPolicy *RefundCancellationPolicyService
	orderEvidenceSnapshot    *OrderEvidenceSnapshotService
	orderEvidence            *OrderEvidenceService
}

var (
	ErrOrderNotFound                                 = errors.New("order not found")
	ErrOrderHideNotAllowed                           = errors.New("only orders in cancelled/unpaid or payment_expired/expired state can be hidden; orders with payment activity must be retained")
	ErrOrderHideTransactionRequired                  = errors.New("hiding an order requires a transaction manager")
	ErrPaidOrderCancellationNotAllowed               = errors.New("paid orders cannot be cancelled directly; please submit an after-sales refund request or contact support")
	ErrOrderCancellationConflict                     = errors.New("order was already cancelled or is no longer eligible for cancellation")
	ErrOrderStatusConflict                           = repository.ErrOrderStatusConflict
	ErrSystemManagedOrderStatus                      = errors.New("order status is managed by payment workflow")
	ErrOrderFulfillmentNotAllowed                    = errors.New("only paid, processing, or already shipped orders can be fulfilled")
	ErrOrderFulfillmentPaymentRequired               = errors.New("only paid orders can be fulfilled")
	ErrOrderFulfillmentOnHold                        = errors.New("order fulfillment is blocked while payment dispute or review is active")
	ErrOrderFulfillmentTransactionNeeded             = errors.New("order fulfillment transaction is not configured")
	ErrOrderCustomsUpdateTransactionNeeded           = errors.New("order customs update transaction is not configured")
	ErrOrderFulfillmentSignatureConfirmationRequired = errors.New("signature confirmation is required before fulfilling this order")
	ErrOrderFulfillmentStatusManaged                 = errors.New("shipped status is managed by the fulfillment workflow")
	ErrOrderCustomsUpdateLocked                      = errors.New("order customs declaration is locked after shipment")
	ErrOrderProductionNotRequired                    = errors.New("order does not require production")
	ErrOrderProductionPaymentRequired                = errors.New("only paid orders can enter production")
	ErrOrderProductionNotAllowed                     = errors.New("order is not eligible for production workflow")
	ErrOrderProductionAlreadyStarted                 = errors.New("order production has already started")
	ErrOrderProductionNotStarted                     = errors.New("order production has not started")
	ErrOrderProductionNotCompleted                   = errors.New("order production must be completed before fulfillment")
	ErrOrderProductionTransactionNeeded              = errors.New("order production transaction is not configured")
	ErrProductionStartedCancellationNotAllowed       = errors.New("custom orders cannot be cancelled after production has started")
	ErrTrackingNumberRequired                        = errors.New("tracking number is required")
	ErrOrderShippingNotConfigured                    = errors.New("order shipping service is not configured")
	ErrOrderNumberNotConfigured                      = errors.New("order number generator is not configured")
	ErrOrderItemNotFound                             = errors.New("order item not found")
	ErrOrderIdempotencyUnavailable                   = errors.New("order idempotency is not configured")
	ErrOrderIdempotencyConflict                      = errors.New("idempotency key was already used for a different order request")
	ErrOrderIdempotencyInProgress                    = errors.New("idempotent order request is already being processed")
	ErrOrderIdempotencyHashRequired                  = errors.New("idempotency request hash is required")
	ErrOrderFulfillmentIdempotencyUnavailable        = errors.New("order fulfillment idempotency is not configured")
	ErrOrderFulfillmentIdempotencyConflict           = errors.New("idempotency key was already used for a different fulfillment request")
	ErrOrderFulfillmentIdempotencyInProgress         = errors.New("idempotent fulfillment request is already being processed")
	ErrOrderEvidenceNotConfigured                    = errors.New("order evidence services are not configured")
	ErrDeclaredValueInvalid                          = errors.New("declared value must be a finite non-negative number")
	ErrDeclaredValueConfirmationRequired             = errors.New("declared value is required when confirming")
	ErrOrderCustomsDeclarationIncomplete             = errors.New("customs declared value must be a positive confirmed value for every order item")
)

func NewOrderService(
	txManager *repository.TxManager,
	orderRepo *repository.OrderRepository,
	checkout *CheckoutService,
	shipping *ShippingService,
	numberGenerators ...*ordernumber.Generator,
) *OrderService {
	var numberGenerator *ordernumber.Generator
	if len(numberGenerators) > 0 {
		numberGenerator = numberGenerators[0]
	}
	service := &OrderService{
		txManager:       txManager,
		orderRepo:       orderRepo,
		checkout:        checkout,
		shipping:        shipping,
		numberGenerator: numberGenerator,
	}
	return service
}

func (s *OrderService) ConfigureRefundCancellationPolicy(policy *RefundCancellationPolicyService) {
	if s == nil {
		return
	}
	s.refundCancellationPolicy = policy
}

func (s *OrderService) ConfigureOrderEvidenceSnapshot(snapshotService *OrderEvidenceSnapshotService) {
	if s == nil {
		return
	}
	s.orderEvidenceSnapshot = snapshotService
}

func (s *OrderService) ConfigureOrderEvidence(evidenceService *OrderEvidenceService) {
	if s == nil {
		return
	}
	s.orderEvidence = evidenceService
}

func (s *OrderService) ConfigureProductCacheInvalidator(invalidator ProductCacheInvalidator) {
	if s == nil {
		return
	}
	s.productCache = invalidator
}

func (s *OrderService) ConfigureProductCacheEventPublisher(publisher ProductCacheEventPublisher) {
	if s == nil {
		return
	}
	s.productCacheEvents = publisher
}

func (s *OrderService) ConfigurePaymentDisputeAnalysis(payment *PaymentService) {
	if s == nil {
		return
	}
	s.payment = payment
}

func (s *OrderService) ConfigureAdminEmailSender(sender email.EmailService) {
	if s == nil {
		return
	}
	s.emailSender = sender
}

type OrderTrackingUpdateInput struct {
	TrackingNumber     string
	TrackingProviderID uint
	CarrierID          *uint
	CarrierServiceID   *uint
	SignatureConfirmed bool
}

type OrderFulfillmentResult struct {
	Order                     *order.Order
	TrackingShipment          *shippingdomain.TrackingShipment
	TrackingRegistrationError string
}

func (s *OrderService) GetOrder(id uint, userID uint) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, normalizeOrderError(err)
	}

	if o.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return o, nil
}

func (s *OrderService) GetOrderByNumber(orderNumber string, userID uint) (*order.Order, error) {
	if !s.validatesKnownProtectedOrderNumber(orderNumber) {
		return nil, ErrOrderNotFound
	}
	o, err := s.orderRepo.FindByOrderNumber(orderNumber)
	if err != nil {
		return nil, normalizeOrderError(err)
	}

	if o.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return o, nil
}

// GetOrderByNumberForPayment is intentionally limited to internal payment
// handlers and does not expose a public order-number lookup endpoint.
func (s *OrderService) GetOrderByNumberForPayment(orderNumber string) (*order.Order, error) {
	if !s.validatesKnownProtectedOrderNumber(orderNumber) {
		return nil, ErrOrderNotFound
	}
	o, err := s.orderRepo.FindByOrderNumberForVerification(orderNumber)
	if err != nil {
		return nil, normalizeOrderError(err)
	}
	return o, nil
}

func (s *OrderService) validatesKnownProtectedOrderNumber(value string) bool {
	if ordernumber.IsSequentialCandidate(value) {
		return false
	}
	if !ordernumber.IsProtectedFormat(value) {
		return true
	}
	return s != nil && s.numberGenerator != nil && s.numberGenerator.Validate(value)
}

func (s *OrderService) GetUserOrders(userID uint, page, pageSize int) ([]order.Order, int64, error) {
	return s.orderRepo.FindByUserID(userID, page, pageSize)
}

func (s *OrderService) findOrder(id uint) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, normalizeOrderError(err)
	}

	return o, nil
}

func (s *OrderService) ProtectedOrderNumberStatus() map[string]interface{} {
	configured := s != nil && s.numberGenerator != nil
	return map[string]interface{}{
		"configured": configured,
		"format":     "TZ-YYYY-<16 char HMAC payload><4 char HMAC checksum>",
		"public_id_policy": []string{
			"Public order numbers are generated by the backend only.",
			"Database auto-increment IDs are not embedded in public order numbers.",
			"Payment gateway merchant order IDs use the public order number.",
		},
		"internal_id_policy": []string{
			"Admin-only routes may still use numeric database IDs after authentication.",
			"Customer-facing routes and chat payloads use order_number only.",
		},
		"verification": "HMAC checksum validation is available for generated order numbers.",
		"key_rotation": "ORDER_NUMBER_PREVIOUS_SECRET may verify old numbers during one rotation window, while new numbers use ORDER_NUMBER_SECRET only.",
	}
}

func normalizeOrderError(err error) error {
	if repository.IsRecordNotFound(err) {
		return ErrOrderNotFound
	}
	return err
}

func (s *OrderService) GetOrderStats(userID uint) (map[string]int64, error) {
	return s.orderRepo.GetOrderStats(userID)
}
