package scheduler

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/crwnl3ss/watchdoge/pkg/check"
)

type CheckTask struct {
	check      check.Checker
	isSetUp    bool
	tearDownFn func() error
}

func NewCheckTask(check check.Checker) *CheckTask {
	return &CheckTask{
		check:      check,
		isSetUp:    false,
		tearDownFn: func() error { return nil },
	}
}

type Scheduler struct {
	Periodicity time.Duration
	Registry    map[int]*CheckTask
}

func NewSheduler(periodicity time.Duration, checks []check.Checker) *Scheduler {
	registry := make(map[int]*CheckTask)
	for i, c := range checks {
		registry[i] = NewCheckTask(c)
	}
	return &Scheduler{Periodicity: periodicity, Registry: registry}
}

func (s *Scheduler) SetUpCheckers() error {
	var combinedError error
	for _, checkStatus := range s.Registry {
		tearDownFn, err := checkStatus.check.SetUp()
		if err != nil {
			combinedError = errors.Join(combinedError, err)
			if err = s.TearDownCheckers(); err != nil {
				combinedError = errors.Join(combinedError, err)
			}
			if err := tearDownFn(); err != nil {
				combinedError = errors.Join(combinedError, err)
			}
			return combinedError
		} else {
			checkStatus.isSetUp = true
			log.Printf("setup successful %v", checkStatus)
		}
	}
	return combinedError
}

func (s *Scheduler) TearDownCheckers() error {
	var finalError error = nil
	for _, checkStatus := range s.Registry {
		if checkStatus.isSetUp {
			if err := checkStatus.tearDownFn(); err != nil {
				finalError = errors.Join(finalError, err)
			} else {
				log.Printf("%v tear down successful", checkStatus)
			}
		}
	}
	return finalError
}

func (s *Scheduler) run(ctx context.Context) error {
	for k, v := range s.Registry {
		if !v.isSetUp {
			log.Printf("skip %v check, checker is not setupuped", k)
		}
		checkResults, err := v.check.Check(ctx)
		log.Println(checkResults, err)
	}
	return nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	t := time.NewTicker(s.Periodicity)
	defer t.Stop()
	s.run(ctx)
	for {
		select {
		case <-t.C:
			s.run(ctx)
		case <-ctx.Done():
			log.Println("shutting down scheduler...")
			return nil
		}
	}
}
