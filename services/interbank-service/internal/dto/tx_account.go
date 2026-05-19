package dto

// TxAccountType is the discriminator for TxAccount.
type TxAccountType string

const (
	TxAccountPerson  TxAccountType = "PERSON"
	TxAccountAccount TxAccountType = "ACCOUNT"
	TxAccountOption  TxAccountType = "OPTION"
)

// CurrencyAccountNumber is a bank-specific currency/current account number.
type CurrencyAccountNumber = string

// TxAccount is the protocol's tagged union over the three account types
// (§2.6). The JSON wire format is:
//
//	{"type": "PERSON",  "id":  ForeignBankId}
//	{"type": "ACCOUNT", "num": CurrencyAccountNumber}
//	{"type": "OPTION",  "id":  ForeignBankId}
//
// All three variants are flat, so the union is expressed as a single
// struct with optional pointers. The Type field tells you which one is set.
type TxAccount struct {
	Type TxAccountType          `json:"type" binding:"required,oneof=PERSON ACCOUNT OPTION"`
	ID   *ForeignBankId         `json:"id,omitempty"`
	Num  *CurrencyAccountNumber `json:"num,omitempty"`
}

// AssetType is the discriminator for Asset.
type AssetType string

const (
	AssetMonas  AssetType = "MONAS"
	AssetStock  AssetType = "STOCK"
	AssetOption AssetType = "OPTION"
)

// MonetaryAsset is a currency-only asset descriptor (§2.7.1).
type MonetaryAsset struct {
	Currency CurrencyCode `json:"currency" binding:"required"`
}

// StockDescription identifies a stock by ticker (§2.7.3).
type StockDescription struct {
	Ticker string `json:"ticker" binding:"required"`
}

// OptionDescription is an OTC option contract description (§2.7.2).
type OptionDescription struct {
	NegotiationID  ForeignBankId               `json:"negotiationId"`
	Stock          StockDescription            `json:"stock"`
	PricePerUnit   MonetaryValue               `json:"pricePerUnit"`
	SettlementDate ISO8601DateTimeWithTimeZone `json:"settlementDate"`
	Amount         float64                     `json:"amount"`
}

// Asset is the protocol's tagged union over asset kinds (§2.7). Wire shape:
//
//	{"type": "MONAS",  "asset": MonetaryAsset}
//	{"type": "STOCK",  "asset": StockDescription}
//	{"type": "OPTION", "asset": OptionDescription}
//
// Because the variants share the wrapper field name "asset" but with
// different inner shapes, we hold the inner object as map[string]any. Gin
// binds it natively; typed access (when we need to act on a specific
// asset kind) is added in the prompt that introduces real transaction
// preparation.
type Asset struct {
	Type AssetType      `json:"type"  binding:"required,oneof=MONAS STOCK OPTION"`
	Body map[string]any `json:"asset" binding:"required"`
}
