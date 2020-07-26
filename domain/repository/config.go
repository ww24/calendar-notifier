package repository

import (
	"time"

	"github.com/ww24/calendar-notifier/domain/model"
)

// Config is the interface to get configurations.
type Config interface {
	ActionNames(model.ScheduleEvent) ([]model.ActionName, bool)
	ActionConfigMap() map[model.ActionName]model.ActionConfig
	RunningMode() model.RunningMode
	SyncInterval() time.Duration
	Calendar() string
	CalendarWebhookURL() (string, bool)
}
