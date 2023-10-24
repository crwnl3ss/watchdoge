package config

import (
	"bufio"
	"io"
	"time"

	"github.com/crwnl3ss/watchdoge/pkg/check"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Periodicity    time.Duration `yaml:"periodicity"`
	Checkers       []BaseCheck   `yaml:"checks"`
	RunnableChecks []check.Checker
}

type BaseCheck struct {
	Type    check.CheckerType `yaml:"type"`
	Name    string            `yaml:"name"`
	Comment string            `yaml:"comment"`
	Options VariableOptions   `yaml:"options"`
}

// See https://github.com/go-yaml/yaml/issues/13
type VariableOptions struct {
	unmarshalFn func(interface{}) error
}

func (c *VariableOptions) UnmarshalYAML(fn func(interface{}) error) error {
	c.unmarshalFn = fn
	return nil
}

func (c *VariableOptions) UnmarshalToCheckType(val interface{}) error {
	return c.unmarshalFn(val)
}

func LoadConfig(r *bufio.Reader) (*Config, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	cfg := new(Config)
	if err = yaml.Unmarshal(body, cfg); err != nil {
		return nil, err
	}
	for _, cc := range cfg.Checkers {
		switch cc.Type {
		case check.Kafka:
			kc := check.NewKafkaCheck(cc.Name, cc.Comment)
			options := check.KafkaOptions{}
			if err := cc.Options.UnmarshalToCheckType(&options); err != nil {
				return nil, err
			}
			kc.Options = &options
			if err := Validate(kc); err != nil {
				return nil, err
			}
			cfg.RunnableChecks = append(cfg.RunnableChecks, kc)
		}
	}

	return cfg, nil
}
