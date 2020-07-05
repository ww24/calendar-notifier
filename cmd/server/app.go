package main

import (
	"context"
	"net/http"

	"github.com/ww24/calendar-notifier/usecase"
)

type app struct {
	h    http.Handler
	sync usecase.Synchronizer
}

func newApp(h http.Handler, sync usecase.Synchronizer) *app {
	return &app{
		h:    h,
		sync: sync,
	}
}

func (a *app) server(port string) *http.Server {
	if port == "" {
		port = defaultPort
	}
	return &http.Server{
		Addr:    ":" + port,
		Handler: a.h,
	}
}

func (a *app) worker(ctx context.Context) error {
	return a.sync.Worker(ctx)
}
