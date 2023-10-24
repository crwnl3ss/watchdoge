package check

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"time"

	"github.com/crwnl3ss/watchdoge/pkg/metrics"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

var _ Checker = &KafkaCheck{}

type KafkaRWCheck struct {
	Topic   string        `yaml:"topic" validate:"non-empty"`
	Host    string        `yaml:"host"`
	Timeout time.Duration `yaml:"timeout" validate:"non-empty"`
}

type KafkaOptions struct {
	Read     *KafkaRWCheck `yaml:"read"`
	Write    *KafkaRWCheck `yaml:"write"`
	User     string        `yaml:"user"`
	Password string        `yaml:"password"`
}

type KafkaCheck struct {
	Name    string `validate:"non-empty"`
	Type    CheckerType
	Comment string
	Jitter  time.Duration
	Options *KafkaOptions `yaml:"options" validate:"non-nil"`
	writer  *kafka.Writer
	reader  *kafka.Reader
}

func NewKafkaCheck(name, comment string) *KafkaCheck {
	return &KafkaCheck{
		Type:    Kafka,
		Name:    name,
		Comment: comment,
	}
}

func (k *KafkaCheck) SetUp() (func() error, error) {
	mechanism, err := scram.Mechanism(
		scram.SHA256,
		k.Options.User,
		k.Options.Password,
	)
	if err != nil {
		log.Fatalln(err)
	}

	dialer := &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}
	k.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{k.Options.Write.Host},
		Dialer:  dialer,
		Topic:   k.Options.Write.Topic,
	})
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{k.Options.Read.Host},
		GroupID: "$GROUP_NAME",
		Topic:   k.Options.Read.Topic,
		Dialer:  dialer,
	})
	tearDown := func() error {
		return errors.Join(k.writer.Close(), k.reader.Close())
	}
	return tearDown, nil
}

// TODO: pointer
func (k *KafkaCheck) Check(ctx context.Context) ([]metrics.Metric, error) {
	results := make([]metrics.Metric, 0)
	k.writer.BatchSize = 1

	log.Printf("write timeout: %s, read timeout: %s", k.Options.Write.Timeout, k.Options.Read.Timeout)

	before := time.Now()
	if err := k.writer.WriteMessages(ctx, kafka.Message{
		Value: []byte("ping"),
	}); err != nil {
		return results, err
	}
	results = append(results, metrics.Metric{
		Name:  k.Name + "_consumer_ms",
		Value: time.Since(before).Milliseconds(),
	})

	before = time.Now()
	msg, err := k.reader.ReadMessage(ctx)
	if err != nil {
		return results, err
	}
	results = append(results, metrics.Metric{
		Name:  k.Name + "_producer_ms",
		Value: int64(time.Since(before).Milliseconds()),
	})
	log.Printf("successfull read message: %s", string(msg.Value))
	return results, nil
}
