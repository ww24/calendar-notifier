//go:generate mockgen -source=$GOFILE -destination=../../mock/mock_$GOPACKAGE/mock_$GOFILE -package=mock_repository

package repository

import (
	"context"
	"time"

	"github.com/ww24/calendar-notifier/domain/model"
)

// Calendar is the interface to control calendar service.
type Calendar interface {
	List(ctx context.Context, since, until time.Time) (model.Schedules, error)
}
