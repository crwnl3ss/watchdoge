package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/crwnl3ss/watchdoge/pkg/check"
	"github.com/crwnl3ss/watchdoge/pkg/metrics"
)

type LongRunningCheck struct{}

func (c *LongRunningCheck) SetUp() (func() error, error) {
	return nil, nil
}

func (LongRunningCheck) Check(ctx context.Context) ([]metrics.Metric, error) {
	time.Sleep(time.Second * 5)
	select {
	case <-time.After(time.Second * 5):
	case <-ctx.Done():
		return nil, nil
	}
	return nil, nil
}

func TestSchedulerRunner(t *testing.T) {
	t.Parallel()

	s := NewSheduler(time.Second, []check.Checker{&LongRunningCheck{}})
	if err := s.SetUpCheckers(); err != nil {
		t.Fatalf("failed to setup checks %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := s.Run(ctx)
	if err != nil {
		t.Fatalf("failed to run scheduler: %v", err)
	}
	t.Logf("scheduler runs counter %d, checks counter: %d, done counter: %d", s.ticks, s.Registry[0].runsCounter, s.Registry[0].doneCounter)
}
