package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/internal/scheduler"
)

const (
	timeout           = 15 * time.Second
	contentTypeHeader = "Content-Type"
)

// HTTP implements respository.Action for http request.
type HTTP struct {
	cli     *Client
	name    model.ActionName
	header  http.Header
	method  string
	url     string
	payload map[string]interface{}
}

// Client represents http client.
type Client struct {
	cli       *http.Client
	scheduler scheduler.Scheduler
}

// NewClient returns http client.
func NewClient(ctx context.Context) (*Client, error) {
	cli := &http.Client{
		Timeout: timeout,
	}
	c := &Client{
		cli:       cli,
		scheduler: scheduler.NewInMemory(ctx),
	}
	return c, nil
}

// New returns an action for http.
func New(cli *Client, ac model.ActionConfig) *HTTP {
	return &HTTP{
		cli:     cli,
		name:    ac.Name,
		header:  ac.Header,
		method:  ac.Method,
		url:     ac.URL,
		payload: ac.Payload,
	}
}

// List lists schedule events from http action scheduler.
func (a *HTTP) List(_ context.Context) (model.ScheduleEvents, error) {
	return a.cli.scheduler.List(a.name)
}

// Register registers schedule events to http action scheduler.
func (a *HTTP) Register(_ context.Context, events ...model.ScheduleEvent) error {
	for _, event := range events {
		req, err := a.newRequest(event)
		if err != nil {
			return err
		}
		a.cli.scheduler.Register(a.name, event, func(ctx context.Context) error {
			resp, err := a.cli.cli.Do(req.WithContext(ctx))
			if err != nil {
				return err
			}
			log.Println("[http action] sent, status:", resp.Status)
			return nil
		})
	}
	return nil
}

func (a *HTTP) newRequest(events model.ScheduleEvent) (*http.Request, error) {
	var body io.Reader
	if a.payload != nil {
		b := &bytes.Buffer{}
		e := json.NewEncoder(b)
		if err := e.Encode(a.payload); err != nil {
			return nil, err
		}
		body = b
	}
	req, err := http.NewRequest(a.method, a.url, body)
	if err != nil {
		return nil, err
	}
	if a.header != nil {
		req.Header = a.header
	}
	if body != nil && req.Header.Get(contentTypeHeader) == "" {
		req.Header.Set(contentTypeHeader, "application/json")
	}
	return req, nil
}

// Unregister unregisters schedule events from http action scheduler.
func (a *HTTP) Unregister(_ context.Context, events ...model.ScheduleEvent) error {
	return a.cli.scheduler.Unregister(a.name, events...)
}
