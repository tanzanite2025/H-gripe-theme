package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/coupon"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	domainpricing "commerce-platform/internal/domain/pricing"
)

var ErrCheckoutPricingSnapshotInvalid = errors.New("checkout pricing snapshot does not match calculated checkout totals")

type checkoutPricingLineInput struct {
	ProductID           uint
	VariantID           uint
	Quantity            int
	ProductCategoryID   *uint
	ProductCategorySlug string
	UnitPrice           domainmoney.Money
	Subtotal            domainmoney.Money
}

// checkoutPricingInput is the exact-money boundary for the pricing snapshot.
// OrderItem is retained only for non-monetary identity and coupon-scope data;
// every amount used by the pricing pipeline is carried as Money.
type checkoutPricingInput struct {
	Items               []checkoutPricingLineInput
	Currency            string
	Subtotal            domainmoney.Money
	MemberDiscount      domainmoney.Money
	CouponDiscount      domainmoney.Money
	Coupon              *coupon.Coupon
	MerchandiseNetTotal domainmoney.Money
}

func buildCheckoutPricingSnapshot(input checkoutPricingInput) (domainpricing.Snapshot, error) {
	lineInputs := make([]domainpricing.LineInput, len(input.Items))
	lineKeys := make([]string, len(input.Items))
	for i, item := range input.Items {
		if err := validateCheckoutPricingMoney(item.UnitPrice, input.Currency, fmt.Sprintf("line %d unit price", i)); err != nil {
			return domainpricing.Snapshot{}, err
		}
		if err := validateCheckoutPricingMoney(item.Subtotal, input.Currency, fmt.Sprintf("line %d subtotal", i)); err != nil {
			return domainpricing.Snapshot{}, err
		}
		key := checkoutPricingLineKey(i, item.ProductID, item.VariantID)
		lineKeys[i] = key
		lineInputs[i] = domainpricing.LineInput{
			Key:       key,
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	stages := make([]domainpricing.DiscountInput, 0, 3)
	if err := validateCheckoutPricingMoney(input.Subtotal, input.Currency, "subtotal"); err != nil {
		return domainpricing.Snapshot{}, err
	}
	memberDiscount := input.MemberDiscount
	if err := validateCheckoutPricingMoney(memberDiscount, input.Currency, "member discount"); err != nil {
		return domainpricing.Snapshot{}, err
	}
	if memberDiscount.AmountMinor() > 0 {
		stages = append(stages, domainpricing.DiscountInput{
			Kind:      domainpricing.DiscountKindMember,
			Reference: "membership",
			Amount:    memberDiscount,
		})
	}

	couponDiscount := input.CouponDiscount
	if err := validateCheckoutPricingMoney(couponDiscount, input.Currency, "coupon discount"); err != nil {
		return domainpricing.Snapshot{}, err
	}
	if input.Coupon != nil || couponDiscount.AmountMinor() > 0 {
		if input.Coupon == nil {
			return domainpricing.Snapshot{}, fmt.Errorf("%w: coupon discount has no coupon", ErrCheckoutPricingSnapshotInvalid)
		}
		eligibleKeys, scopeErr := checkoutCouponEligibleLineKeys(input.Items, lineKeys, input.Coupon)
		if scopeErr != nil {
			return domainpricing.Snapshot{}, scopeErr
		}
		stages = append(stages, domainpricing.DiscountInput{
			Kind:             domainpricing.DiscountKindCoupon,
			Reference:        input.Coupon.Code,
			Amount:           couponDiscount,
			EligibleLineKeys: eligibleKeys,
		})
	}

	pipeline, err := domainpricing.NewPipeline(stages...)
	if err != nil {
		return domainpricing.Snapshot{}, err
	}
	snapshot, err := pipeline.Calculate(lineInputs)
	if err != nil {
		return domainpricing.Snapshot{}, err
	}
	if err := validateCheckoutPricingParity(snapshot, input); err != nil {
		return domainpricing.Snapshot{}, err
	}
	return snapshot, nil
}

func validateCheckoutPricingParity(snapshot domainpricing.Snapshot, input checkoutPricingInput) error {
	lines := snapshot.Lines()
	if len(lines) != len(input.Items) {
		return fmt.Errorf("%w: line count differs", ErrCheckoutPricingSnapshotInvalid)
	}
	for i, line := range lines {
		expectedSubtotal := input.Items[i].Subtotal
		if line.BaseSubtotal().AmountMinor() != expectedSubtotal.AmountMinor() {
			return fmt.Errorf(
				"%w: line %s exact subtotal is %d minor units, expected subtotal is %d",
				ErrCheckoutPricingSnapshotInvalid,
				line.Key(),
				line.BaseSubtotal().AmountMinor(),
				expectedSubtotal.AmountMinor(),
			)
		}
	}

	expectedSubtotal := input.Subtotal
	if snapshot.BaseTotal().AmountMinor() != expectedSubtotal.AmountMinor() {
		return fmt.Errorf(
			"%w: exact subtotal is %d minor units, expected subtotal is %d",
			ErrCheckoutPricingSnapshotInvalid,
			snapshot.BaseTotal().AmountMinor(),
			expectedSubtotal.AmountMinor(),
		)
	}

	expectedNet := input.MerchandiseNetTotal
	if snapshot.NetTotal().AmountMinor() != expectedNet.AmountMinor() {
		return fmt.Errorf(
			"%w: exact merchandise net is %d minor units, expected net is %d",
			ErrCheckoutPricingSnapshotInvalid,
			snapshot.NetTotal().AmountMinor(),
			expectedNet.AmountMinor(),
		)
	}
	return nil
}

func checkoutCouponEligibleLineKeys(items []checkoutPricingLineInput, lineKeys []string, appliedCoupon *coupon.Coupon) ([]string, error) {
	applicable, err := parseCouponProductIDs(appliedCoupon.ApplicableProducts)
	if err != nil {
		return nil, fmt.Errorf("coupon applicable products: %w", err)
	}
	excluded, err := parseCouponProductIDs(appliedCoupon.ExcludedProducts)
	if err != nil {
		return nil, fmt.Errorf("coupon excluded products: %w", err)
	}
	applicableCategories, err := parseCouponCategoryValues(appliedCoupon.ApplicableCategories)
	if err != nil {
		return nil, fmt.Errorf("coupon applicable categories: %w", err)
	}
	if len(applicable) == 0 && len(excluded) == 0 && len(applicableCategories.IDs) == 0 && len(applicableCategories.Slugs) == 0 {
		return nil, nil
	}

	eligible := make([]string, 0, len(items))
	for i, item := range items {
		if _, blocked := excluded[item.ProductID]; blocked {
			continue
		}
		if len(applicable) > 0 {
			if _, ok := applicable[item.ProductID]; !ok {
				continue
			}
		}
		if len(applicableCategories.IDs) > 0 || len(applicableCategories.Slugs) > 0 {
			if item.ProductCategoryID == nil {
				if _, ok := applicableCategories.Slugs[strings.ToLower(strings.TrimSpace(item.ProductCategorySlug))]; !ok {
					continue
				}
			} else {
				_, idOK := applicableCategories.IDs[*item.ProductCategoryID]
				_, slugOK := applicableCategories.Slugs[strings.ToLower(strings.TrimSpace(item.ProductCategorySlug))]
				if !idOK && !slugOK {
					continue
				}
			}
		}
		eligible = append(eligible, lineKeys[i])
	}
	if len(eligible) == 0 {
		return nil, fmt.Errorf("%w: coupon has no eligible pricing lines", ErrCheckoutPricingSnapshotInvalid)
	}
	return eligible, nil
}

func validateCheckoutPricingMoney(value domainmoney.Money, currencyCode string, label string) error {
	if err := value.Validate(); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if value.Currency().String() != currencyCode {
		return fmt.Errorf(
			"%w: %s currency %s does not match checkout currency %s",
			ErrCheckoutPricingSnapshotInvalid,
			label,
			value.Currency(),
			currencyCode,
		)
	}
	if value.AmountMinor() < 0 {
		return fmt.Errorf("%w: %s is negative", ErrCheckoutPricingSnapshotInvalid, label)
	}
	return nil
}

func checkoutPricingLineKey(index int, productID, variantID uint) string {
	return fmt.Sprintf("%d:%d:%d", index, productID, variantID)
}

func orderItemVariantID(item order.OrderItem) uint {
	if item.VariantID == nil {
		return 0
	}
	return *item.VariantID
}
