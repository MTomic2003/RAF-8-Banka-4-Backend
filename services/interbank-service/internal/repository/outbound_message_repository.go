package repository

import (
	"context"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/interbank-service/internal/model"
)

type OutboundMessageRepository interface {
	Enqueue(ctx context.Context, m *model.OutboundMessage) error
	NextBatch(ctx context.Context, limit int) ([]model.OutboundMessage, error)
	MarkSent(ctx context.Context, id uint, status int, body []byte) error
	MarkFailed(ctx context.Context, id uint, lastErr string) error
	Reschedule(ctx context.Context, id uint, attempts int, lastErr string, lastStatus int, lastBody []byte, nextRetryAt time.Time) error
	Cancel(ctx context.Context, id uint) error
}
