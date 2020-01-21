package action

import (
	"context"
	"time"

	"github.com/ww24/calendar-worker"
)

const (
	requestTimeout = 15 * time.Second
)

// Action is interface for action behavior.
type Action interface {
	Exec(context.Context, *calendar.EventItem) error
}
