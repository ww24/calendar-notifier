//+build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/ww24/calendar-notifier/domain/repository"
	"github.com/ww24/calendar-notifier/domain/service"
	"github.com/ww24/calendar-notifier/interface/action"
	"github.com/ww24/calendar-notifier/interface/calendar"
	"github.com/ww24/calendar-notifier/interface/http/handler"
	"github.com/ww24/calendar-notifier/usecase"
)

func initialize(ctx context.Context, cnf repository.Config) (*app, error) {
	wire.Build(
		wire.Bind(new(repository.Calendar), new(*calendar.Calendar)),
		calendar.New,
		wire.Bind(new(repository.ActionConfigurator), new(*action.Action)),
		action.New,
		service.NewConfig,
		service.NewSynchronizer,
		usecase.NewSynchronizer,
		handler.New,
		newApp,
	)
	return nil, nil
}
