package scheduler

import (
	"context"
	"log"
	"time"
)

type Scheduler struct {
	st   SchedulerStorage
	tg   Sender
	tick time.Duration
}

func New(st SchedulerStorage, tg Sender, tick time.Duration) *Scheduler {
	return &Scheduler{
		st:   st,
		tg:   tg,
		tick: tick,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	t := time.NewTicker(s.tick)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if err := s.step(ctx); err != nil {
				log.Printf("scheduler step error: %v", err)
			}
		}
	}
}
