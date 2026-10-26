package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Scheduler dispatches due checks to a bounded pool of workers (CU-03 RN-5): a
// slow check only occupies one worker, and a service is never checked twice at
// the same time.
type Scheduler struct {
	due     func(ctx context.Context, limit int) ([]string, error)
	run     func(ctx context.Context, serviceID string) error
	cleanUp func(ctx context.Context, now time.Time) error
	workers int
	tick    time.Duration
	log     *slog.Logger

	mu       sync.Mutex
	inFlight map[string]bool
}

func New(
	due func(ctx context.Context, limit int) ([]string, error),
	run func(ctx context.Context, serviceID string) error,
	cleanUp func(ctx context.Context, now time.Time) error,
	workers int, tick time.Duration, log *slog.Logger,
) *Scheduler {
	if workers < 1 {
		workers = 1
	}
	return &Scheduler{due: due, run: run, cleanUp: cleanUp, workers: workers, tick: tick, log: log, inFlight: map[string]bool{}}
}

// Run blocks until ctx is cancelled (SIGINT/SIGTERM from main) and then waits
// for the workers to finish their current check.
func (s *Scheduler) Run(ctx context.Context) {
	jobs := make(chan string, s.workers*4)
	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				if err := s.run(ctx, id); err != nil && ctx.Err() == nil {
					s.log.Error("health check failed", "service_id", id, "error", err)
				}
				s.release(id)
			}
		}()
	}

	ticker := time.NewTicker(s.tick)
	cleanup := time.NewTicker(time.Hour)
	defer ticker.Stop()
	defer cleanup.Stop()
	s.log.Info("health check scheduler started", "workers", s.workers, "tick", s.tick.String())

	s.dispatch(ctx, jobs)
	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			s.log.Info("health check scheduler stopped")
			return
		case <-ticker.C:
			s.dispatch(ctx, jobs)
		case now := <-cleanup.C:
			if err := s.cleanUp(ctx, now); err != nil && ctx.Err() == nil {
				s.log.Error("health check cleanup failed", "error", err)
			}
		}
	}
}

func (s *Scheduler) dispatch(ctx context.Context, jobs chan<- string) {
	ids, err := s.due(ctx, s.workers*4)
	if err != nil {
		if ctx.Err() == nil {
			s.log.Error("list due services failed", "error", err)
		}
		return
	}
	for _, id := range ids {
		if !s.claim(id) {
			continue
		}
		select {
		case jobs <- id:
		case <-ctx.Done():
			s.release(id)
			return
		default:
			// Pool saturado: se reintenta en el próximo tick.
			s.release(id)
		}
	}
}

func (s *Scheduler) claim(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inFlight[id] {
		return false
	}
	s.inFlight[id] = true
	return true
}

func (s *Scheduler) release(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.inFlight, id)
}
