package model

import (
	"reflect"
	"testing"
	"time"
)

func TestSchedule_Events(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s    *Schedule
		t    time.Time
		want ScheduleEvents
	}{
		{
			s: &Schedule{
				ID:          "id",
				Summary:     "summary",
				Description: "desc",
				StartAt:     time.Unix(1, 0),
				EndAt:       time.Unix(20, 0),
			},
			t: time.Unix(0, 0),
			want: ScheduleEvents{
				{
					ScheduleID:  "id",
					Summary:     "summary",
					Description: "desc",
					EventType:   Start,
					ExecuteAt:   time.Unix(1, 0),
				},
				{
					ScheduleID:  "id",
					Summary:     "summary",
					Description: "desc",
					EventType:   End,
					ExecuteAt:   time.Unix(20, 0),
				},
			},
		},
		{
			s: &Schedule{
				ID:          "id",
				Summary:     "summary",
				Description: "desc",
				StartAt:     time.Unix(1, 0),
				EndAt:       time.Unix(20, 0),
			},
			t: time.Unix(1, 1),
			want: ScheduleEvents{
				{
					ScheduleID:  "id",
					Summary:     "summary",
					Description: "desc",
					EventType:   End,
					ExecuteAt:   time.Unix(20, 0),
				},
			},
		},
		{
			s: &Schedule{
				ID:          "id",
				Summary:     "summary",
				Description: "desc",
				StartAt:     time.Unix(1, 0),
				EndAt:       time.Unix(20, 0),
			},
			t:    time.Unix(20, 1),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.s.Events(tt.t)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("\nwant: %+v\n got: %+v", tt.want, got)
			}
		})
	}
}

func TestSchedules_Events(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ss   Schedules
		t    time.Time
		want ScheduleEvents
	}{
		{
			ss: Schedules{
				{
					ID:          "id1",
					Summary:     "summary",
					Description: "desc",
					StartAt:     time.Unix(1, 0),
					EndAt:       time.Unix(10, 0),
				},
				{
					ID:          "id2",
					Summary:     "summary",
					Description: "desc",
					StartAt:     time.Unix(10, 0),
					EndAt:       time.Unix(20, 0),
				},
				{
					ID:          "id3",
					Summary:     "summary",
					Description: "desc",
					StartAt:     time.Unix(2, 0),
					EndAt:       time.Unix(3, 0),
				},
			},
			t: time.Unix(10, 0),
			want: ScheduleEvents{
				{
					ScheduleID:  "id1",
					Summary:     "summary",
					Description: "desc",
					EventType:   End,
					ExecuteAt:   time.Unix(10, 0),
				},
				{
					ScheduleID:  "id2",
					Summary:     "summary",
					Description: "desc",
					EventType:   Start,
					ExecuteAt:   time.Unix(10, 0),
				},
				{
					ScheduleID:  "id2",
					Summary:     "summary",
					Description: "desc",
					EventType:   End,
					ExecuteAt:   time.Unix(20, 0),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.ss.Events(tt.t)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("\nwant: %+v\n got: %+v", tt.want, got)
			}
		})
	}
}

func TestScheduleEvent_ID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s         *ScheduleEvent
		delimiter string
		want      string
	}{
		{
			s:         &ScheduleEvent{},
			delimiter: "",
			want:      ":-62135596800",
		},
		{
			s: &ScheduleEvent{
				ScheduleID: "scheduleId1",
				ExecuteAt:  time.Unix(1, 0),
			},
			delimiter: "_",
			want:      "scheduleId1_1",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.s.ID(tt.delimiter)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("\nwant: %+v\n got: %+v", tt.want, got)
			}
		})
	}
}

func TestScheduleEvents_Sub(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ss   ScheduleEvents
		tt   ScheduleEvents
		want ScheduleEvents
	}{
		{
			ss:   nil,
			tt:   nil,
			want: ScheduleEvents{},
		},
		{
			ss:   ScheduleEvents{},
			tt:   ScheduleEvents{},
			want: ScheduleEvents{},
		},
		{
			ss: ScheduleEvents{
				{
					ScheduleID: "scheduleId1",
					ExecuteAt:  time.Unix(1, 0),
				},
			},
			tt: ScheduleEvents{},
			want: ScheduleEvents{
				{
					ScheduleID: "scheduleId1",
					ExecuteAt:  time.Unix(1, 0),
				},
			},
		},
		{
			ss: ScheduleEvents{
				{
					ScheduleID: "scheduleId1",
					ExecuteAt:  time.Unix(1, 0),
				},
				{
					ScheduleID: "scheduleId2",
					ExecuteAt:  time.Unix(1, 0),
				},
			},
			tt: ScheduleEvents{
				{
					ScheduleID: "scheduleId1",
					ExecuteAt:  time.Unix(1, 0),
				},
			},
			want: ScheduleEvents{
				{
					ScheduleID: "scheduleId2",
					ExecuteAt:  time.Unix(1, 0),
				},
			},
		},
		{
			ss: ScheduleEvents{
				{
					ScheduleID: "scheduleId1",
					ExecuteAt:  time.Unix(1, 0),
				},
			},
			tt: ScheduleEvents{
				{
					ScheduleID: "scheduleId1",
					ExecuteAt:  time.Unix(1, 0),
				},
				{
					ScheduleID: "scheduleId2",
					ExecuteAt:  time.Unix(1, 0),
				},
			},
			want: ScheduleEvents{},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tt.ss.Sub(tt.tt)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("\nwant: %+v\n got: %+v", tt.want, got)
			}
		})
	}
}
