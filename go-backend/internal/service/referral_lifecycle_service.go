package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

var ErrReferralLifecycleEventInvalid = errors.New("invalid referral lifecycle event")

// ReferralLifecycleScanResult describes one bounded lifecycle sweep, including
// the number of matured records settled through the idempotent reward path.
type ReferralLifecycleScanResult struct {
	// PendingScanned and Expired are retained for backwards-compatible worker
	// metrics. Bound referral relationships are permanent, so the current
	// scanner intentionally leaves both counters at zero.
	PendingScanned           int
	Expired                  int
	OrderedScanned           int
	AdvancedToVesting        int
	MaturedVestingCandidates int
	Settled                  int
}

// ScanLifecycle advances lifecycle states and settles matured vesting records.
// Every mutation is reloaded under a row lock and uses record_version CAS;
// reward issuance is idempotent and runs in the same transaction as settlement.
func (s *ReferralService) ScanLifecycle(ctx context.Context, now time.Time, batchLimit int) (ReferralLifecycleScanResult, error) {
	result := ReferralLifecycleScanResult{}
	if s == nil || s.repo == nil || s.txManager == nil {
		return result, ErrReferralServiceUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if now.IsZero() {
		now = s.now().UTC()
	} else {
		now = now.UTC()
	}
	if batchLimit <= 0 || batchLimit > 500 {
		batchLimit = 100
	}

	// Attribution TTL applies only while a visitor has an unbound signed
	// cookie. Once the account has accepted a referral code, the relationship
	// is permanent and must not be expired by a background sweep. In
	// particular, an account's registration points cannot be forfeited because
	// the user has not placed an order within the cookie TTL. Keep the legacy
	// ListPendingExpired repository method for historical data/reporting, but do
	// not run a destructive expiration pass for newly bound records.

	ordered, err := s.repo.ListOrderedCandidates(now, batchLimit)
	if err != nil {
		return result, err
	}
	if len(ordered) > batchLimit {
		ordered = ordered[:batchLimit]
	}
	result.OrderedScanned = len(ordered)
	for index := range ordered {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		record := ordered[index]
		err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			if repos.Referral == nil || repos.Order == nil || repos.ReferralProgram == nil || record.OrderID == nil {
				return nil
			}
			// Payment/refund handlers lock the order before the referral record;
			// retain that order here to avoid cross-event deadlocks.
			orderRecord, findErr := repos.Order.FindByIDForUpdate(*record.OrderID)
			if repository.IsRecordNotFound(findErr) {
				return nil
			}
			if findErr != nil {
				return findErr
			}
			locked, findErr := repos.Referral.FindRecordByIDForUpdate(record.ID)
			if repository.IsRecordNotFound(findErr) {
				return nil
			}
			if findErr != nil {
				return findErr
			}
			if locked.Status != loyalty.ReferralStatusOrdered || locked.OrderID == nil {
				return nil
			}
			if *locked.OrderID != orderRecord.ID {
				return nil
			}
			config, findErr := repos.ReferralProgram.FindByID(locked.ProgramConfigID)
			if findErr != nil {
				return findErr
			}

			if orderRecord.DeliveredAt != nil {
				deliveredAt := orderRecord.DeliveredAt.UTC()
				if deliveredAt.After(now) {
					return nil
				}
				if err := transitionReferralRecord(
					repos,
					locked,
					loyalty.ReferralStatusVesting,
					"lifecycle_scan",
					"authoritative delivery timestamp observed",
					fmt.Sprintf("referral.lifecycle.delivered:%d", locked.ID),
					map[string]any{
						"delivered_at":  deliveredAt,
						"vesting_until": deliveredAt.AddDate(0, 0, config.VestingPeriodDays),
					},
				); err != nil {
					return err
				}
				result.AdvancedToVesting++
				return nil
			}

			if orderRecord.ShippedAt == nil {
				return nil
			}
			fallbackAt := orderRecord.ShippedAt.UTC().AddDate(0, 0, config.UndeliveredFallbackDays)
			if fallbackAt.After(now) {
				return nil
			}
			// Do not populate delivered_at: this is an operational fallback, not
			// proof that the parcel was delivered. The vesting deadline is the
			// configured maximum wait from shipment and remains auditable.
			if err := transitionReferralRecord(
				repos,
				locked,
				loyalty.ReferralStatusVesting,
				"lifecycle_scan",
				"undelivered fallback deadline reached",
				fmt.Sprintf("referral.lifecycle.fallback:%d", locked.ID),
				map[string]any{"vesting_until": fallbackAt},
			); err != nil {
				return err
			}
			result.AdvancedToVesting++
			return nil
		})
		if err != nil {
			return result, err
		}
	}

	matured, err := s.repo.ListMaturedVesting(now, batchLimit)
	if err != nil {
		return result, err
	}
	result.MaturedVestingCandidates = len(matured)
	for _, record := range matured {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if _, err := s.SettleReferral(record.ID, "vesting period completed", nil); err != nil {
			if errors.Is(err, ErrReferralActionInvalid) {
				continue
			}
			return result, err
		}
		result.Settled++
	}
	return result, nil
}

