package service

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/interbank-service/internal/dto"
)

// MessageProcessor handles inbound bank-to-bank messages: NEW_TX prepares a
// transaction and casts a vote, COMMIT_TX commits a previously-prepared
// transaction, ROLLBACK_TX releases reservations.
//
// Per §2.11, the int return is the HTTP status (200 with body, 202 / 204
// without) the transport layer should write back.
type MessageProcessor struct{}

func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{}
}

func (p *MessageProcessor) ProcessNewTx(_ context.Context, _ int, _ *dto.Transaction) (int, any, error) {
	// Local preparation is not implemented yet — vote NO with a placeholder
	// reason until real preparation lands.
	return 200, dto.TransactionVote{
		Vote:    dto.VoteNo,
		Reasons: []dto.NoVoteReason{{Reason: dto.ReasonUnbalancedTx}},
	}, nil
}

func (p *MessageProcessor) ProcessCommitTx(_ context.Context, _ int, _ *dto.CommitTransaction) (int, any, error) {
	return 204, nil, nil
}

func (p *MessageProcessor) ProcessRollbackTx(_ context.Context, _ int, _ *dto.RollbackTransaction) (int, any, error) {
	return 204, nil, nil
}
