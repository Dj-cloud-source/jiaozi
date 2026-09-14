package scheduler

import (
	"context"
	"log"
	"time"

	"jiaozi/internal/orderapp"
)

type OrderTimeoutScheduler struct {
	service  *orderapp.Service
	interval time.Duration
}

func NewOrderTimeoutScheduler(service *orderapp.Service, interval time.Duration) *OrderTimeoutScheduler {
	return &OrderTimeoutScheduler{
		service:  service,
		interval: interval,
	}
}

func (s *OrderTimeoutScheduler) Start(ctx context.Context) {
	go func() {
		s.cancelExpiredOrders(ctx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.cancelExpiredOrders(ctx)
			}
		}
	}()
}

func (s *OrderTimeoutScheduler) cancelExpiredOrders(ctx context.Context) {
	cancelled, err := s.service.CancelExpiredOrders(ctx, time.Now())
	if err != nil {
		log.Printf("cancel expired orders failed: %v", err)
		return
	}
	if cancelled > 0 {
		log.Printf("cancelled expired orders: %d", cancelled)
	}
}