func (s *ReferralService) HandleOrderPaidOutbox(ctx context.Context, event outbox.Event) error {
	if err := validateReferralLifecycleHandler(s, ctx, event, outbox.EventTypeReferralOrderPaid); err != nil {
		return err
	}
	var payload outbox.ReferralOrderPaidPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("%w: decode order paid payload: %v", ErrReferralLifecycleEventInvalid, err)
	}
	if payload.OrderID == 0 || payload.UserID == 0 || payload.AmountMinor < 0 ||
		strings.TrimSpace(payload.Currency) == "" || payload.PaidAt.IsZero() {
		return ErrReferralLifecycleEventInvalid
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Referral == nil || repos.ReferralProgram == nil || repos.Order == nil {
			return ErrReferralServiceUnavailable
		}
		orderRecord, err := repos.Order.FindByIDForUpdate(payload.OrderID)
		if err != nil {
			return err
		}
		if orderRecord.UserID != payload.UserID {
			return fmt.Errorf("%w: order user mismatch", ErrReferralLifecycleEventInvalid)
		}
		record, err := repos.Referral.FindRecordByRefereeIDForUpdate(payload.UserID)
		if repository.IsRecordNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if record.Status != loyalty.ReferralStatusPending {
			return nil
		}

		config, err := repos.ReferralProgram.FindByID(record.ProgramConfigID)
		if err != nil {
			return err
		}
		signalUpdates, err := s.referralRiskSignalUpdates(repos.Order, repos.Referral, repos.Payment, record, orderRecord)
		if err != nil {
			return err
		}
		commonUpdates := map[string]any{
			"order_id":           payload.OrderID,
			"order_amount_minor": payload.AmountMinor,
			"currency":           strings.ToUpper(strings.TrimSpace(payload.Currency)),
		}
		for key, value := range signalUpdates {
			commonUpdates[key] = value
		}
		paidAt := payload.PaidAt.UTC()
		if config.AntiFraudMode == loyalty.ReferralFraudModeStrict && referralRiskFlagsPresent(signalUpdates["risk_flags"]) {
			commonUpdates["revoked_at"] = paidAt
			commonUpdates["revoke_reason"] = "strict anti-fraud match"
			return transitionReferralRecord(repos, record, loyalty.ReferralStatusRevoked, "order_paid", "strict anti-fraud match", event.EventKey, commonUpdates)
		}
		// The signed cookie TTL was consumed when this account bound the code.
		// A bound relationship does not expire, so an order paid months later is
		// still evaluated by the normal first-order and anti-fraud rules below.
		if orderRecord.PaidAt == nil || !strings.EqualFold(orderRecord.PaymentStatus, "paid") {
			commonUpdates["revoked_at"] = paidAt
			commonUpdates["revoke_reason"] = "order is no longer paid"
			return transitionReferralRecord(repos, record, loyalty.ReferralStatusRevoked, "order_paid", "order is no longer paid", event.EventKey, commonUpdates)
		}
		priorPaidOrders, err := repos.Order.CountEverPaidOrdersForUserBefore(payload.UserID, payload.OrderID)
		if err != nil {
			return err
		}
		if priorPaidOrders > 0 {
			commonUpdates["revoked_at"] = paidAt
			commonUpdates["revoke_reason"] = "referee is not a first-time purchaser"
			return transitionReferralRecord(repos, record, loyalty.ReferralStatusRevoked, "order_paid", "referee is not a first-time purchaser", event.EventKey, commonUpdates)
		}
		if !strings.EqualFold(payload.Currency, config.Currency) || payload.AmountMinor < config.MinOrderAmountMinor {
			commonUpdates["revoked_at"] = paidAt
			commonUpdates["revoke_reason"] = "first order does not meet referral threshold"
			return transitionReferralRecord(repos, record, loyalty.ReferralStatusRevoked, "order_paid", "first order does not meet referral threshold", event.EventKey, commonUpdates)
		}
		monthStart := time.Date(paidAt.Year(), paidAt.Month(), 1, 0, 0, 0, 0, time.UTC)
		nextMonth := monthStart.AddDate(0, 1, 0)
		convertedThisMonth, countErr := repos.Referral.CountMonthlyConvertedByReferrer(record.ReferrerID, monthStart, nextMonth)
		if countErr != nil {
			return countErr
		}
		if convertedThisMonth >= int64(config.MonthlyCapPerReferrer) {
			commonUpdates["revoked_at"] = paidAt
			commonUpdates["revoke_reason"] = "monthly referral cap reached"
			return transitionReferralRecord(repos, record, loyalty.ReferralStatusRevoked, "order_paid", "monthly referral cap reached", event.EventKey, commonUpdates)
		}

		commonUpdates["ordered_at"] = paidAt
		if err := transitionReferralRecord(repos, record, loyalty.ReferralStatusOrdered, "order_paid", "", event.EventKey, commonUpdates); err != nil {
			return err
		}
		// Registration points were already credited during binding. Order events
		// never issue, reverse, or otherwise inspect the referral points.
		return nil
	})
}

