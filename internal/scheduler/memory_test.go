package scheduler

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ww24/calendar-notifier/domain/model"
)

func TestScheduler_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testTime := time.Now()
	tests := []struct {
		name          string
		initialEvents []*scheduledEvent
		actionName    model.ActionName
		want          model.ScheduleEvents
	}{
		{
			name:          "empty",
			initialEvents: []*scheduledEvent{},
			actionName:    "test",
			want:          model.ScheduleEvents{},
		},
		{
			name: "select by action name",
			initialEvents: []*scheduledEvent{
				{
					actionName: "hoge",
					event: model.ScheduleEvent{
						ScheduleID:  "sid1",
						Summary:     "hoge",
						Description: "hoge schedule",
						EventType:   model.End,
						ExecuteAt:   testTime.Add(time.Hour),
					},
				}, {
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.Start,
						ExecuteAt:   testTime.Add(2 * time.Hour),
					},
				}, {
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.End,
						ExecuteAt:   testTime.Add(3 * time.Hour),
					},
				},
			},
			actionName: "test",
			want: model.ScheduleEvents{
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.Start,
					ExecuteAt:   testTime.Add(2 * time.Hour),
				},
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.End,
					ExecuteAt:   testTime.Add(3 * time.Hour),
				},
			},
		},
		{
			name: "old schedule event is removed",
			initialEvents: []*scheduledEvent{
				{
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.Start,
						ExecuteAt:   testTime.Add(-time.Hour),
					},
				}, {
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.End,
						ExecuteAt:   testTime.Add(time.Hour),
					},
				},
			},
			actionName: "test",
			want: model.ScheduleEvents{
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.End,
					ExecuteAt:   testTime.Add(time.Hour),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemory(ctx)
			for _, e := range tt.initialEvents {
				s.eventMap.Store(e.key(), e)
			}

			got, err := s.List(tt.actionName)
			if err != nil {
				t.Fatalf("err should be nil but got %+v", err)
			}
			got.SortByExecuteAtAsc()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("\nwant: %+v\n got: %+v", tt.want, got)
			}
		})
	}
}

func TestScheduler_Register(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testTime := time.Now()
	handler := func(context.Context) error { return nil }
	tests := []struct {
		name          string
		initialEvents []*scheduledEvent
		actionName    model.ActionName
		event         model.ScheduleEvent
		want          model.ScheduleEvents
		wantErr       error
	}{
		{
			name: "already registered",
			initialEvents: []*scheduledEvent{
				{
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.Start,
						ExecuteAt:   testTime.Add(time.Hour),
					},
					handler: handler,
				},
			},
			actionName: "test",
			event: model.ScheduleEvent{
				ScheduleID:  "sid2",
				Summary:     "test",
				Description: "test schedule",
				EventType:   model.Start,
				ExecuteAt:   testTime.Add(time.Hour),
			},
			want: model.ScheduleEvents{
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.Start,
					ExecuteAt:   testTime.Add(time.Hour),
				},
			},
			wantErr: ErrAlreadyExists,
		},
		{
			name: "register",
			initialEvents: []*scheduledEvent{
				{
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.Start,
						ExecuteAt:   testTime.Add(time.Hour),
					},
					handler: handler,
				},
			},
			actionName: "test",
			event: model.ScheduleEvent{
				ScheduleID:  "sid2",
				Summary:     "test",
				Description: "test schedule",
				EventType:   model.End,
				ExecuteAt:   testTime.Add(2 * time.Hour),
			},
			want: model.ScheduleEvents{
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.Start,
					ExecuteAt:   testTime.Add(time.Hour),
				},
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.End,
					ExecuteAt:   testTime.Add(2 * time.Hour),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemory(ctx)
			for _, e := range tt.initialEvents {
				s.eventMap.Store(e.key(), e)
			}

			err := s.Register(tt.actionName, tt.event, handler)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("\nwant: %+v\n got: %+v", tt.wantErr, err)
			}

			got := make(model.ScheduleEvents, 0)
			s.eventMap.Range(func(key, value interface{}) bool {
				e := value.(*scheduledEvent)
				got = append(got, e.event)
				return true
			})
			got.SortByExecuteAtAsc()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("\nwant: %#v\n got: %#v", tt.want, got)
			}
		})
	}
}

func TestScheduler_Unregister(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testTime := time.Now()
	handler := func(context.Context) error { return nil }
	tests := []struct {
		name          string
		initialEvents []*scheduledEvent
		actionName    model.ActionName
		events        []model.ScheduleEvent
		want          model.ScheduleEvents
	}{
		{
			name:          "not registered",
			initialEvents: []*scheduledEvent{},
			actionName:    "test",
			events: []model.ScheduleEvent{
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.Start,
					ExecuteAt:   testTime.Add(time.Hour),
				},
			},
			want: model.ScheduleEvents{},
		},
		{
			name: "unregister one event",
			initialEvents: []*scheduledEvent{
				{
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.Start,
						ExecuteAt:   testTime.Add(time.Hour),
					},
					handler: handler,
				}, {
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.End,
						ExecuteAt:   testTime.Add(2 * time.Hour),
					},
					handler: handler,
				},
			},
			actionName: "test",
			events: []model.ScheduleEvent{
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.End,
					ExecuteAt:   testTime.Add(2 * time.Hour),
				},
			},
			want: model.ScheduleEvents{
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.Start,
					ExecuteAt:   testTime.Add(time.Hour),
				},
			},
		},
		{
			name: "unregister some events",
			initialEvents: []*scheduledEvent{
				{
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.Start,
						ExecuteAt:   testTime.Add(time.Hour),
					},
					handler: handler,
				}, {
					actionName: "test",
					event: model.ScheduleEvent{
						ScheduleID:  "sid2",
						Summary:     "test",
						Description: "test schedule",
						EventType:   model.End,
						ExecuteAt:   testTime.Add(2 * time.Hour),
					},
					handler: handler,
				},
			},
			actionName: "test",
			events: []model.ScheduleEvent{
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.End,
					ExecuteAt:   testTime.Add(2 * time.Hour),
				},
				{
					ScheduleID:  "sid2",
					Summary:     "test",
					Description: "test schedule",
					EventType:   model.Start,
					ExecuteAt:   testTime.Add(time.Hour),
				},
			},
			want: model.ScheduleEvents{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemory(ctx)
			for _, e := range tt.initialEvents {
				s.eventMap.Store(e.key(), e)
			}

			err := s.Unregister(tt.actionName, tt.events...)
			if err != nil {
				t.Fatalf("err should be nil but got %+v", err)
			}

			got := make(model.ScheduleEvents, 0)
			s.eventMap.Range(func(key, value interface{}) bool {
				e := value.(*scheduledEvent)
				got = append(got, e.event)
				return true
			})
			got.SortByExecuteAtAsc()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("\nwant: %#v\n got: %#v", tt.want, got)
			}
		})
	}
}
