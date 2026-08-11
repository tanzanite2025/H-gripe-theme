package service

import (
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/config"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPaymentProtectionControlLifecycleAndEvaluation(t *testing.T) {
	db := newPaymentProtectionTestDB(t)
	service := NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)
	actor := PaymentProtectionActor{
		UserID:    7,
		Username:  "operator",
		IPAddress: "203.0.113.7",
		Method:    "POST",
		Path:      "/api/admin/payment/risk/controls",
	}

	control, err := service.CreateControl(CreatePaymentProtectionControlInput{
		Action:     "force_3ds",
		ScopeType:  "country",
		ScopeValue: "us",
		Reason:     "manual review of elevated US card activity",
		ExpiresAt:  time.Now().UTC().Add(2 * time.Hour),
	}, actor)
	require.NoError(t, err)
	require.Equal(t, "active", control.Status)

	decision, err := service.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider: "stripe",
		Country:  "US",
	})
	require.NoError(t, err)
	require.True(t, decision.Force3DS)
	require.Contains(t, decision.Reasons, "manual_force_3ds_control_"+strconv.FormatUint(uint64(control.ID), 10))

	nonMatchingDecision, err := service.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider: "stripe",
		Country:  "CA",
	})
	require.NoError(t, err)
	require.False(t, nonMatchingDecision.Force3DS)

	logs, total, err := service.ListControlAuditLogs(control.ID, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
	require.Equal(t, "create", logs[0]["action"])
	require.Equal(t, "operator", logs[0]["username"])

	revoked, err := service.RevokeControl(control.ID, actor)
	require.NoError(t, err)
	require.Equal(t, "revoked", revoked.Status)

	decision, err = service.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider: "stripe",
		Country:  "US",
	})
	require.NoError(t, err)
	require.False(t, decision.Force3DS)

	logs, total, err = service.ListControlAuditLogs(control.ID, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, logs, 2)
	require.Equal(t, "revoke", logs[0]["action"])
}

func TestPaymentProtectionControlRequiresBoundedExpiry(t *testing.T) {
	db := newPaymentProtectionTestDB(t)
	service := NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)

	_, err := service.CreateControl(CreatePaymentProtectionControlInput{
		Action:    "force_3ds",
		ScopeType: "global",
		Reason:    "missing bounded expiry",
		ExpiresAt: time.Now().UTC().Add(25 * time.Hour),
	}, PaymentProtectionActor{UserID: 1})
	require.ErrorIs(t, err, ErrInvalidPaymentProtectionControl)
}

func TestPaymentProtectionPausePaymentHasShorterDurationLimits(t *testing.T) {
	db := newPaymentProtectionTestDB(t)
	service := NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                            true,
			MaxControlDurationHours:            168,
			MaxPausePaymentDurationHours:       24,
			MaxGlobalPausePaymentDurationHours: 2,
		},
	)

	_, err := service.CreateControl(CreatePaymentProtectionControlInput{
		Action:     "pause_payment",
		ScopeType:  "provider",
		ScopeValue: string(pgateway.GatewayStripe),
		Reason:     "provider incident review window",
		ExpiresAt:  time.Now().UTC().Add(25 * time.Hour),
	}, PaymentProtectionActor{UserID: 1})
	require.ErrorIs(t, err, ErrInvalidPaymentProtectionControl)

	_, err = service.CreateControl(CreatePaymentProtectionControlInput{
		Action:    "pause_payment",
		ScopeType: "global",
		Reason:    "global incident bridge",
		ExpiresAt: time.Now().UTC().Add(3 * time.Hour),
	}, PaymentProtectionActor{UserID: 1})
	require.ErrorIs(t, err, ErrInvalidPaymentProtectionControl)

	control, err := service.CreateControl(CreatePaymentProtectionControlInput{
		Action:    "pause_payment",
		ScopeType: "global",
		Reason:    "short global incident bridge",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}, PaymentProtectionActor{UserID: 1})
	require.NoError(t, err)
	require.Equal(t, "active", control.Status)
}

func TestPaymentProtectionPausePaymentControlIsScopedAndRevocable(t *testing.T) {
	db := newPaymentProtectionTestDB(t)
	service := NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)
	actor := PaymentProtectionActor{
		UserID:   11,
		Username: "risk-lead",
	}

	control, err := service.CreateControl(CreatePaymentProtectionControlInput{
		Action:     "pause_payment",
		ScopeType:  "country",
		ScopeValue: "GB",
		Reason:     "temporary acquiring incident under review",
		ExpiresAt:  time.Now().UTC().Add(2 * time.Hour),
	}, actor)
	require.NoError(t, err)

	decision, err := service.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider:      "stripe",
		Country:       "GB",
		PaymentMethod: "stripe",
	})
	require.NoError(t, err)
	require.True(t, decision.PausePayment)
	require.Contains(t, decision.Reasons, "manual_pause_payment_control_"+strconv.FormatUint(uint64(control.ID), 10))

	nonMatchingDecision, err := service.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider:      "stripe",
		Country:       "US",
		PaymentMethod: "stripe",
	})
	require.NoError(t, err)
	require.False(t, nonMatchingDecision.PausePayment)

	_, err = service.RevokeControl(control.ID, actor)
	require.NoError(t, err)

	decision, err = service.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider:      "stripe",
		Country:       "GB",
		PaymentMethod: "stripe",
	})
	require.NoError(t, err)
	require.False(t, decision.PausePayment)
}

func TestPaymentProtectionProviderScopeAcceptsAllGatewayTypes(t *testing.T) {
	db := newPaymentProtectionTestDB(t)
	service := NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)

	control, err := service.CreateControl(CreatePaymentProtectionControlInput{
		Action:     "pause_payment",
		ScopeType:  "provider",
		ScopeValue: string(pgateway.GatewayAlipay),
		Reason:     "temporary wallet acquiring incident",
		ExpiresAt:  time.Now().UTC().Add(2 * time.Hour),
	}, PaymentProtectionActor{UserID: 12})
	require.NoError(t, err)
	require.Equal(t, string(pgateway.GatewayAlipay), control.ScopeValue)

	decision, err := service.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider: string(pgateway.GatewayAlipay),
		Country:  "CN",
	})
	require.NoError(t, err)
	require.True(t, decision.PausePayment)

	decision, err = service.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider: string(pgateway.GatewayStripe),
		Country:  "CN",
	})
	require.NoError(t, err)
	require.False(t, decision.PausePayment)
}

func newPaymentProtectionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(
		&paymentdomain.PaymentProtectionControl{},
		&audit.AuditLog{},
	))
	return db
}
