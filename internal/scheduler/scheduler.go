package scheduler

import (
	"context"

	"github.com/ww24/calendar-notifier/domain/model"
)

// Scheduler represents event scheduler.
type Scheduler interface {
	List(model.ActionName) (model.ScheduleEvents, error)
	Register(an model.ActionName, event model.ScheduleEvent, handler func(context.Context) error) error
	Unregister(model.ActionName, ...model.ScheduleEvent) error
}
