package currency

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/datatypes"
)

const OrderFXSnapshotVersion = 1

// OrderFXSnapshot is the immutable conversion contract captured when an order
// is created. RateDecimal means 1 base-currency unit equals this many
// order-currency units. It is kept as a decimal string so the captured
// contract never passes through binary floating point.
type OrderFXSnapshot struct {
	Version       int        `json:"version"`
	BaseCurrency  string     `json:"base_currency"`
	OrderCurrency string     `json:"order_currency"`
	RateDecimal   string     `json:"rate_decimal"`
	Source        string     `json:"source"`
	CapturedAt    time.Time  `json:"captured_at"`
	RateFetchedAt *time.Time `json:"rate_fetched_at,omitempty"`
}

// RateRat parses the captured decimal exactly for domain calculations. The
// returned rational is independent from the snapshot and may be safely
// inverted or otherwise modified by the caller.
func (s OrderFXSnapshot) RateRat() (*big.Rat, error) {
	value := strings.TrimSpace(s.RateDecimal)
	if value == "" || strings.Contains(value, "/") {
		return nil, errors.New("order FX snapshot rate must be a decimal")
	}
	rate, ok := new(big.Rat).SetString(value)
	if !ok || rate.Sign() <= 0 {
		return nil, errors.New("order FX snapshot rate must be greater than zero")
	}
	return rate, nil
}

func (s OrderFXSnapshot) Validate(expectedOrderCurrency string) error {
	base := NormalizeCode(s.BaseCurrency)
	order := NormalizeCode(s.OrderCurrency)
	expected := NormalizeCode(expectedOrderCurrency)
	if s.Version != OrderFXSnapshotVersion {
		return fmt.Errorf("unsupported order FX snapshot version %d", s.Version)
	}
	if !IsCatalogCode(base) {
		return errors.New("order FX snapshot base currency is invalid")
	}
	if !IsCatalogCode(order) {
		return errors.New("order FX snapshot order currency is invalid")
	}
	if expected != "" && order != expected {
		return fmt.Errorf("order FX snapshot currency %s does not match transaction currency %s", order, expected)
	}
	rate, err := s.RateRat()
	if err != nil {
		return err
	}
	if strings.TrimSpace(s.Source) == "" {
		return errors.New("order FX snapshot source is required")
	}
	if s.CapturedAt.IsZero() {
		return errors.New("order FX snapshot capture time is required")
	}
	if base == order && rate.Cmp(big.NewRat(1, 1)) != 0 {
		return errors.New("same-currency order FX snapshot must use rate 1")
	}
	return nil
}

func OrderFXSnapshotJSON(snapshot OrderFXSnapshot) datatypes.JSON {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(encoded)
}

func ParseOrderFXSnapshot(raw datatypes.JSON) (OrderFXSnapshot, error) {
	if len(raw) == 0 || string(raw) == "{}" {
		return OrderFXSnapshot{}, errors.New("order FX snapshot is missing")
	}
	var snapshot OrderFXSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return OrderFXSnapshot{}, fmt.Errorf("decode order FX snapshot: %w", err)
	}
	if err := snapshot.Validate(snapshot.OrderCurrency); err != nil {
		return OrderFXSnapshot{}, err
	}
	return snapshot, nil
}
