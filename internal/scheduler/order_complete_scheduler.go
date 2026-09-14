package scheduler

import (
	"context"
	"log"
	"time"

	"jiaozi/internal/orderapp"
)

type OrderCompleteScheduler struct {
	service  *orderapp.Service
	interval time.Duration
}

func NewOrderCompleteScheduler(service *orderapp.Service, interval time.Duration) *OrderCompleteScheduler {
	return &OrderCompleteScheduler{
		service:  service,
		interval: interval,
	}
}

func (s *OrderCompleteScheduler) Start(ctx context.Context) {
	go func() {
		s.completeArrivedOrders(ctx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.completeArrivedOrders(ctx)
			}
		}
	}()
}

func (s *OrderCompleteScheduler) completeArrivedOrders(ctx context.Context) {
	completed, err := s.service.CompleteArrivedOrders(ctx, time.Now())
	if err != nil {
		log.Printf("complete arrived orders failed: %v", err)
		return
	}
	if completed > 0 {
		log.Printf("completed arrived orders: %d", completed)
	}
}
