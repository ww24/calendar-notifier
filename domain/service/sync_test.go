package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tenntenn/testtime"

	"github.com/ww24/calendar-notifier/mock/mock_repository"
)

func TestSynchronizer_Sync(t *testing.T) {
	t.Parallel()

	errCalendar := errors.New("calendar error")
	ts := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()
	tests := []struct {
		name     string
		injector func(
			*mock_repository.MockConfig,
			*mock_repository.MockCalendar,
			*mock_repository.MockActionConfigurator,
			*mock_repository.MockAction,
		)
		want error
	}{
		{
			name: "Sync with calendar error",
			injector: func(
				cnf *mock_repository.MockConfig,
				cal *mock_repository.MockCalendar,
				ac *mock_repository.MockActionConfigurator,
				action *mock_repository.MockAction,
			) {
				cal.EXPECT().List(ctx, ts, ts.Add(24*time.Hour)).Return(nil, errCalendar)
			},
			want: errCalendar,
		},
		// TODO: add more tests
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.True(t, testtime.SetTime(t, ts))

			ctrl := gomock.NewController(t)
			cnf := mock_repository.NewMockConfig(ctrl)
			cal := mock_repository.NewMockCalendar(ctrl)
			ac := mock_repository.NewMockActionConfigurator(ctrl)
			action := mock_repository.NewMockAction(ctrl)
			tt.injector(cnf, cal, ac, action)

			s := NewSynchronizer(cnf, cal, ac)
			err := s.Sync(ctx)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}
