package time

import (
	"reflect"
	"testing"
	"time"
)

func TestImmediateTicker(t *testing.T) {
	accuracy := 100 * time.Millisecond
	tests := []struct {
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
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := make([]time.Duration, 0, len(tt.want))
			start := time.Now()
			ticker := NewImmediateTicker(tt.interval)
			for t := range ticker.C {
				got = append(got, t.Sub(start).Truncate(accuracy))
				if len(got) >= len(tt.want) {
					break
				}
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("got: %+v, but want: %+v", got, tt.want)
			}
		})
	}
}
