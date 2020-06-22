package action

import (
	"context"
	"log"
	"time"

	"github.com/ww24/calendar-notifier"
)

const (
	requestTimeout = 15 * time.Second
)

// Action is interface for action behavior.
type Action interface {
	Register(context.Context, *calendar.EventItem) error
	Exec(context.Context, *calendar.EventItem) error
}

type action struct {
	ExecOnSchedule  func(context.Context, *calendar.EventItem) error
	ExecImmediately func(context.Context, *calendar.EventItem) error
}

func (a *action) Register(ctx context.Context, e *calendar.EventItem) error {
	if a.ExecImmediately == nil {
		return nil
	}
	return a.ExecImmediately(ctx, e)
}

func (a *action) Exec(ctx context.Context, e *calendar.EventItem) error {
	if a.ExecOnSchedule == nil {
		return nil
	}
	return a.ExecOnSchedule(ctx, e)
}

// wrapAction wraps scheduled, immediate action.
func wrapAction(a interface{}) Action {
	act := &action{}
	switch a := a.(type) {
	case scheduledAction:
		act.ExecOnSchedule = a.ExecOnSchedule
	case immediateAction:
		act.ExecImmediately = a.ExecImmediately
	default:
		log.Println("Error: action is not implemented")
	}
	return act
}

// scheduledAction provides executor for scheduled action.
type scheduledAction interface {
	ExecOnSchedule(context.Context, *calendar.EventItem) error
}

// immediateAction provides executor for immediate action.
type immediateAction interface {
	ExecImmediately(context.Context, *calendar.EventItem) error
}
