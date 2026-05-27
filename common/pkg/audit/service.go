package audit

import (
	"context"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, actionType string, performedByID uint, details string) error {
	return s.repo.Save(ctx, &AuditLog{
		ActionType:    actionType,
		PerformedByID: performedByID,
		Details:       details,
	})
}

func (s *Service) GetAll(ctx context.Context, actionType string, performedByID *uint, dateFrom, dateTo *time.Time, page, pageSize int) ([]AuditLog, int64, error) {
	return s.repo.GetAll(ctx, actionType, performedByID, dateFrom, dateTo, page, pageSize)
}
