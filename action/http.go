package action

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ww24/calendar-notifier"
)

const (
	contentTypeHeader = "Content-Type"
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
	payload map[string]interface{}
}

// NewHTTP returns an action for http.
func NewHTTP(header http.Header, method, url string, payload map[string]interface{}) Action {
	action := newHTTPAction(header, method, url, payload)
	return wrapAction(action)
}

// newHTTPAction returns a new http action.
func newHTTPAction(header http.Header, method, url string, payload map[string]interface{}) *HTTPAction {
	return &HTTPAction{
		header:  header,
		method:  method,
		url:     url,
		payload: payload,
	}
}

// ExecImmediately executes pubsub action.
func (a *HTTPAction) ExecImmediately(ctx context.Context, e *calendar.EventItem) error {
	var body io.Reader
	if a.payload != nil {
		b := &bytes.Buffer{}
		e := json.NewEncoder(b)
		if err := e.Encode(a.payload); err != nil {
			return err
		}
		body = b
	}
	req, err := http.NewRequestWithContext(ctx, a.method, a.url, body)
	if err != nil {
		return err
	}
	if a.header != nil {
		req.Header = a.header
	}
	if body != nil && req.Header.Get(contentTypeHeader) == "" {
		req.Header.Set(contentTypeHeader, "application/json")
	}
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	log.Printf("Sent, status: %v\n", resp.Status)
	return nil
}
