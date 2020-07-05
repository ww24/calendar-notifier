package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
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

// Schedule is calendar schedule item.
type Schedule struct {
	ID          string
	Summary     string
	Description string
	StartAt     time.Time
	EndAt       time.Time
}

// Events returns schedule events from schedule.
func (s *Schedule) Events(t time.Time) ScheduleEvents {
	if t.After(s.EndAt) {
		return nil
	}
	if t.After(s.StartAt) {
		return []ScheduleEvent{s.EndEvent()}
	}
	return []ScheduleEvent{s.StartEvent(), s.EndEvent()}
}

// StartEvent returns start event of schedule.
func (s *Schedule) StartEvent() ScheduleEvent {
	return ScheduleEvent{
		ScheduleID:  s.ID,
		Summary:     s.Summary,
		Description: s.Description,
		EventType:   Start,
		ExecuteAt:   s.StartAt,
	}
}

// EndEvent returnsend event of schedule.
func (s *Schedule) EndEvent() ScheduleEvent {
	return ScheduleEvent{
		ScheduleID:  s.ID,
		Summary:     s.Summary,
		Description: s.Description,
		EventType:   End,
		ExecuteAt:   s.EndAt,
	}
}

// Schedules represents schedule slice.
type Schedules []Schedule

// Events returns schedule events from schedules.
func (ss Schedules) Events(t time.Time) ScheduleEvents {
	events := make([]ScheduleEvent, 0, len(ss)*2)
	for _, s := range ss {
		events = append(events, s.Events(t)...)
	}
	return events
}

// ScheduleEvent is (start or end) event of schedule.
type ScheduleEvent struct {
	ScheduleID  string
	Summary     string
	Description string
	EventType   EventType
	ExecuteAt   time.Time
}

// ID returns schedule event id.
func (s *ScheduleEvent) ID(delimiter string) string {
	if delimiter == "" {
		delimiter = ":"
	}
	return fmt.Sprintf("%s%s%d", s.ScheduleID, delimiter, s.ExecuteAt.Unix())
}

// ParseID parses ID and returns ScheduleID.
func (s *ScheduleEvent) ParseID(id, delimiter string) string {
	return strings.TrimSuffix(id, delimiter+strconv.FormatInt(s.ExecuteAt.Unix(), 10))
}

// ScheduleEvents represents schedule event slice.
type ScheduleEvents []ScheduleEvent

// SortByExecuteAtAsc sorts schedule events by ExecuteAt ascending.
func (ss ScheduleEvents) SortByExecuteAtAsc() {
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].ExecuteAt.Before(ss[j].ExecuteAt)
	})
}

// Sub returns schedule events ss-tt.
func (ss ScheduleEvents) Sub(tt ScheduleEvents) ScheduleEvents {
	m := tt.toMap()
	events := make([]ScheduleEvent, 0, len(ss))
	for _, s := range ss {
		if _, ok := m[s.ID("")]; !ok {
			events = append(events, s)
		}
	}
	return events
}

func (ss ScheduleEvents) toMap() map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s.ID("")] = struct{}{}
	}
	return m
}
