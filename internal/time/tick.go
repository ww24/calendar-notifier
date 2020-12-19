package time

import (
	"context"
	"time"
)

// ImmediateTicker is not wait for first tick.
type ImmediateTicker struct {
	C      <-chan time.Time
	cancel context.CancelFunc
}

// NewImmediateTicker returns wrapped ticker.
func NewImmediateTicker(interval time.Duration) *ImmediateTicker {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan time.Time)
	ticker := &ImmediateTicker{
		C:      c,
		cancel: cancel,
	}
	go func() {
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
	return ticker
}

// Stop stops internal ticker.
func (t *ImmediateTicker) Stop() {
	t.cancel()
}
