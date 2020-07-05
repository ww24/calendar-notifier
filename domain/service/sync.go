package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/domain/repository"
)

const calendarScanRange = 24 * time.Hour

// Synchronizer is schedule synchronizer service.
type Synchronizer interface {
	Sync(context.Context) error
}

// NewSynchronizer returns synchronizer.
func NewSynchronizer(
	cnf repository.Config,
	cal repository.Calendar,
	ac repository.ActionConfigurator) Synchronizer {
	return &synchronizer{
		cnf:     cnf,
		cal:     cal,
		ac:      ac,
		timeNow: time.Now,
	}
}

type synchronizer struct {
	cnf     repository.Config
	cal     repository.Calendar
	ac      repository.ActionConfigurator
	timeNow func() time.Time
}

type action struct {
	action repository.Action
	events model.ScheduleEvents
}

func (s *synchronizer) Sync(ctx context.Context) error {
	now := s.timeNow()
	schedules, err := s.cal.List(ctx, now, now.Add(calendarScanRange))
	if err != nil {
		return fmt.Errorf("calendar.List: %w", err)
	}
	log.Println("calendar.List:", len(schedules))

	acm := s.cnf.ActionConfigMap()
	am := make(map[model.ActionName]*action, len(acm))

	if err := s.initialize(ctx, am, acm); err != nil {
		return err
	}
	log.Println("s.initialize:", len(am))

	if err := s.register(ctx, am, schedules.Events(now)); err != nil {
		return err
	}

	if err := s.unregister(ctx, am); err != nil {
		return err
	}

	return nil
}

func (s *synchronizer) initialize(ctx context.Context, am map[model.ActionName]*action, acm map[model.ActionName]model.ActionConfig) error {
	for an, ac := range acm {
		a, err := s.ac.Configure(ac)
		if err != nil {
			return fmt.Errorf("actionConfig.Configure: %w", err)
		}
		events, err := a.List(ctx)
		if err != nil {
			return fmt.Errorf("action.List: %w", err)
		}
		am[an] = &action{
			action: a,
			events: events,
		}
	}
	return nil
}

func (s *synchronizer) register(ctx context.Context, am map[model.ActionName]*action, events model.ScheduleEvents) error {
	for actionName, events := range s.route(events) {
		act, ok := am[actionName]
		if !ok {
			log.Println("action not defined:", actionName)
			continue
		}
		if err := act.action.Register(ctx, events.Sub(act.events)...); err != nil {
			return fmt.Errorf("action.Register: %w", err)
		}
		if len(events.Sub(act.events)) > 0 {
			log.Printf("action.Register[%s]: %d\n", actionName, len(events.Sub(act.events)))
		}
		act.events = act.events.Sub(events)
	}
	return nil
}

func (s *synchronizer) unregister(ctx context.Context, am map[model.ActionName]*action) error {
	for actionName, act := range am {
		if len(act.events) == 0 {
			continue
		}
		if err := act.action.Unregister(ctx, act.events...); err != nil {
			return fmt.Errorf("action.Unregister: %w", err)
		}
		if len(act.events) > 0 {
			log.Printf("action.Unegister[%s]: %d\n", actionName, len(act.events))
		}
	}
	return nil
}

func (s *synchronizer) route(events []model.ScheduleEvent) map[model.ActionName]model.ScheduleEvents {
	routedEvents := make(map[model.ActionName]model.ScheduleEvents)
	for _, e := range events {
		actionNames, ok := s.cnf.ActionNames(e)
		if !ok {
			log.Println("schedule event not defined:", e.Summary)
			continue
		}
		for _, actionName := range actionNames {
			if routedEvents[actionName] == nil {
				routedEvents[actionName] = make([]model.ScheduleEvent, 0, 1)
			}
			routedEvents[actionName] = append(routedEvents[actionName], e)
		}
	}
	return routedEvents
}
