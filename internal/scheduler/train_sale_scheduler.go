package scheduler

import (
	"context"
	"log"
	"time"

	"jiaozi/internal/train"
)

type TrainSaleScheduler struct {
	service  *train.Service
	interval time.Duration
}

func NewTrainSaleScheduler(service *train.Service, interval time.Duration) *TrainSaleScheduler {
	return &TrainSaleScheduler{
		service:  service,
		interval: interval,
	}
}

func (s *TrainSaleScheduler) Start(ctx context.Context) {
	go func() {
		s.openDueTrains(ctx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.openDueTrains(ctx)
			}
		}
	}()
}

func (s *TrainSaleScheduler) openDueTrains(ctx context.Context) {
	opened, err := s.service.OpenDueTrains(ctx, time.Now())
	if err != nil {
		log.Printf("open due trains failed: %v", err)
		return
	}
	if opened > 0 {
		log.Printf("opened due trains: %d", opened)
	}
}