func referralRiskFlagsPresent(value any) bool {
	switch flags := value.(type) {
	case []byte:
		var values []any
		return json.Unmarshal(flags, &values) == nil && len(values) > 0
	case string:
		var values []any
		return json.Unmarshal([]byte(flags), &values) == nil && len(values) > 0
	case []map[string]any:
		return len(flags) > 0
	default:
		return false
	}
}

type referralRiskSignal struct {
	Type       string   `json:"type"`
	Level      string   `json:"level"`
	Source     string   `json:"source"`
	Dimensions []string `json:"dimensions,omitempty"`
}

func (s *ReferralService) referralRiskSignalUpdates(
	orderRepo *repository.OrderRepository,
	referralRepo *repository.ReferralRepository,
	paymentRepo *repository.PaymentRepository,
	record *loyalty.ReferralRecord,
	orderRecord *order.Order,
) (map[string]any, error) {
	updates := map[string]any{}
	if s == nil || orderRepo == nil || referralRepo == nil || record == nil || orderRecord == nil {
		return updates, nil
	}
	addressHash := s.hashSensitiveValue("shipping-address", referralShippingAddressKey(orderRecord.ShippingAddress))
	phoneHash := s.hashSensitiveValue("shipping-phone", referralShippingPhoneKey(orderRecord.ShippingAddress.Phone))
	if addressHash != "" {
		updates["shipping_address_hash"] = addressHash
	}
	if phoneHash != "" {
		updates["shipping_phone_hash"] = phoneHash
	}
	var flags []map[string]any
	if len(record.RiskFlags) > 0 {
		if err := json.Unmarshal(record.RiskFlags, &flags); err != nil {
			flags = nil
		}
	}
	riskDetected := false
	addFlag := func(flag referralRiskSignal) {
		for _, existing := range flags {
			if existingType, ok := existing["type"].(string); ok && existingType == flag.Type {
				return
			}
		}
		entry := map[string]any{"type": flag.Type, "level": flag.Level, "source": flag.Source}
		if len(flag.Dimensions) > 0 {
			entry["dimensions"] = flag.Dimensions
		}
		flags = append(flags, entry)
		riskDetected = true
	}
	historical, err := orderRepo.FindHistoricalShippingIdentitySignals(record.ReferrerID, orderRecord.ID)
	if err != nil {
		return nil, err
	}
	addressMatch, addressSimilar, phoneMatch := false, false, false
	currentAddressKey := referralShippingAddressKey(orderRecord.ShippingAddress)
	for _, signal := range historical {
		historicalAddressKey := referralShippingIdentitySignalAddressKey(signal)
		if addressHash != "" && addressHash == s.hashSensitiveValue("shipping-address", historicalAddressKey) {
			addressMatch = true
		}
		if addressHash != "" && referralAddressSimilarity(currentAddressKey, historicalAddressKey) >= 0.88 {
			addressSimilar = true
		}
		if phoneHash != "" && phoneHash == s.hashSensitiveValue("shipping-phone", referralShippingPhoneKey(signal.Phone)) {
			phoneMatch = true
		}
	}
	if addressMatch {
		addFlag(referralRiskSignal{Type: "shipping_address_match", Level: "high", Source: "referrer_paid_order", Dimensions: []string{"shipping_address"}})
	} else if addressSimilar {
		addFlag(referralRiskSignal{Type: "shipping_address_similarity", Level: "medium", Source: "referrer_paid_order", Dimensions: []string{"shipping_address_levenshtein"}})
	}
	if phoneMatch {
		addFlag(referralRiskSignal{Type: "shipping_phone_match", Level: "high", Source: "referrer_paid_order", Dimensions: []string{"shipping_phone"}})
	}
	if s.userRepo != nil && record.RefereeEmailHash != "" {
		if referrer, findErr := s.userRepo.FindByID(record.ReferrerID); findErr == nil && referrer != nil &&
			record.RefereeEmailHash == s.hashSensitiveValue("referee-email", referrer.Email) {
			addFlag(referralRiskSignal{Type: "referee_email_match", Level: "critical", Source: "referrer_account", Dimensions: []string{"email"}})
		}
	}
	if referrerRecord, recordErr := referralRepo.FindRecordByRefereeID(record.ReferrerID); recordErr == nil {
		if record.DeviceFingerprintHash != "" && record.DeviceFingerprintHash == referrerRecord.DeviceFingerprintHash {
			addFlag(referralRiskSignal{Type: "device_fingerprint_match", Level: "high", Source: "referrer_historical_attribution", Dimensions: []string{"device_fingerprint"}})
		}
	} else if recordErr != nil && !repository.IsRecordNotFound(recordErr) {
		return nil, recordErr
	}
	if paymentRepo != nil {
		transaction, transactionErr := paymentRepo.FindCompletedTransactionByOrderIDForUpdate(orderRecord.ID)
		if transactionErr == nil {
			if fingerprint := paymentFingerprintFromGatewayResponse(transaction.GatewayResponse); fingerprint != "" {
				paymentHash := s.hashSensitiveValue("payment-fingerprint", fingerprint)
				updates["payment_fingerprint_hash"] = paymentHash
				historicalTransactions, historyErr := paymentRepo.FindCompletedTransactionsByUserID(record.ReferrerID, orderRecord.ID)
				if historyErr != nil {
					return nil, historyErr
				}
				for _, historicalTransaction := range historicalTransactions {
					if historicalFingerprint := paymentFingerprintFromGatewayResponse(historicalTransaction.GatewayResponse); historicalFingerprint != "" &&
						paymentHash == s.hashSensitiveValue("payment-fingerprint", historicalFingerprint) {
						addFlag(referralRiskSignal{Type: "payment_fingerprint_match", Level: "critical", Source: "referrer_paid_order", Dimensions: []string{"payment_fingerprint"}})
						break
					}
				}
			}
		} else if !repository.IsRecordNotFound(transactionErr) {
			return nil, transactionErr
		}
	}
	if !riskDetected && len(flags) == 0 {
		return updates, nil
	}
	encoded, err := json.Marshal(flags)
	if err != nil {
		return nil, err
	}
	updates["risk_flags"] = encoded
	return updates, nil
}

