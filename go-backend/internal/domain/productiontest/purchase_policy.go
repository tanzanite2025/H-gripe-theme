package productiontest

import "time"

const (
	AccountStatusActive = TestAccountStatusActive

	ReasonNormalProduct                   = "normal_product"
	ReasonTestAccountRequired             = "test_account_required"
	ReasonTestAccountInactive             = "test_account_inactive"
	ReasonTestAccountUserMismatch         = "test_account_user_mismatch"
	ReasonTestAccountNotAllowed           = "test_account_not_allowed"
	ReasonTestProductPurchase             = "test_product_purchase"
	ReasonEmptyOrder                      = "empty_order"
	ReasonOrderRequiresProductionTestMark = "order_requires_production_test_mark"
)

type PurchaseItem struct {
	ProductID uint
	VariantID *uint
}

type PurchaseInput struct {
	UserID    uint
	ProductID uint
	VariantID *uint
	Account   *TestAccount
	Gate      ProductGate
	At        time.Time
}

type PurchaseDecision struct {
	Allowed         bool
	MarkerRequired  bool
	UserID          uint
	TestAccountID   uint
	HoldFulfillment bool
	Reason          string
}

// EvaluatePurchase answers only the product eligibility question. It does not
// load data, mutate stock, create orders, or invoke payment/after-sales code.
func EvaluatePurchase(input PurchaseInput) PurchaseDecision {
	item := PurchaseItem{
		ProductID: input.ProductID,
		VariantID: input.VariantID,
	}
	if !input.Gate.AppliesTo(item, input.At) {
		return PurchaseDecision{
			Allowed: true,
			Reason:  ReasonNormalProduct,
		}
	}

	if input.Account == nil {
		return PurchaseDecision{
			Reason: ReasonTestAccountRequired,
		}
	}
	if !input.Account.IsActiveAt(input.At) {
		return PurchaseDecision{
			Reason: ReasonTestAccountInactive,
		}
	}
	if input.Account.UserID != input.UserID {
		return PurchaseDecision{
			Reason: ReasonTestAccountUserMismatch,
		}
	}
	if !input.Gate.Allows(*input.Account) {
		return PurchaseDecision{
			Reason: ReasonTestAccountNotAllowed,
		}
	}

	return PurchaseDecision{
		Allowed:         true,
		MarkerRequired:  true,
		UserID:          input.UserID,
		TestAccountID:   input.Account.ID,
		HoldFulfillment: input.Gate.HoldFulfillment,
		Reason:          ReasonTestProductPurchase,
	}
}

type OrderLineInput struct {
	ProductID uint
	VariantID *uint
	Gate      ProductGate
}

type OrderInput struct {
	UserID  uint
	Account *TestAccount
	Lines   []OrderLineInput
	At      time.Time
}

// EvaluateOrder enforces every line and marks the order when at least one
// allowed line is an active test-only product.
func EvaluateOrder(input OrderInput) PurchaseDecision {
	if len(input.Lines) == 0 {
		return PurchaseDecision{Reason: ReasonEmptyOrder}
	}

	requiresMarker := false
	holdFulfillment := false
	for _, line := range input.Lines {
		decision := EvaluatePurchase(PurchaseInput{
			UserID:    input.UserID,
			ProductID: line.ProductID,
			VariantID: line.VariantID,
			Account:   input.Account,
			Gate:      line.Gate,
			At:        input.At,
		})
		if !decision.Allowed {
			return decision
		}
		if decision.MarkerRequired {
			requiresMarker = true
			holdFulfillment = holdFulfillment || decision.HoldFulfillment
		}
	}

	if requiresMarker {
		return PurchaseDecision{
			Allowed:         true,
			MarkerRequired:  true,
			UserID:          input.UserID,
			TestAccountID:   input.Account.ID,
			HoldFulfillment: holdFulfillment,
			Reason:          ReasonOrderRequiresProductionTestMark,
		}
	}

	return PurchaseDecision{
		Allowed: true,
		Reason:  ReasonNormalProduct,
	}
}
