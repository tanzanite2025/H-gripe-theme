package productiontest

import (
	"testing"
	"time"
)

func TestEvaluatePurchaseAllowsNormalProductWithoutTestMarker(t *testing.T) {
	decision := EvaluatePurchase(PurchaseInput{
		UserID:    10,
		ProductID: 20,
		At:        testPolicyTime(),
		Gate:      ProductGate{ProductID: 20, IsTestOnly: false, Enabled: true},
	})

	if !decision.Allowed || decision.MarkerRequired {
		t.Fatalf("normal product decision = %+v", decision)
	}
	if decision.Reason != ReasonNormalProduct {
		t.Fatalf("normal product reason = %q", decision.Reason)
	}
}

func TestEvaluatePurchaseRejectsMissingOrInactiveTestAccount(t *testing.T) {
	activeGate := ProductGate{
		ProductID:  20,
		IsTestOnly: true,
		Enabled:    true,
	}

	tests := []struct {
		name    string
		account *TestAccount
		reason  string
	}{
		{name: "missing account", reason: ReasonTestAccountRequired},
		{
			name:    "disabled account",
			account: &TestAccount{ID: 1, UserID: 10, Status: "disabled"},
			reason:  ReasonTestAccountInactive,
		},
		{
			name: "expired account",
			account: &TestAccount{
				ID:        1,
				UserID:    10,
				Status:    AccountStatusActive,
				ExpiresAt: timePtr(testPolicyTime().Add(-time.Minute)),
			},
			reason: ReasonTestAccountInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := EvaluatePurchase(PurchaseInput{
				UserID:    10,
				ProductID: 20,
				Account:   tt.account,
				Gate:      activeGate,
				At:        testPolicyTime(),
			})
			if decision.Allowed || decision.MarkerRequired {
				t.Fatalf("expected denied decision, got %+v", decision)
			}
			if decision.Reason != tt.reason {
				t.Fatalf("reason = %q, want %q", decision.Reason, tt.reason)
			}
		})
	}
}

func TestEvaluatePurchaseAllowsActiveAuthorizedTestAccountAndRequiresMarker(t *testing.T) {
	account := &TestAccount{
		ID:     7,
		UserID: 10,
		Status: TestAccountStatusActive,
	}
	decision := EvaluatePurchase(PurchaseInput{
		UserID:    10,
		ProductID: 20,
		Account:   account,
		Gate: ProductGate{
			ProductID:            20,
			IsTestOnly:           true,
			Enabled:              true,
			HoldFulfillment:      true,
			AllowedTestAccountID: uintPtr(7),
		},
		At: testPolicyTime(),
	})

	if !decision.Allowed || !decision.MarkerRequired {
		t.Fatalf("authorized test purchase decision = %+v", decision)
	}
	if decision.Reason != ReasonTestProductPurchase {
		t.Fatalf("reason = %q, want %q", decision.Reason, ReasonTestProductPurchase)
	}
	if decision.UserID != 10 || decision.TestAccountID != 7 || !decision.HoldFulfillment {
		t.Fatalf("test marker context = %+v", decision)
	}
}

func TestEvaluatePurchaseDoesNotMarkNormalProductForTestAccount(t *testing.T) {
	decision := EvaluatePurchase(PurchaseInput{
		UserID:    10,
		ProductID: 20,
		Account:   &TestAccount{ID: 7, UserID: 10, Status: TestAccountStatusActive},
		Gate:      ProductGate{ProductID: 20, Enabled: true},
		At:        testPolicyTime(),
	})

	if !decision.Allowed || decision.MarkerRequired || decision.Reason != ReasonNormalProduct {
		t.Fatalf("test account normal product decision = %+v", decision)
	}
}

