package scheduler

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ww24/calendar-notifier/domain/model"
)

const (
	delimiter = ":"
)

var (
	// ErrAlreadyExists is error that scheduled event already exists.
	ErrAlreadyExists = errors.New("already exists")
)

// InMemory implements in-memory scheduler.
type InMemory struct {
	parent   context.Context
	eventMap sync.Map
}

// NewInMemory returns in-memory scheduler.
func NewInMemory(ctx context.Context) *InMemory {
	return &InMemory{
		parent: ctx,
	}
}

type scheduledEvent struct {
	actionName model.ActionName
	event      model.ScheduleEvent
	handler    func(context.Context) error
	cancel     context.CancelFunc
	sync.Mutex
}

func (e *scheduledEvent) key() string {
	return prefix(e.actionName) + e.event.ID(delimiter)
}

func (e *scheduledEvent) register(ctx context.Context) {
	e.Lock()
	defer e.Unlock()
	ctx, e.cancel = context.WithCancel(ctx)
	timer := time.NewTimer(time.Until(e.event.ExecuteAt))
	go func() {
		defer timer.Stop()
		select {
		case <-ctx.Done():
			log.Println("schedule event canceled:", e.key())
		case <-timer.C:
			if err := e.handler(ctx); err != nil {
				log.Println("schedule event execute error:", err)
			}
		}
	}()
}

func (e *scheduledEvent) unregister() {
	e.Lock()
	defer e.Unlock()
	if e.cancel != nil {
		e.cancel()
	}
	e.cancel = nil
}

func prefix(an model.ActionName) string {
	return string(an) + delimiter
}

// List lists schedule events from in-memory scheduler.
func (s *InMemory) List(an model.ActionName) (model.ScheduleEvents, error) {
	p := prefix(an)
	res := make(model.ScheduleEvents, 0)
	now := time.Now()
	s.eventMap.Range(func(key, value interface{}) bool {
		e := value.(*scheduledEvent)
		if now.After(e.event.ExecuteAt) {
			s.eventMap.Delete(key)
			return true
		}
		if strings.HasPrefix(e.key(), p) {
			res = append(res, e.event)
		}
		return true
	})
	return res, nil
}

// Register registers schedule event to in-memory scheduler.
func (s *InMemory) Register(an model.ActionName, event model.ScheduleEvent, handler func(context.Context) error) error {
	e := &scheduledEvent{actionName: an, event: event, handler: handler}
	if _, loaded := s.eventMap.LoadOrStore(e.key(), e); loaded {
		return ErrAlreadyExists
	}
	e.register(s.parent)
	return nil
}

// Unregister unregisters schedule events from in-memory scheduler.
func (s *InMemory) Unregister(an model.ActionName, events ...model.ScheduleEvent) error {
	for _, event := range events {
		e := &scheduledEvent{actionName: an, event: event}
		v, ok := s.eventMap.Load(e.key())
		if !ok {
			continue
		}
		e = v.(*scheduledEvent)
		e.unregister()
		s.eventMap.Delete(e.key())
	}
	return nil
}