func referralShippingAddressKey(address order.Address) string {
	return referralShippingIdentitySignalAddressKey(repository.ShippingIdentitySignal{
		Address1: address.Address1, Address2: address.Address2, City: address.City,
		State: address.State, PostalCode: address.PostalCode, Country: address.Country,
	})
}

func referralShippingIdentitySignalAddressKey(signal repository.ShippingIdentitySignal) string {
	parts := []string{signal.Address1, signal.Address2, signal.City, signal.State, signal.PostalCode, signal.Country}
	for index := range parts {
		parts[index] = strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(parts[index])), " "))
	}
	return strings.Join(parts, "\x1f")
}

func referralShippingPhoneKey(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}
	var digits strings.Builder
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digits.WriteRune(char)
		}
	}
	return digits.String()
}

func referralAddressSimilarity(left, right string) float64 {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" {
		return 0
	}
	if left == right {
		return 1
	}
	leftRunes, rightRunes := []rune(left), []rune(right)
	if len(leftRunes) == 0 || len(rightRunes) == 0 {
		return 0
	}
	previous := make([]int, len(rightRunes)+1)
	for index := range previous {
		previous[index] = index
	}
	for i, leftRune := range leftRunes {
		current := make([]int, len(rightRunes)+1)
		current[0] = i + 1
		for j, rightRune := range rightRunes {
			cost := 0
			if leftRune != rightRune {
				cost = 1
			}
			current[j+1] = minReferralInt(current[j]+1, previous[j+1]+1, previous[j]+cost)
		}
		previous = current
	}
	distance := previous[len(rightRunes)]
	maxLength := len(leftRunes)
	if len(rightRunes) > maxLength {
		maxLength = len(rightRunes)
	}
	return 1 - float64(distance)/float64(maxLength)
}

