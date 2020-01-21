package calendar

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestImmediateTick(t *testing.T) {
	t.Parallel()
	accuracy := 100 * time.Millisecond
	patterns := []struct {
		interval time.Duration
		want     []time.Duration
	}{
		{
			interval: 3 * accuracy,
			want:     []time.Duration{0, 3 * accuracy, 6 * accuracy},
		},
		{
			interval: 5 * accuracy,
			want:     []time.Duration{0, 5 * accuracy, 10 * accuracy},
		},
	}
	for _, p := range patterns {
		p := p
		t.Run("", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			got := make([]time.Duration, 0, len(p.want))
			start := time.Now()
			for t := range ImmediateTick(ctx, p.interval) {
				got = append(got, t.Sub(start).Truncate(accuracy))
				if len(got) >= len(p.want) {
					cancel()
				}
			}
			if !reflect.DeepEqual(p.want, got) {
				t.Fatalf("got: %+v, but want: %+v", got, p.want)
			}
		})
	}
}
