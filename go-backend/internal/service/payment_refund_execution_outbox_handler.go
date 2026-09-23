package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/resilience"
	"commerce-platform/internal/repository"
)

// PaymentRefundExecutionOutboxHandler executes a refund only after the local
// request transaction has committed. Credentials are resolved at dispatch time
// and never appear in the durable event payload.
type PaymentRefundExecutionOutboxHandler struct {
	paymentService  *PaymentService
	settingsService *AdminSettingsService
	newGateway      func(*pgateway.Config) (pgateway.PaymentGateway, error)
}

func NewPaymentRefundExecutionOutboxHandler(
	paymentService *PaymentService,
	settingsService *AdminSettingsService,
	newGateway ...func(*pgateway.Config) (pgateway.PaymentGateway, error),
) *PaymentRefundExecutionOutboxHandler {
	var gatewayFactory func(*pgateway.Config) (pgateway.PaymentGateway, error)
	if len(newGateway) > 0 {
		gatewayFactory = newGateway[0]
	}
	return &PaymentRefundExecutionOutboxHandler{
		paymentService:  paymentService,
		settingsService: settingsService,
		newGateway:      gatewayFactory,
	}
}

func (h *PaymentRefundExecutionOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.paymentService == nil {
		return errors.New("payment refund execution service is not configured")
	}
	var payload outbox.PaymentRefundExecutionRequestedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode payment refund execution request: %w", err)
	}
	if payload.RefundID == 0 {
		return errors.New("payment refund execution request refund id is required")
	}
	provider, err := pgateway.ParseGatewayType(payload.Provider)
	if err != nil {
		return err
	}
	plan, err := h.paymentService.loadPendingRefundExecutionPlan(payload.RefundID, payload.Attempt, string(provider))
	if err != nil {
		return err
	}
	// A newer attempt, an already completed execution, or a reconciled webhook
	// makes this older command a harmless no-op.
	if plan == nil {
		return nil
	}

	refundMoney, err := plan.Refund.AmountMoney()
	if err != nil {
		return err
	}
	var gateway pgateway.PaymentGateway
	if refundMoney.AmountMinor() > 0 {
		if h.newGateway != nil {
			gateway, err = h.newGateway(&pgateway.Config{Type: provider})
		} else {
			var config *pgateway.Config
			config, err = refundExecutionGatewayConfig(h.settingsService, provider)
			if err == nil {
				gateway, err = pgateway.NewPaymentGateway(config)
			}
		}
		if err != nil {
			_, failErr := h.paymentService.failPendingRefundExecution(plan.Execution.RefundID, err.Error())
			if failErr != nil {
				return errors.Join(err, failErr)
			}
			return nil
		}
	}

	_, _, err = h.paymentService.executePendingRefundPlan(ctx, plan, gateway)
	if err != nil {
		// The external provider may have accepted the request even though the
		// transport failed. Outbox must mark this event unknown and wait for
		// reconciliation instead of issuing a blind retry or local failure.
		if errors.Is(err, resilience.ErrExternalOutcomeUnknown) {
			return err
		}
		return err
	}
	return nil
}

func (s *PaymentService) loadPendingRefundExecutionPlan(
	refundID uint,
	attempt int,
	provider string,
) (*pendingRefundExecutionPlan, error) {
	var plan *pendingRefundExecutionPlan
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.RefundExecution == nil {
			return errors.New("payment refund execution repository is not configured")
		}
		refund, err := repos.Payment.FindRefundByIDForUpdate(refundID)
		if err != nil {
			return err
		}
		execution, err := repos.RefundExecution.FindByRefundIDForUpdate(refundID)
		if err != nil {
			return err
		}
		if execution.Status == paymentdomain.PaymentRefundExecutionStatusSucceeded || refund.Status == "completed" {
			return nil
		}
		if execution.Status != paymentdomain.PaymentRefundExecutionStatusProcessing || refund.Status != "pending" {
			return nil
		}
		if attempt > 0 && execution.AttemptCount != attempt {
			return nil
		}
		if provider != "" && !strings.EqualFold(strings.TrimSpace(execution.Provider), strings.TrimSpace(provider)) {
			return fmt.Errorf("refund provider %s does not match execution provider %s", provider, execution.Provider)
		}
		transaction, err := repos.Payment.FindTransactionByIDForUpdate(refund.TransactionID)
		if err != nil {
			return err
		}
		orderRecord, err := repos.Order.FindByIDForUpdate(refund.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		plan = &pendingRefundExecutionPlan{
			Refund:      refund,
			Transaction: transaction,
			Order:       orderRecord,
			Execution:   execution,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func refundExecutionGatewayConfig(settingsService *AdminSettingsService, provider pgateway.GatewayType) (*pgateway.Config, error) {
	if settingsService != nil {
		stored, found, err := readRefundExecutionSecureGatewayConfig(settingsService, provider)
		if err != nil {
			return nil, err
		}
		if found {
			config := pgateway.GatewayConfigFromSecureConfig(stored)
			if strings.TrimSpace(config.APIKey) == "" {
				return nil, errors.New(string(provider) + " API key is not configured")
			}
			return config, nil
		}
	}
	config := pgateway.LoadConfigFromEnv(provider)
	if strings.TrimSpace(config.APIKey) == "" {
		return nil, errors.New(string(provider) + " API key is not configured")
	}
	return config, nil
}

func readRefundExecutionSecureGatewayConfig(
	settingsService *AdminSettingsService,
	provider pgateway.GatewayType,
) (pgateway.SecureGatewayConfig, bool, error) {
	setting, err := settingsService.GetDomainManagedSetting(pgateway.SecureGatewaySettingKey(provider), "global")
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return pgateway.SecureGatewayConfig{}, false, nil
		}
		return pgateway.SecureGatewayConfig{}, false, err
	}
	config, err := pgateway.DecodeStoredSecureGatewayConfig(setting.Value, provider)
	if err != nil {
		return pgateway.SecureGatewayConfig{}, true, err
	}
	return config, true, nil
}
