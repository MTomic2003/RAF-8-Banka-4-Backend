package dto

// MessageType discriminates the payload of a Message envelope (§2.12).
type MessageType string

const (
	MessageTypeNewTx      MessageType = "NEW_TX"
	MessageTypeCommitTx   MessageType = "COMMIT_TX"
	MessageTypeRollbackTx MessageType = "ROLLBACK_TX"
)

type MessageEnvelope struct {
	IdempotenceKey IdempotenceKey `json:"idempotenceKey" binding:"required"`
	MessageType    MessageType    `json:"messageType"    binding:"required,oneof=NEW_TX COMMIT_TX ROLLBACK_TX"`
}

// NewTxMessage is the wire shape for a NEW_TX message: a Transaction body
// to be locally prepared and voted on.
type NewTxMessage struct {
	IdempotenceKey IdempotenceKey `json:"idempotenceKey" binding:"required"`
	MessageType    MessageType    `json:"messageType"    binding:"required"`
	Message        Transaction    `json:"message"        binding:"required"`
}

// CommitTxMessage is the wire shape for a COMMIT_TX message: the id of a
// previously-prepared transaction that is now to be committed.
type CommitTxMessage struct {
	IdempotenceKey IdempotenceKey    `json:"idempotenceKey" binding:"required"`
	MessageType    MessageType       `json:"messageType"    binding:"required"`
	Message        CommitTransaction `json:"message"        binding:"required"`
}

// RollbackTxMessage is the wire shape for a ROLLBACK_TX message: the id of a
// previously-prepared transaction whose reservations are to be released.
type RollbackTxMessage struct {
	IdempotenceKey IdempotenceKey      `json:"idempotenceKey" binding:"required"`
	MessageType    MessageType         `json:"messageType"    binding:"required"`
	Message        RollbackTransaction `json:"message"        binding:"required"`
}

// VoteKind is the outcome of a transaction vote.
type VoteKind string

const (
	VoteYes VoteKind = "YES"
	VoteNo  VoteKind = "NO"
)

// TransactionVote is the response body for a NEW_TX message (§2.12.1).
type TransactionVote struct {
	Vote    VoteKind       `json:"vote"`
	Reasons []NoVoteReason `json:"reasons,omitempty"`
}

// NoVoteReasonKind enumerates protocol-defined reasons for a NO vote.
type NoVoteReasonKind string

const (
	ReasonUnbalancedTx              NoVoteReasonKind = "UNBALANCED_TX"
	ReasonNoSuchAccount             NoVoteReasonKind = "NO_SUCH_ACCOUNT"
	ReasonNoSuchAsset               NoVoteReasonKind = "NO_SUCH_ASSET"
	ReasonUnacceptableAsset         NoVoteReasonKind = "UNACCEPTABLE_ASSET"
	ReasonInsufficientAsset         NoVoteReasonKind = "INSUFFICIENT_ASSET"
	ReasonOptionAmountIncorrect     NoVoteReasonKind = "OPTION_AMOUNT_INCORRECT"
	ReasonOptionUsedOrExpired       NoVoteReasonKind = "OPTION_USED_OR_EXPIRED"
	ReasonOptionNegotiationNotFound NoVoteReasonKind = "OPTION_NEGOTIATION_NOT_FOUND"
)

// NoVoteReason carries a reason for a NO vote. For UNBALANCED_TX the
// Posting is omitted; for all other reasons it points to the offending
// posting.
type NoVoteReason struct {
	Reason  NoVoteReasonKind `json:"reason"`
	Posting *Posting         `json:"posting,omitempty"`
}