func TestEvaluatePurchaseRejectsWrongUserOrAccountScope(t *testing.T) {
	account := &TestAccount{
		ID:     7,
		UserID: 10,
		Status: TestAccountStatusActive,
	}
	gate := ProductGate{
		ProductID:            20,
		IsTestOnly:           true,
		Enabled:              true,
		AllowedTestAccountID: uintPtr(8),
	}

	decision := EvaluatePurchase(PurchaseInput{
		UserID:    11,
		ProductID: 20,
		Account:   account,
		Gate:      gate,
		At:        testPolicyTime(),
	})
	if decision.Allowed || decision.Reason != ReasonTestAccountUserMismatch {
		t.Fatalf("wrong user decision = %+v", decision)
	}

	decision = EvaluatePurchase(PurchaseInput{
		UserID:    10,
		ProductID: 20,
		Account:   account,
		Gate:      gate,
		At:        testPolicyTime(),
	})
	if decision.Allowed || decision.Reason != ReasonTestAccountNotAllowed {
		t.Fatalf("wrong account scope decision = %+v", decision)
	}
}

func TestEvaluatePurchaseTreatsInactiveGateAsNormalProduct(t *testing.T) {
	account := &TestAccount{ID: 7, UserID: 10, Status: TestAccountStatusActive}
	decision := EvaluatePurchase(PurchaseInput{
		UserID:    10,
		ProductID: 20,
		Account:   account,
		Gate: ProductGate{
			ProductID:  20,
			IsTestOnly: true,
			Enabled:    true,
			StartsAt:   timePtr(testPolicyTime().Add(time.Hour)),
		},
		At: testPolicyTime(),
	})

	if !decision.Allowed || decision.MarkerRequired || decision.Reason != ReasonNormalProduct {
		t.Fatalf("inactive gate decision = %+v", decision)
	}
}

func TestEvaluateOrderMarksMixedCartWhenAnyTestProductIsAllowed(t *testing.T) {
	account := &TestAccount{ID: 7, UserID: 10, Status: AccountStatusActive}
	decision := EvaluateOrder(OrderInput{
		UserID:  10,
		Account: account,
		At:      testPolicyTime(),
		Lines: []OrderLineInput{
			{ProductID: 1, Gate: ProductGate{ProductID: 1, Enabled: true}},
			{
				ProductID: 2,
				Gate:      ProductGate{ProductID: 2, IsTestOnly: true, Enabled: true, HoldFulfillment: true},
			},
		},
	})

	if !decision.Allowed || !decision.MarkerRequired {
		t.Fatalf("mixed order decision = %+v", decision)
	}
	if decision.Reason != ReasonOrderRequiresProductionTestMark {
		t.Fatalf("reason = %q, want %q", decision.Reason, ReasonOrderRequiresProductionTestMark)
	}
	if decision.UserID != 10 || decision.TestAccountID != 7 || !decision.HoldFulfillment {
		t.Fatalf("mixed order marker context = %+v", decision)
	}
}

func TestEvaluateOrderRejectsAnyUnauthorizedTestProduct(t *testing.T) {
	decision := EvaluateOrder(OrderInput{
		UserID: 10,
		At:     testPolicyTime(),
		Lines: []OrderLineInput{
			{ProductID: 1, Gate: ProductGate{ProductID: 1, Enabled: true}},
			{ProductID: 2, Gate: ProductGate{ProductID: 2, IsTestOnly: true, Enabled: true}},
		},
	})

	if decision.Allowed || decision.MarkerRequired {
		t.Fatalf("unauthorized mixed order decision = %+v", decision)
	}
	if decision.Reason != ReasonTestAccountRequired {
		t.Fatalf("reason = %q, want %q", decision.Reason, ReasonTestAccountRequired)
	}
}

func TestEvaluateOrderRejectsEmptyOrder(t *testing.T) {
	decision := EvaluateOrder(OrderInput{UserID: 10, At: testPolicyTime()})
	if decision.Allowed || decision.MarkerRequired || decision.Reason != ReasonEmptyOrder {
		t.Fatalf("empty order decision = %+v", decision)
	}
}

func testPolicyTime() time.Time {
	return time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func uintPtr(value uint) *uint {
	return &value
}
