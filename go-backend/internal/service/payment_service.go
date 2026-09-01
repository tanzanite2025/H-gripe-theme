package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/antifraud"
	"commerce-platform/internal/repository"
)

type PaymentService struct {
	txManager                                 *repository.TxManager
	paymentRepo                               *repository.PaymentRepository
	orderRepo                                 *repository.OrderRepository
	policyDisclosureRepo                      *repository.OrderPolicyDisclosureRepository
	shippingRepo                              *repository.ShippingRepository
	ticketRepo                                *repository.TicketRepository
	risk                                      *antifraud.Service
	stripeDisputeEvidenceSubmitter            stripeDisputeEvidenceSubmitter
	paypalDisputeEvidenceSubmitter            PayPalDisputeEvidenceSubmitter
	paypalDisputeDocumentStorage              PayPalDisputeEvidenceDocumentStorage
	paypalDisputeInvoiceOptions               PayPalDisputeInvoiceOptions
	paypalDisputeInvoiceSellerProfileProvider PayPalDisputeInvoiceSellerProfileProvider
	paypalDisputeCommercialInvoiceRenderer    paypalDisputeCommercialInvoiceRendererFunc
	productCache                              ProductCacheInvalidator
	productCacheEvents                        ProductCacheEventPublisher
}

func (s *PaymentService) ConfigureRisk(orderRepo *repository.OrderRepository, risk *antifraud.Service) {
	s.orderRepo = orderRepo
	s.risk = risk
}

func (s *PaymentService) ConfigureProductCacheInvalidator(invalidator ProductCacheInvalidator) {
	if s == nil {
		return
	}
	s.productCache = invalidator
}

func (s *PaymentService) ConfigureProductCacheEventPublisher(publisher ProductCacheEventPublisher) {
	if s == nil {
		return
	}
	s.productCacheEvents = publisher
}

func (s *PaymentService) ConfigureEvidenceSources(orderRepo *repository.OrderRepository, shippingRepo *repository.ShippingRepository, ticketRepo *repository.TicketRepository) {
	if orderRepo != nil {
		s.orderRepo = orderRepo
	}
	s.shippingRepo = shippingRepo
	s.ticketRepo = ticketRepo
}

func (s *PaymentService) ConfigurePolicyDisclosureRepository(repo *repository.OrderPolicyDisclosureRepository) {
	if s == nil {
		return
	}
	s.policyDisclosureRepo = repo
}

func (s *PaymentService) ConfigurePayPalDisputeEvidenceSubmitter(submitter PayPalDisputeEvidenceSubmitter) {
	if s == nil {
		return
	}
	s.paypalDisputeEvidenceSubmitter = submitter
}

func (s *PaymentService) ConfigurePayPalDisputeEvidenceDocumentStorage(storage PayPalDisputeEvidenceDocumentStorage) {
	if s == nil {
		return
	}
	s.paypalDisputeDocumentStorage = storage
}

func (s *PaymentService) ConfigurePayPalDisputeInvoiceOptions(options PayPalDisputeInvoiceOptions) {
	if s == nil {
		return
	}
	s.paypalDisputeInvoiceOptions = options
}

func (s *PaymentService) ConfigurePayPalDisputeInvoiceSellerProfileProvider(provider PayPalDisputeInvoiceSellerProfileProvider) {
	if s == nil {
		return
	}
	s.paypalDisputeInvoiceSellerProfileProvider = provider
}

type GatewayPaymentAttemptInput struct {
	Provider           string
	OrderNumber        string
	TransactionID      string
	AttemptKey         string
	ProviderRequestKey string
	PaymentMethod      string
	Status             string
	Amount             float64
	Currency           string
	GatewayResponse    string
	ErrorMessage       string
}

type EnsureGatewayPaymentAttemptInput struct {
	Provider           string
	OrderNumber        string
	AttemptKey         string
	ProviderRequestKey string
	PaymentMethod      string
	Amount             float64
	Currency           string
}

// NormalizePaymentAttemptKey scopes a client retry key to one provider and
// order. A missing key keeps legacy callers on the old order-scoped behavior;
// API handlers should always receive a key from the Idempotency-Key middleware.
func NormalizePaymentAttemptKey(provider string, orderID uint, requestKey string) string {
	requestKey = strings.TrimSpace(requestKey)
	if requestKey == "" {
		return fmt.Sprintf("legacy:%s:%d", strings.TrimSpace(provider), orderID)
	}
	return requestKey
}

