//go:generate mockgen -source=$GOFILE -destination=../../mock/mock_$GOPACKAGE/mock_$GOFILE -package=mock_repository

package repository

import (
	"context"

	"github.com/ww24/calendar-notifier/domain/model"
)

// ActionConfigurator is the interface to configure action.
type ActionConfigurator interface {
	Configure(model.ActionConfig) (Action, error)
}

// Action is the interface to register schedule event to action.
type Action interface {
	List(context.Context) (model.ScheduleEvents, error)
	Register(context.Context, ...model.ScheduleEvent) error
	Unregister(context.Context, ...model.ScheduleEvent) error
}
