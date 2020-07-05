package action

import (
	"context"
	"fmt"
	"sync"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/domain/repository"
	"github.com/ww24/calendar-notifier/interface/action/http"
	"github.com/ww24/calendar-notifier/interface/action/pubsub"
	"github.com/ww24/calendar-notifier/interface/action/tasks"
)

// Action represents notify actions.
type Action struct {
	parent    context.Context
	tasksCli  *tasks.Client
	pubsubCli *pubsub.Client
	httpCli   *http.Client
	sync.Mutex
}

// New returns action.
func New(ctx context.Context) (*Action, error) {
	return &Action{parent: ctx}, nil
}

// Configure returns action from action config.
func (a *Action) Configure(ac model.ActionConfig) (repository.Action, error) {
	a.Lock()
	defer a.Unlock()

	switch ac.Type {
	case model.ActionHTTP:
		return a.configureHTTPAction(ac)
	case model.ActionPubSub:
		return a.configurePubSubAction(ac)
	case model.ActionTasks:
		return a.configureTasksAction(ac)
	}

	return nil, fmt.Errorf("Not implemented: %s", ac.Type)
}

func (a *Action) configureHTTPAction(ac model.ActionConfig) (repository.Action, error) {
	if a.httpCli == nil {
		cli, err := http.NewClient(a.parent)
		if err != nil {
			return nil, err
		}
		a.httpCli = cli
	}
	return http.New(a.httpCli, ac), nil
}

func (a *Action) configurePubSubAction(ac model.ActionConfig) (repository.Action, error) {
	if a.pubsubCli == nil {
		cli, err := pubsub.NewClient(a.parent)
		if err != nil {
			return nil, err
		}
		a.pubsubCli = cli
	}
	return pubsub.New(a.pubsubCli, ac), nil
}

func (a *Action) configureTasksAction(ac model.ActionConfig) (repository.Action, error) {
	if a.tasksCli == nil {
		cli, err := tasks.NewClient(a.parent)
		if err != nil {
			return nil, err
		}
		a.tasksCli = cli
	}
	return tasks.New(a.tasksCli, ac), nil
}
