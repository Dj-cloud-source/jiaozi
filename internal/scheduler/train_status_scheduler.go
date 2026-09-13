package scheduler

import (
	"context"
	"log"
	"time"

	"jiaozi/internal/train"
)

type TrainStatusScheduler struct {
	service  *train.Service
	interval time.Duration
}

func NewTrainStatusScheduler(service *train.Service, interval time.Duration) *TrainStatusScheduler {
	return &TrainStatusScheduler{
		service:  service,
		interval: interval,
	}
}

func (s *TrainStatusScheduler) Start(ctx context.Context) {
	go func() {
		s.updateTrainStatus(ctx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.updateTrainStatus(ctx)
			}
		}
	}()
}

func (s *TrainStatusScheduler) updateTrainStatus(ctx context.Context) {
	now := time.Now()
	s.openDueTrains(ctx, now)
	s.stopDueTrains(ctx, now)
}

func (s *TrainStatusScheduler) openDueTrains(ctx context.Context, now time.Time) {
	opened, err := s.service.OpenDueTrains(ctx, now)
	if err != nil {
		log.Printf("open due trains failed: %v", err)
		return
	}
	if opened > 0 {
		log.Printf("opened due trains: %d", opened)
	}
}

func (s *TrainStatusScheduler) stopDueTrains(ctx context.Context, now time.Time) {
	stopped, err := s.service.StopDueTrains(ctx, now)
	if err != nil {
		log.Printf("stop due trains failed: %v", err)
		return
	}
	if stopped > 0 {
		log.Printf("stopped due trains: %d", stopped)
	}
}
