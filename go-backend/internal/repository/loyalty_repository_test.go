package repository

import (
	"testing"

	"tanzanite/internal/domain/loyalty"
)

func TestApplyPointsDelta(t *testing.T) {
	userLoyalty := &loyalty.UserLoyalty{
		TotalPoints:     100,
		AvailablePoints: 40,
		UsedPoints:      60,
	}

	applyPointsDelta(userLoyalty, 25, "earn")
	if userLoyalty.TotalPoints != 125 || userLoyalty.AvailablePoints != 65 || userLoyalty.UsedPoints != 60 {
		t.Fatalf("earn delta = total %d available %d used %d", userLoyalty.TotalPoints, userLoyalty.AvailablePoints, userLoyalty.UsedPoints)
	}

	applyPointsDelta(userLoyalty, -30, "spend")
	if userLoyalty.TotalPoints != 125 || userLoyalty.AvailablePoints != 35 || userLoyalty.UsedPoints != 90 {
		t.Fatalf("spend delta = total %d available %d used %d", userLoyalty.TotalPoints, userLoyalty.AvailablePoints, userLoyalty.UsedPoints)
	}

	applyPointsDelta(userLoyalty, 20, "refund")
	if userLoyalty.TotalPoints != 125 || userLoyalty.AvailablePoints != 55 || userLoyalty.UsedPoints != 70 {
		t.Fatalf("refund delta = total %d available %d used %d", userLoyalty.TotalPoints, userLoyalty.AvailablePoints, userLoyalty.UsedPoints)
	}
}

func TestApplyPointsDeltaRefundDoesNotMakeUsedNegative(t *testing.T) {
	userLoyalty := &loyalty.UserLoyalty{
		AvailablePoints: 10,
		UsedPoints:      5,
	}

	applyPointsDelta(userLoyalty, 10, "refund")
	if userLoyalty.AvailablePoints != 20 || userLoyalty.UsedPoints != 0 {
		t.Fatalf("refund clamp = available %d used %d", userLoyalty.AvailablePoints, userLoyalty.UsedPoints)
	}
}
