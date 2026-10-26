package scheduler

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerRunsDueServicesWithoutOverlap(t *testing.T) {
	var mu sync.Mutex
	runs := map[string]int{}
	var concurrent, maxConcurrent int32

	run := func(_ context.Context, id string) error {
		n := atomic.AddInt32(&concurrent, 1)
		for {
			m := atomic.LoadInt32(&maxConcurrent)
			if n <= m || atomic.CompareAndSwapInt32(&maxConcurrent, m, n) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		mu.Lock()
		runs[id]++
		mu.Unlock()
		atomic.AddInt32(&concurrent, -1)
		return nil
	}
	due := func(context.Context, int) ([]string, error) { return []string{"a", "b", "c"}, nil }
	cleanup := func(context.Context, time.Time) error { return nil }

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	New(due, run, cleanup, 2, 10*time.Millisecond, slog.New(slog.NewTextHandler(io.Discard, nil))).Run(ctx)

	mu.Lock()
	defer mu.Unlock()
	for _, id := range []string{"a", "b", "c"} {
		if runs[id] == 0 {
			t.Errorf("service %s was never checked", id)
		}
	}
	if maxConcurrent > 2 {
		t.Errorf("worker pool exceeded: %d concurrent checks", maxConcurrent)
	}
}
