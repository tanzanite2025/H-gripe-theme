package shipping

import "time"

// YanwenOrderFacts contains only the immutable order facts required to build
// an express.order.create request. It is a read-only cross-domain projection;
// the Yanwen domain must not receive the complete order aggregate.
type YanwenOrderFacts struct {
	ID              uint
	OrderNumber     string
	Status          string
	PaymentStatus   string
	Currency        string
	ShippingAddress YanwenOrderFactAddress
	Items           []YanwenOrderFactItem
	PaidAt          *time.Time
}

// YanwenOrderFactAddress contains the recipient fields sent to Yanwen.
type YanwenOrderFactAddress struct {
	FirstName  string
	LastName   string
	Company    string
	Address1   string
	Address2   string
	City       string
	State      string
	PostalCode string
	Country    string
	Phone      string
	Email      string
}

// YanwenOrderFactItem contains the declaration fields sent to Yanwen.
type YanwenOrderFactItem struct {
	ProductName            string
	SKU                    string
	Quantity               int
	WeightGrams            int
	HSCode                 string
	CustomsDescription     string
	DeclaredValueMinor     *int64
	DeclaredValueConfirmed bool
}
