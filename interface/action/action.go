package action

import (
	"context"
	"fmt"
	"sync"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/domain/repository"
	"github.com/ww24/calendar-notifier/interface/action/tasks"
)

// Action represents notify actions.
type Action struct {
	parent   context.Context
	tasksCli *tasks.Client
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
	case model.ActionTasks:
		return a.configureTasksAction(ac)
	}

	return nil, fmt.Errorf("Not implemented: %s", ac.Type)
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