func minReferralInt(values ...int) int {
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}

func paymentFingerprintFromGatewayResponse(raw string) string {
	var payload any
	if strings.TrimSpace(raw) == "" || json.Unmarshal([]byte(raw), &payload) != nil {
		return ""
	}
	var visit func(any) string
	visit = func(value any) string {
		switch node := value.(type) {
		case map[string]any:
			for key, child := range node {
				if strings.EqualFold(strings.TrimSpace(key), "fingerprint") {
					if fingerprint, ok := child.(string); ok && strings.TrimSpace(fingerprint) != "" {
						return strings.TrimSpace(fingerprint)
					}
				}
			}
			for _, child := range node {
				if fingerprint := visit(child); fingerprint != "" {
					return fingerprint
				}
			}
		case []any:
			for _, child := range node {
				if fingerprint := visit(child); fingerprint != "" {
					return fingerprint
				}
			}
		}
		return ""
	}
	return visit(payload)
}

func (s *ReferralService) HandleOrderDeliveredOutbox(ctx context.Context, event outbox.Event) error {
	if err := validateReferralLifecycleHandler(s, ctx, event, outbox.EventTypeReferralOrderDelivered); err != nil {
		return err
	}
	var payload outbox.ReferralOrderDeliveredPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("%w: decode order delivered payload: %v", ErrReferralLifecycleEventInvalid, err)
	}
	if payload.OrderID == 0 || payload.DeliveredAt.IsZero() {
		return ErrReferralLifecycleEventInvalid
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Referral == nil || repos.ReferralProgram == nil {
			return ErrReferralServiceUnavailable
		}
		record, err := repos.Referral.FindRecordByOrderIDForUpdate(payload.OrderID)
		if repository.IsRecordNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if record.Status == loyalty.ReferralStatusVesting || record.Status == loyalty.ReferralStatusSettled ||
			record.Status == loyalty.ReferralStatusRevoked || record.Status == loyalty.ReferralStatusReversed {
			return nil
		}
		if record.Status != loyalty.ReferralStatusOrdered {
			return fmt.Errorf("referral %d cannot process delivery while %s", record.ID, record.Status)
		}
		config, err := repos.ReferralProgram.FindByID(record.ProgramConfigID)
		if err != nil {
			return err
		}
		deliveredAt := payload.DeliveredAt.UTC()
		return transitionReferralRecord(repos, record, loyalty.ReferralStatusVesting, "order_delivered", "", event.EventKey, map[string]any{
			"delivered_at":  deliveredAt,
			"vesting_until": deliveredAt.AddDate(0, 0, config.VestingPeriodDays),
		})
	})
}

