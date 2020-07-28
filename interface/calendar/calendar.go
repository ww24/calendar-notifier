package calendar

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	calendar "google.golang.org/api/calendar/v3"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/domain/repository"
)

// Calendar is calendar API wrapper.
type Calendar struct {
	calendarID string
	newService func(ctx context.Context) (*calendar.Service, error)
	token      syncToken
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

	c.token.update(events.NextSyncToken)

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

// Watch watches google calendar update event.
func (c *Calendar) Watch(ctx context.Context, address string, ttl time.Duration) error {
	svc, err := c.newService(ctx)
	if err != nil {
		return err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate UUIDv4 as channel id: %w", err)
	}

	channel := &calendar.Channel{
		Address: address,
		Id:      id.String(),
		Params: map[string]string{
			"ttl": strconv.Itoa(int(ttl.Seconds())),
		},
		Payload: true,
		Token:   "", // TODO
	}
	watchCall := svc.Events.Watch(c.calendarID, channel).
		Context(ctx).
		SyncToken(c.token.get())

	ch, err := watchCall.Do()
	if err != nil {
		return fmt.Errorf("failed to watch calendar events: %w", err)
	}

	// DEBUG
	fmt.Printf("channel: %+v\n", ch)

	return nil
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