func PaymentProviderRequestKey(provider string, orderID uint, attemptKey string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		"payment-provider-request-v1",
		strings.TrimSpace(provider),
		fmt.Sprint(orderID),
		strings.TrimSpace(attemptKey),
	}, ":")))
	return "pay-" + hex.EncodeToString(sum[:16])
}

func paymentAttemptTransactionID(provider string, orderID uint, attemptKey string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		"payment-attempt-v1",
		strings.TrimSpace(provider),
		fmt.Sprint(orderID),
		strings.TrimSpace(attemptKey),
	}, ":")))
	return "attempt-" + hex.EncodeToString(sum[:16])
}

func (s *PaymentService) RecordGatewayPaymentFailure(ctx context.Context, provider, orderNumber, transactionID string) error {
	if strings.TrimSpace(orderNumber) == "" {
		return nil
	}
	if strings.TrimSpace(transactionID) != "" {
		if err := s.RecordGatewayPaymentAttempt(GatewayPaymentAttemptInput{
			Provider:      provider,
			OrderNumber:   orderNumber,
			TransactionID: transactionID,
			Status:        "failed",
			ErrorMessage:  "payment intent failed",
		}); err != nil {
			return err
		}
	}

	if s.risk != nil {
		s.risk.RecordProviderFailure(provider)
	}
	return nil
}

func NewPaymentService(txManager *repository.TxManager, paymentRepo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		txManager:   txManager,
		paymentRepo: paymentRepo,
	}
}

func (s *PaymentService) EnsureGatewayPaymentAttempt(input EnsureGatewayPaymentAttemptInput) (*payment.Transaction, error) {
	input.Provider = strings.TrimSpace(input.Provider)
	input.OrderNumber = strings.TrimSpace(input.OrderNumber)
	input.PaymentMethod = strings.TrimSpace(input.PaymentMethod)
	if input.Provider == "" {
		return nil, errors.New("provider is required")
	}
	if input.OrderNumber == "" {
		return nil, errors.New("order_number is required")
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = input.Provider
	}
	input.Currency = normalizePaymentCurrency(input.Currency)

	var attempt *payment.Transaction
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		o, err := repos.Order.FindByOrderNumberForVerification(input.OrderNumber)
		if err != nil {
			return normalizeOrderError(err)
		}
		expectedCurrency, err := orderPaymentCurrency(o)
		if err != nil {
			return err
		}
		if input.Currency == "" {
			input.Currency = expectedCurrency
		} else if input.Currency != expectedCurrency {
			return fmt.Errorf("transaction currency %s does not match order currency %s", input.Currency, expectedCurrency)
		}
		if input.Amount <= 0 {
			input.Amount = o.TotalAmount
		}
		if math.Abs(o.TotalAmount-input.Amount) >= gatewayPaymentAmountTolerance {
			return fmt.Errorf("payment amount %.2f does not match order total %.2f", input.Amount, o.TotalAmount)
		}

		input.AttemptKey = NormalizePaymentAttemptKey(input.Provider, o.ID, input.AttemptKey)
		if input.ProviderRequestKey == "" {
			input.ProviderRequestKey = PaymentProviderRequestKey(input.Provider, o.ID, input.AttemptKey)
		}
		if existing, err := repos.Payment.FindTransactionByAttemptKeyForUpdate(o.ID, input.PaymentMethod, input.AttemptKey); err == nil {
			attempt = existing
			return nil
		} else if !repository.IsRecordNotFound(err) {
			return err
		}

		now := time.Now().UTC()
		attempt = &payment.Transaction{
			OrderID:            o.ID,
			TransactionID:      paymentAttemptTransactionID(input.Provider, o.ID, input.AttemptKey),
			AttemptKey:         input.AttemptKey,
			ProviderRequestKey: input.ProviderRequestKey,
			PaymentMethod:      input.PaymentMethod,
			Amount:             input.Amount,
			Currency:           input.Currency,
			Status:             "pending",
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		created, err := repos.Payment.CreateTransactionIfAbsent(attempt)
		if err != nil {
			return err
		}
		if created {
			return nil
		}

		existing, err := repos.Payment.FindTransactionByAttemptKeyForUpdate(
			o.ID,
			input.PaymentMethod,
			input.AttemptKey,
		)
		if err != nil {
			return err
		}
		attempt = existing
		return nil
	})
	return attempt, err
}

