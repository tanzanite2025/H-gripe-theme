package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/antifraud"
	"commerce-platform/internal/repository"
)

type PaymentService struct {
	txManager                                 *repository.TxManager
	paymentRepo                               *repository.PaymentRepository
	orderRepo                                 *repository.OrderRepository
	policyDisclosureRepo                      *repository.OrderPolicyDisclosureRepository
	ticketRepo                                *repository.TicketRepository
	orderEvidenceAssembler                    *OrderEvidencePackageAssembler
	orderEvidenceSubmissionRepo               *repository.OrderEvidenceSubmissionSnapshotRepository
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

func (s *PaymentService) ConfigureEvidenceSources(
	orderRepo *repository.OrderRepository,
	ticketRepo *repository.TicketRepository,
) {
	if orderRepo != nil {
		s.orderRepo = orderRepo
	}
	s.ticketRepo = ticketRepo
}

func (s *PaymentService) ConfigureOrderEvidenceAssembler(
	assembler *OrderEvidencePackageAssembler,
) {
	if s == nil {
		return
	}
	s.orderEvidenceAssembler = assembler
}

func (s *PaymentService) ConfigureOrderEvidenceSubmissionSnapshotRepository(
	repo *repository.OrderEvidenceSubmissionSnapshotRepository,
) {
	if s == nil {
		return
	}
	s.orderEvidenceSubmissionRepo = repo
}

func (s *PaymentService) assembleOrderEvidencePackage(
	orderID uint,
) (*OrderEvidencePackageAssembly, error) {
	if s == nil || s.orderEvidenceAssembler == nil {
		return nil, nil
	}
	return s.orderEvidenceAssembler.Assemble(orderID)
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
	Amount             domainmoney.Money
	GatewayResponse    string
	ErrorMessage       string
}

type EnsureGatewayPaymentAttemptInput struct {
	Provider           string
	OrderNumber        string
	AttemptKey         string
	ProviderRequestKey string
	PaymentMethod      string
	Amount             domainmoney.Money
}

// NormalizePaymentAttemptKey accepts only the client-provided retry key.
// Payment-facing routes require this value through idempotency middleware.
func NormalizePaymentAttemptKey(requestKey string) string {
	return strings.TrimSpace(requestKey)
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
	var attempt *payment.Transaction
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		o, err := repos.Order.FindByOrderNumberForVerification(input.OrderNumber)
		if err != nil {
			return normalizeOrderError(err)
		}
		expectedSettlement, err := orderPaymentSettlement(o)
		if err != nil {
			return err
		}
		expectedCurrency := expectedSettlement.Currency().String()
		if input.Amount.Currency().String() == "" {
			return errors.New("payment attempt amount is required")
		}
		if input.Amount.Currency().String() != expectedCurrency {
			return fmt.Errorf("payment amount currency %s does not match order currency %s", input.Amount.Currency().String(), expectedCurrency)
		}
		if input.Amount.AmountMinor() <= 0 {
			return errors.New("payment attempt amount must be greater than zero")
		}
		if input.Amount.AmountMinor() != expectedSettlement.AmountMinor() {
			actualAmount, _ := input.Amount.MajorFloat()
			expectedAmount, _ := expectedSettlement.MajorFloat()
			return fmt.Errorf("payment amount %.2f does not match payable amount %.2f", actualAmount, expectedAmount)
		}
		inputAmount, amountErr := input.Amount.MajorFloat()
		if amountErr != nil {
			return amountErr
		}

		input.AttemptKey = NormalizePaymentAttemptKey(input.AttemptKey)
		if input.AttemptKey == "" {
			return errors.New("payment attempt key is required")
		}
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
			AmountMinor:        input.Amount.AmountMinor(),
			Amount:             inputAmount,
			Currency:           expectedCurrency,
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
	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		o, err := repos.Order.FindByOrderNumberForVerification(input.OrderNumber)
		if err != nil {
			return normalizeOrderError(err)
		}
		expectedSettlement, err := orderPaymentSettlement(o)
		if err != nil {
			return err
		}
		expectedCurrency := expectedSettlement.Currency().String()
		expectedAmount, err := expectedSettlement.MajorFloat()
		if err != nil {
			return err
		}
		amountMoney := input.Amount
		if amountMoney.Currency().String() == "" || amountMoney.AmountMinor() <= 0 {
			amountMoney = expectedSettlement
		} else if amountMoney.Currency().String() != expectedCurrency {
			return fmt.Errorf("payment amount currency %s does not match order currency %s", amountMoney.Currency().String(), expectedCurrency)
		}
		if amountMoney.AmountMinor() != expectedSettlement.AmountMinor() {
			actualAmount, _ := amountMoney.MajorFloat()
			return fmt.Errorf("payment amount %.2f does not match payable amount %.2f", actualAmount, expectedAmount)
		}
		amount, amountErr := amountMoney.MajorFloat()
		if amountErr != nil {
			return amountErr
		}

		var existing *payment.Transaction
		if input.AttemptKey != "" {
			input.AttemptKey = NormalizePaymentAttemptKey(input.AttemptKey)
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
			if existing.Status == "completed" ||
				existing.Status == payment.TransactionStatusDuplicatePaid ||
				existing.Status == "refunded" ||
				existing.Status == "expired" {
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
			existing.AmountMinor = amountMoney.AmountMinor()
			existing.Amount = amount
			existing.Currency = expectedCurrency
			existing.Status = input.Status
			existing.GatewayResponse = input.GatewayResponse
			existing.ErrorMessage = input.ErrorMessage
			return repos.Payment.UpdateTransaction(existing)
		}

		return repos.Payment.CreateTransaction(&payment.Transaction{
			OrderID:            o.ID,
			TransactionID:      input.TransactionID,
			AttemptKey:         input.AttemptKey,
			ProviderRequestKey: input.ProviderRequestKey,
			PaymentMethod:      input.PaymentMethod,
			AmountMinor:        amountMoney.AmountMinor(),
			Amount:             amount,
			Currency:           expectedCurrency,
			Status:             input.Status,
			GatewayResponse:    input.GatewayResponse,
			ErrorMessage:       input.ErrorMessage,
		})
	})
}
