package calendar

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	calendar "google.golang.org/api/calendar/v3"
)

// EventType represents calendar event type.
//go:generate stringer -type=EventType
type EventType int

const (
	// None is uncategorized event.
	None EventType = iota
	// Start is schedule started event.
	Start
	// End is schedule ended event.
	End
)

// MarshalJSON implements json.Marshaler.
func (t EventType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strings.ToLower(t.String()) + `"`), nil
}

// EventItem is extracted calendar schedule.
type EventItem struct {
	ID          string    `json:"id"`
	EventType   EventType `json:"event_type,string"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	mutex      sync.Mutex
	cancel     func()
	registered bool
}

// Register registeres event handler.
func (e *EventItem) Register(now time.Time, handler func(context.Context, *EventItem)) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	defer func() { e.registered = true }()
	if e.registered {
		return false
	}

	started := now.After(e.StartAt)
	var startTimer *time.Timer
	endTimer := time.NewTimer(e.EndAt.Sub(now))
	if !started {
		startTimer = time.NewTimer(e.StartAt.Sub(now))
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	go func() {
		channels := make([]<-chan time.Time, 2)
		if startTimer != nil {
			defer startTimer.Stop()
			channels[0] = startTimer.C
		} else {
			closed := make(chan time.Time)
			close(closed)
			channels[0] = closed
		}
		defer endTimer.Stop()
		channels[1] = endTimer.C

		for i, ch := range channels {
			switch i {
			case 0:
				e.EventType = Start
			case 1:
				e.EventType = End
			}
			select {
			case _, ok := <-ch:
				if ok {
					handler(ctx, e)
				}
			case <-ctx.Done():
				log.Println("Canceled:", e.ID)
				return
			}
		}
	}()

	return true
}

// Cancel canceles event handler.
func (e *EventItem) Cancel() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.cancel != nil {
		e.cancel()
	}
}

// Calendar is calendar API wrapper.
type Calendar struct {
	calendarID string
	newService func(ctx context.Context) (*calendar.Service, error)
}

// NewCalendar returns new calendar API wrapper.
func NewCalendar(calendarID string) *Calendar {
	return &Calendar{
		calendarID: calendarID,
		newService: func(ctx context.Context) (*calendar.Service, error) {
			return calendar.NewService(ctx)
		},
	}
}

// Events gets calendar events list.
func (c *Calendar) Events(ctx context.Context, since, until time.Time) ([]*EventItem, error) {
	srv, err := c.newService(ctx)
	if err != nil {
		return nil, err
	}
	events, err := srv.Events.List(c.calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(since.Format(time.RFC3339)).
		TimeMax(until.Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}

	res := make([]*EventItem, 0, len(events.Items))
	for _, item := range events.Items {
		e := toEventItem(item)
		if e.StartAt.IsZero() || e.EndAt.IsZero() {
			continue
		}
		res = append(res, e)
	}

	return res, nil
}

func toEventItem(item *calendar.Event) *EventItem {
	e := &EventItem{
		ID:          item.Id,
		Summary:     item.Summary,
		Description: item.Description,
	}
	st, err := time.Parse(time.RFC3339, item.Start.DateTime)
	if err != nil {
		log.Printf("Unable parse %v: %v\n", item.Start.DateTime, err)
	}
	e.StartAt = st
	et, err := time.Parse(time.RFC3339, item.End.DateTime)
	if err != nil {
		log.Printf("Unable parse %v: %v\n", item.End.DateTime, err)
	}
	e.EndAt = et
	cat, err := time.Parse(time.RFC3339, item.Created)
	if err != nil {
		log.Printf("Unable parse %v: %v\n", item.Created, err)
	}
	e.CreatedAt = cat
	uat, err := time.Parse(time.RFC3339, item.Updated)
	if err != nil {
		log.Printf("Unable parse %v: %v\n", item.Updated, err)
	}
	e.UpdatedAt = uat
	return e
}

// ImmediateTick returns channel which wrappes ticker.
func ImmediateTick(ctx context.Context, interval time.Duration) <-chan time.Time {
	c := make(chan time.Time, 0)
	go func() {
		defer close(c)
		c <- time.Now()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case tick := <-t.C:
				c <- tick
			case <-ctx.Done():
				return
			}
		}
	}()
	return c
}
