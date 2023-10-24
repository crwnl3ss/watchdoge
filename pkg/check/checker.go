package check

import (
	"context"

	"github.com/crwnl3ss/watchdoge/pkg/metrics"
)

type CheckerType string

const (
	Kafka CheckerType = "kafka"
	Redis CheckerType = "redis"
)

type Checker interface {
	SetUp() (func() error, error)
	Check(context.Context) ([]metrics.Metric, error)
}
