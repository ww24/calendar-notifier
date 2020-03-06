package action

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2/google"

	"github.com/ww24/calendar-notifier"
)

var (
	pubsubClient      *pubsub.Client
	pubsubClientMutex sync.Mutex
)

// PubSubAction implements action for pubsub.
type PubSubAction struct {
	topic string
}

// NewPubSubAction returns a new pubsub action.
func NewPubSubAction(topic string) (*PubSubAction, error) {
	pubsubClientMutex.Lock()
	defer pubsubClientMutex.Unlock()
	if pubsubClient == nil {
		ctx := context.Background()
		cred, err := google.FindDefaultCredentials(ctx)
		if err != nil {
			return nil, err
		}
		cli, err := pubsub.NewClient(ctx, cred.ProjectID)
		if err != nil {
			return nil, err
		}
		pubsubClient = cli
	}
	return &PubSubAction{
		topic: topic,
	}, nil
}

// Exec executes pubsub action.
func (a *PubSubAction) Exec(ctx context.Context, e *calendar.EventItem) error {
	topic := pubsubClient.Topic(a.topic)
	defer topic.Stop()
	topic.PublishSettings.Timeout = requestTimeout
	d, err := json.Marshal(e)
	if err != nil {
		return err
	}
	pr := topic.Publish(ctx, &pubsub.Message{Data: d})
	id, err := pr.Get(ctx)
	if err != nil {
		return err
	}
	log.Printf("Published, server_id: %v\n", id)
	return nil
}
