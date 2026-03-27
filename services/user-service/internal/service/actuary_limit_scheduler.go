package service

import (
	"context"
	"log"
	"sync"
	"time"
)

type ActuaryLimitScheduler struct {
	actuaryService *ActuaryService

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewActuaryLimitScheduler(actuaryService *ActuaryService) *ActuaryLimitScheduler {
	return &ActuaryLimitScheduler{actuaryService: actuaryService}
}

func (s *ActuaryLimitScheduler) Start() {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	go s.runDailyReset(ctx)
}

func (s *ActuaryLimitScheduler) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (s *ActuaryLimitScheduler) runDailyReset(ctx context.Context) {
	for {
		timer := time.NewTimer(time.Until(nextActuaryReset()))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if err := s.actuaryService.ResetAllUsedLimits(ctx); err != nil {
				log.Printf("[ActuaryLimitScheduler] reset failed: %v", err)
			}
		}
	}
}

func nextActuaryReset() time.Time {
	now := time.Now()
	resetAt := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
	if !resetAt.After(now) {
		resetAt = resetAt.Add(24 * time.Hour)
	}

	return resetAt
}
