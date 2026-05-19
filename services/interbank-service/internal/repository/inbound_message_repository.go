package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/interbank-service/internal/model"
)

type InboundMessageRepository interface {
	FindByKey(ctx context.Context, peerRoutingNumber int, locallyGeneratedKey string) (*model.InboundMessage, error)
	Save(ctx context.Context, m *model.InboundMessage) error
}
