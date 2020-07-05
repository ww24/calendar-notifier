package usecase

import (
	"context"
	"errors"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/domain/service"
	"github.com/ww24/calendar-notifier/internal/time"
)

// Synchronizer is schedule synchronizer service.
type Synchronizer interface {
	RunningMode() model.RunningMode
	Sync(context.Context) error
	Worker(ctx context.Context) error
}

// NewSynchronizer returns synchronizer.
func NewSynchronizer(cnf service.Config, sync service.Synchronizer) Synchronizer {
	return &synchronizer{
		cnf:  cnf,
		sync: sync,
	}
}

type synchronizer struct {
	cnf  service.Config
	sync service.Synchronizer
}

func (s *synchronizer) RunningMode() model.RunningMode {
	return s.cnf.RunningMode()
}

func (s *synchronizer) Sync(ctx context.Context) error {
	if s.cnf.RunningMode() == model.ModeResident {
		return errors.New("launch handler is unavailable if running mode is resident")
	}
	return s.sync.Sync(ctx)
}

// Worker launchs worker and blocking until context canceled if running mode is resident.
func (s *synchronizer) Worker(ctx context.Context) error {
	if s.cnf.RunningMode() != model.ModeResident {
		return nil
	}
	ticker := time.NewImmediateTicker(s.cnf.SyncInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.sync.Sync(ctx); err != nil {
				return err
			}
		}
	}
}
