package dto

// Primitives shared across the bank-to-bank protocol. See sections 2.1–2.7 of
// the protocol specification.

// RoutingNumber is the three-digit prefix of account numbers that uniquely
// identifies a bank.
type RoutingNumber = int

// IdempotenceKey uniquely tags a message exchanged between banks. The sender
// MUST NOT reuse keys; the receiver MUST track them indefinitely.
type IdempotenceKey struct {
	RoutingNumber       RoutingNumber `json:"routingNumber"       binding:"required"`
	LocallyGeneratedKey string        `json:"locallyGeneratedKey" binding:"required,max=64"`
}

// ForeignBankId identifies an object held by another bank. The id string is
// opaque to anyone other than the bank with the matching routing number.
type ForeignBankId struct {
	RoutingNumber RoutingNumber `json:"routingNumber" binding:"required"`
	ID            string        `json:"id"            binding:"required,max=64"`
}

// ISO8601DateTimeWithTimeZone is an ISO-8601 timestamp string, e.g.
// "2025-04-16T15:32:44+02:00".
type ISO8601DateTimeWithTimeZone = string

// CurrencyCode is the ISO 4217 currency code limited to the set agreed upon
// by participating banks.
type CurrencyCode string

const (
	CurrencyRSD CurrencyCode = "RSD"
	CurrencyEUR CurrencyCode = "EUR"
	CurrencyUSD CurrencyCode = "USD"
	CurrencyCHF CurrencyCode = "CHF"
	CurrencyJPY CurrencyCode = "JPY"
	CurrencyAUD CurrencyCode = "AUD"
	CurrencyCAD CurrencyCode = "CAD"
	CurrencyGBP CurrencyCode = "GBP"
)

// MonetaryValue is an amount in a given currency. The spec recommends a
// BigDecimal-style representation; we use float64 to remain consistent with
// the rest of this codebase.
type MonetaryValue struct {
	Currency CurrencyCode `json:"currency" binding:"required"`
	Amount   float64      `json:"amount"   binding:"required"`
}
