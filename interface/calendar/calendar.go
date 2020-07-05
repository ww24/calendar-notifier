package calendar

import (
	"context"
	"fmt"
	"log"
	"time"

	calendar "google.golang.org/api/calendar/v3"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/domain/repository"
)

// Calendar is calendar API wrapper.
type Calendar struct {
	calendarID string
	newService func(ctx context.Context) (*calendar.Service, error)
}

// New returns new calendar API wrapper.
func New(config repository.Config) *Calendar {
	return &Calendar{
		calendarID: config.Calendar(),
		newService: func(ctx context.Context) (*calendar.Service, error) {
			return calendar.NewService(ctx)
		},
	}
}

// List lists schedules from google calendar.
func (c *Calendar) List(ctx context.Context, since, until time.Time) (model.Schedules, error) {
	svc, err := c.newService(ctx)
	if err != nil {
		return nil, err
	}
	events, err := svc.Events.List(c.calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(since.Format(time.RFC3339)).
		TimeMax(until.Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}

	schedules := make([]model.Schedule, 0, len(events.Items))
	for _, item := range events.Items {
		s, err := toModelSchedule(item)
		if err != nil {
			log.Printf("Warn: %+v\n", err)
			continue
		}
		if s.StartAt.IsZero() || s.EndAt.IsZero() {
			continue
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

func toModelSchedule(item *calendar.Event) (model.Schedule, error) {
	s := model.Schedule{
		ID:          item.Id,
		Summary:     item.Summary,
		Description: item.Description,
	}
	t, err := parseDate(item.Start.DateTime)
	if err != nil {
		return model.Schedule{}, err
	}
	s.StartAt = t
	t, err = parseDate(item.End.DateTime)
	if err != nil {
		return model.Schedule{}, err
	}
	s.EndAt = t
	return s, nil
}

func parseDate(dt string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, dt)
	if err != nil {
		return time.Time{}, fmt.Errorf("Unable parse %v: %w", dt, err)
	}
	return t, nil
}