func (s *ReferralService) HandleOrderInvalidatedOutbox(ctx context.Context, event outbox.Event) error {
	if err := validateReferralLifecycleHandler(s, ctx, event, outbox.EventTypeReferralOrderInvalidated); err != nil {
		return err
	}
	var payload outbox.ReferralOrderInvalidatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("%w: decode order invalidated payload: %v", ErrReferralLifecycleEventInvalid, err)
	}
	if payload.OrderID == 0 || payload.OccurredAt.IsZero() || strings.TrimSpace(payload.Reason) == "" {
		return ErrReferralLifecycleEventInvalid
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Referral == nil || repos.Order == nil {
			return ErrReferralServiceUnavailable
		}
		orderRecord, err := repos.Order.FindByIDForUpdate(payload.OrderID)
		if err != nil {
			return err
		}
		record, err := repos.Referral.FindRecordByOrderIDForUpdate(payload.OrderID)
		if repository.IsRecordNotFound(err) {
			record, err = repos.Referral.FindRecordByRefereeIDForUpdate(orderRecord.UserID)
		}
		if repository.IsRecordNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if record.OrderID != nil && *record.OrderID != payload.OrderID {
			return nil
		}

		occurredAt := payload.OccurredAt.UTC()
		reason := strings.TrimSpace(payload.Reason)
		updates := map[string]any{}
		if record.OrderID == nil {
			updates["order_id"] = payload.OrderID
		}
		switch record.Status {
		case loyalty.ReferralStatusPending, loyalty.ReferralStatusOrdered, loyalty.ReferralStatusVesting:
			// This state transition is for the order-bound referral attribution
			// and referrer settlement. It must not touch a referee's registration
			// points; those already belong to the user's unified balance.
			updates["revoked_at"] = occurredAt
			updates["revoke_reason"] = reason
			return transitionReferralRecord(repos, record, loyalty.ReferralStatusRevoked, "order_invalidated", reason, event.EventKey, updates)
		case loyalty.ReferralStatusSettled:
			if repos.Loyalty == nil {
				return ErrReferralServiceUnavailable
			}
			// A settled record may have a referrer reward to reverse. The
			// referee's registration points remain immutable and are handled as
			// ordinary account points, independent of this order event.
			rewardKey := fmt.Sprintf("referral:%d:referrer:points:v1", record.ID)
			reward, rewardErr := repos.Referral.FindRewardByIdempotencyKey(rewardKey)
			if rewardErr != nil && !repository.IsRecordNotFound(rewardErr) {
				return rewardErr
			}
			if reward != nil && reward.Status == loyalty.ReferralRewardStatusReleased && reward.PointsAmount > 0 {
				transaction, adjustErr := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
					record.ReferrerID, -reward.PointsAmount, "adjust", "referral_reversal", record.ID,
					fmt.Sprintf("Referral reward reversal for record #%d", record.ID), &record.ProgramConfigID,
				)
				if adjustErr != nil {
					return adjustErr
				}
				if err := repos.Referral.UpdateReward(reward.ID, map[string]any{"status": loyalty.ReferralRewardStatusReversed, "reversed_at": occurredAt, "loyalty_transaction_id": transaction.ID, "updated_at": occurredAt}); err != nil {
					return err
				}
			}
			updates["reversed_at"] = occurredAt
			updates["reverse_reason"] = reason
			return transitionReferralRecord(repos, record, loyalty.ReferralStatusReversed, "order_invalidated", reason, event.EventKey, updates)
		default:
			return nil
		}
	})
}

func validateReferralLifecycleHandler(s *ReferralService, ctx context.Context, event outbox.Event, expectedType string) error {
	if s == nil || s.txManager == nil {
		return ErrReferralServiceUnavailable
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if event.EventType != expectedType || strings.TrimSpace(event.EventKey) == "" {
		return ErrReferralLifecycleEventInvalid
	}
	return nil
}

func transitionReferralRecord(
	repos repository.TxRepositories,
	record *loyalty.ReferralRecord,
	nextStatus string,
	trigger string,
	reason string,
	eventKey string,
	updates map[string]any,
) error {
	return transitionReferralRecordAs(repos, record, nextStatus, trigger, reason, eventKey, updates, "system", nil)
}

func transitionReferralRecordAs(
	repos repository.TxRepositories,
	record *loyalty.ReferralRecord,
	nextStatus string,
	trigger string,
	reason string,
	eventKey string,
	updates map[string]any,
	actorType string,
	actorID *uint,
) error {
	if record == nil || repos.Referral == nil || !record.CanTransitionTo(nextStatus) {
		return ErrReferralLifecycleEventInvalid
	}
	if updates == nil {
		updates = map[string]any{}
	}
	updates["status"] = nextStatus
	updates["updated_at"] = time.Now().UTC()
	if err := repos.Referral.UpdateRecordState(record.ID, record.Status, record.RecordVersion, updates); err != nil {
		return err
	}
	eventKey = strings.TrimSpace(eventKey)
	if err := repos.Referral.AppendTransition(&loyalty.ReferralTransition{
		ReferralRecordID: record.ID,
		FromStatus:       record.Status,
		ToStatus:         nextStatus,
		Trigger:          trigger,
		Reason:           reason,
		ActorType:        actorType,
		ActorID:          actorID,
		EventKey:         &eventKey,
	}); err != nil {
		return err
	}
	record.Status = nextStatus
	record.RecordVersion++
	return nil
}
