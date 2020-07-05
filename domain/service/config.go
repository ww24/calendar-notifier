package service

import (
	"time"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/domain/repository"
)

// Config is config service.
type Config interface {
	RunningMode() model.RunningMode
	SyncInterval() time.Duration
}

// NewConfig returns config.
func NewConfig(cnf repository.Config) Config {
	return &config{
		cnf: cnf,
	}
}

type config struct {
	cnf repository.Config
}

func (c *config) RunningMode() model.RunningMode {
	return c.cnf.RunningMode()
}

func (c *config) SyncInterval() time.Duration {
	return c.cnf.SyncInterval()
}
