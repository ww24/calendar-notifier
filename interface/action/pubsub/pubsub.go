package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2/google"

	"github.com/ww24/calendar-notifier/domain/model"
	"github.com/ww24/calendar-notifier/internal/scheduler"
)

const (
	timeout = 15 * time.Second
)

// PubSub implements repository.Action for pubsub.
type PubSub struct {
	cli     *Client
	name    model.ActionName
	topic   *pubsub.Topic
	payload map[string]interface{}
}

// Client represents cloud pubsub client.
type Client struct {
	cli       *pubsub.Client
	scheduler scheduler.Scheduler
}

// NewClient returns cloud pubsub client.
func NewClient(ctx context.Context) (*Client, error) {
	cred, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	cli, err := pubsub.NewClient(ctx, cred.ProjectID)
	if err != nil {
		return nil, err
	}
	c := &Client{
		cli:       cli,
		scheduler: scheduler.NewInMemory(ctx),
	}
	return c, nil
}

// New returns an action for cloud pubsub.
func New(cli *Client, ac model.ActionConfig) *PubSub {
	topic := cli.cli.Topic(ac.Topic)
	topic.PublishSettings.Timeout = timeout
	return &PubSub{
		cli:     cli,
		name:    ac.Name,
		topic:   topic,
		payload: ac.Payload,
	}
}

// List lists schedule events from cloud pubsub action scheduler.
func (a *PubSub) List(_ context.Context) (model.ScheduleEvents, error) {
	return a.cli.scheduler.List(a.name)
}

// Register registers schedule events to cloud pubsub action scheduler.
func (a *PubSub) Register(_ context.Context, events ...model.ScheduleEvent) error {
	for _, event := range events {
		var payload interface{}
		if a.payload != nil {
			payload = a.payload
		} else {
			payload = event
		}
		d, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		a.cli.scheduler.Register(a.name, event, func(ctx context.Context) error {
			pr := a.topic.Publish(ctx, &pubsub.Message{Data: d})
			id, err := pr.Get(ctx)
			if err != nil {
				return err
			}
			log.Println("Published, server_id:", id)
			return nil
		})
	}
	return nil
}

// Unregister unregisters schedule events from cloud pubsub action scheduler.
func (a *PubSub) Unregister(_ context.Context, events ...model.ScheduleEvent) error {
	return a.cli.scheduler.Unregister(a.name, events...)
}
