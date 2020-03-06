package action

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"

	"github.com/ww24/calendar-notifier"
)

var (
	httpClient = &http.Client{
		Timeout: requestTimeout,
	}
)

// HTTPAction implements action for HTTP.
type HTTPAction struct {
	header  http.Header
	method  string
	url     string
	payload string
}

// NewHTTPAction returns a new http action.
func NewHTTPAction(header http.Header, method, url, payload string) *HTTPAction {
	return &HTTPAction{
		header:  header,
		method:  method,
		url:     url,
		payload: payload,
	}
}

// Exec executes pubsub action.
func (a *HTTPAction) Exec(ctx context.Context, e *calendar.EventItem) error {
	var body io.Reader
	if a.payload != "" {
		body = bytes.NewBufferString(a.payload)
	}
	req, err := http.NewRequest(a.method, a.url, body)
	if err != nil {
		return err
	}
	req.Header = a.header
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	log.Printf("Sent, status: %v\n", resp.Status)
	return nil
}
