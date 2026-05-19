package dto

// Posting is a single credit or debit in a transaction (§2.8.1). A negative
// amount credits the account; a positive amount debits it.
type Posting struct {
	Account TxAccount `json:"account" binding:"required"`
	Amount  float64   `json:"amount"  binding:"required"`
	Asset   Asset     `json:"asset"   binding:"required"`
}

// Transaction is a balanced list of postings together with metadata (§2.8.2).
// The sum of amounts across all postings must equal zero.
type Transaction struct {
	Postings      []Posting     `json:"postings"      binding:"required,min=1,dive"`
	TransactionID ForeignBankId `json:"transactionId" binding:"required"`

	Message        string `json:"message"`
	CallNumber     string `json:"callNumber,omitempty"`
	PaymentCode    string `json:"paymentCode"`
	PaymentPurpose string `json:"paymentPurpose"`
}

// CommitTransaction is the body of a COMMIT_TX message (§2.12.2).
type CommitTransaction struct {
	TransactionID ForeignBankId `json:"transactionId" binding:"required"`
}

// RollbackTransaction is the body of a ROLLBACK_TX message (§2.12.3).
type RollbackTransaction struct {
	TransactionID ForeignBankId `json:"transactionId" binding:"required"`
}