func (s *PaymentService) RecordGatewayPaymentAttempt(input GatewayPaymentAttemptInput) error {
	input.Provider = strings.TrimSpace(input.Provider)
	input.OrderNumber = strings.TrimSpace(input.OrderNumber)
	input.TransactionID = strings.TrimSpace(input.TransactionID)
	input.AttemptKey = strings.TrimSpace(input.AttemptKey)
	input.ProviderRequestKey = strings.TrimSpace(input.ProviderRequestKey)
	if input.Provider == "" {
		return errors.New("provider is required")
	}
	if input.OrderNumber == "" {
		return errors.New("order_number is required")
	}
	if input.TransactionID == "" {
		return errors.New("transaction_id is required")
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = input.Provider
	}
	input.Status = normalizeGatewayAttemptStatus(input.Status)
	input.Currency = normalizePaymentCurrency(input.Currency)

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		o, err := repos.Order.FindByOrderNumberForVerification(input.OrderNumber)
		if err != nil {
			return normalizeOrderError(err)
		}
		expectedCurrency, err := orderPaymentCurrency(o)
		if err != nil {
			return err
		}
		if input.Currency == "" {
			input.Currency = expectedCurrency
		} else if input.Currency != expectedCurrency {
			return fmt.Errorf("transaction currency %s does not match order currency %s", input.Currency, expectedCurrency)
		}
		if input.Amount > 0 && math.Abs(o.TotalAmount-input.Amount) >= gatewayPaymentAmountTolerance {
			return fmt.Errorf("payment amount %.2f does not match order total %.2f", input.Amount, o.TotalAmount)
		}

		var existing *payment.Transaction
		if input.AttemptKey != "" {
			input.AttemptKey = NormalizePaymentAttemptKey(input.Provider, o.ID, input.AttemptKey)
			if attempt, attemptErr := repos.Payment.FindTransactionByAttemptKeyForUpdate(o.ID, input.PaymentMethod, input.AttemptKey); attemptErr == nil {
				existing = attempt
			} else if !repository.IsRecordNotFound(attemptErr) {
				return attemptErr
			}
		}
		if existing == nil {
			if transaction, transactionErr := repos.Payment.FindTransactionByTransactionIDForUpdate(input.TransactionID); transactionErr == nil {
				existing = transaction
			} else if !repository.IsRecordNotFound(transactionErr) {
				return transactionErr
			}
		}
		if existing != nil {
			if existing.Status == "completed" || existing.Status == "refunded" || existing.Status == "expired" {
				return nil
			}
			if input.AttemptKey != "" {
				existing.AttemptKey = input.AttemptKey
			}
			if input.ProviderRequestKey != "" {
				existing.ProviderRequestKey = input.ProviderRequestKey
			}
			existing.TransactionID = input.TransactionID
			existing.OrderID = o.ID
			existing.PaymentMethod = input.PaymentMethod
			if input.Amount > 0 {
				existing.Amount = input.Amount
			} else if existing.Amount <= 0 {
				existing.Amount = o.TotalAmount
			}
			existing.Currency = input.Currency
			existing.Status = input.Status
			existing.GatewayResponse = input.GatewayResponse
			existing.ErrorMessage = input.ErrorMessage
			return repos.Payment.UpdateTransaction(existing)
		}

		amount := input.Amount
		if amount <= 0 {
			amount = o.TotalAmount
		}
		return repos.Payment.CreateTransaction(&payment.Transaction{
			OrderID:            o.ID,
			TransactionID:      input.TransactionID,
			AttemptKey:         input.AttemptKey,
			ProviderRequestKey: input.ProviderRequestKey,
			PaymentMethod:      input.PaymentMethod,
			Amount:             amount,
			Currency:           input.Currency,
			Status:             input.Status,
			GatewayResponse:    input.GatewayResponse,
			ErrorMessage:       input.ErrorMessage,
		})
	})
}
