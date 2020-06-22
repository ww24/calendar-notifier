package main

import (
	"context"
	"log"
	"strings"

	"github.com/ww24/calendar-notifier"
	"github.com/ww24/calendar-notifier/action"
	"github.com/ww24/calendar-notifier/config"
	"golang.org/x/sync/errgroup"
)

func newActionHander(actions map[config.ActionName]config.Action) (map[config.ActionName]action.Action, error) {
	actionHandler := make(map[config.ActionName]action.Action, len(actions))
	for k, a := range actions {
		switch a.Type {
		case config.ActionHTTP:
			actionHandler[k] = action.NewHTTP(a.Header, a.Method, a.URL, a.Payload)
		case config.ActionPubSub:
			act, err := action.NewPubSub(a.Topic, a.Payload)
			if err != nil {
				return nil, err
			}
			actionHandler[k] = act
		case config.ActionTasks:
			act, err := action.NewTasks(a.Location, a.Queue, a.TaskIDPrefix, a.ServiceAccountEmail, a.Method, a.URL, a.Header, a.Payload, k)
			if err != nil {
				return nil, err
			}
			actionHandler[k] = act
		}
	}
	return actionHandler, nil
}

func newExecutorHandler(handler map[string]config.EventHandler, actionHandler map[config.ActionName]action.Action) func(context.Context, *calendar.EventItem) {
	return func(ctx context.Context, e *calendar.EventItem) {
		handleEventItem(e, handler, actionHandler, func(a action.Action) error {
			return a.Exec(ctx, e)
		})
	}
}

func newRegistratorHandler(handler map[string]config.EventHandler, actionHandler map[config.ActionName]action.Action) func(context.Context, *calendar.EventItem) {
	return func(ctx context.Context, e *calendar.EventItem) {
		handleEventItem(e, handler, actionHandler, func(a action.Action) error {
			return a.Register(ctx, e)
		})
	}
}

func handleEventItem(e *calendar.EventItem, handler map[string]config.EventHandler, actionHandler map[config.ActionName]action.Action, h func(action.Action) error) {
	name := strings.TrimSpace(e.Summary)
	eh, ok := handler[name]
	if !ok {
		log.Printf("skipped because handler not exists: %s\n", name)
		return
	}
	var ans []config.ActionName
	switch e.EventType {
	case calendar.Start:
		ans = eh.Start
	case calendar.End:
		ans = eh.End
	default:
		return
	}
	var g errgroup.Group
	defer g.Wait()
	for _, an := range ans {
		an := an
		g.Go(func() error {
			if err := h(actionHandler[an]); err != nil {
				log.Printf("Exec error[%v]: %+v\n", an, err)
				return err
			}
			return nil
		})
	}
}
