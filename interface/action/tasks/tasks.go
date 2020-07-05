package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ww24/calendar-notifier/domain/model"
)

const (
	delimiter         = "_"
	contentTypeHeader = "Content-Type"
)

// Tasks implements repository.Action for tasks.
type Tasks struct {
	cli                 *cloudtasks.Client
	queuePath           string
	taskIDPrefix        string
	serviceAccountEmail string
	method              string
	url                 string
	header              http.Header
	payload             map[string]interface{}
}

// Client represents cloud tasks client.
type Client struct {
	cli       *cloudtasks.Client
	projectID string
}

// NewClient returns cloud tasks client.
func NewClient(ctx context.Context) (*Client, error) {
	cli, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	cred, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	c := &Client{
		cli:       cli,
		projectID: cred.ProjectID,
	}
	return c, nil
}

// New returns an action for cloud tasks.
func New(cli *Client, ac model.ActionConfig) *Tasks {
	return &Tasks{
		cli:                 cli.cli,
		queuePath:           fmt.Sprintf("projects/%s/locations/%s/queues/%s", cli.projectID, ac.Location, ac.Queue),
		taskIDPrefix:        ac.TaskIDPrefix + string(ac.Name),
		serviceAccountEmail: ac.ServiceAccountEmail,
		method:              ac.Method,
		url:                 ac.URL,
		header:              ac.Header,
		payload:             ac.Payload,
	}
}

// List lists schedule events from cloud tasks.
func (a *Tasks) List(ctx context.Context) (model.ScheduleEvents, error) {
	req := &taskspb.ListTasksRequest{
		Parent:       a.queuePath,
		ResponseView: taskspb.Task_BASIC,
		PageSize:     1000,
		PageToken:    "",
	}
	events := make([]model.ScheduleEvent, 0)
	it := a.cli.ListTasks(ctx, req)
	for {
		task, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(task.Name, a.generateTaskName("")) {
			continue
		}
		event := model.ScheduleEvent{
			ExecuteAt: time.Unix(task.ScheduleTime.Seconds, int64(task.ScheduleTime.Nanos)),
		}
		event.ScheduleID = event.ParseID(a.parseTaskName(task.Name), delimiter)
		events = append(events, event)
	}
	return events, nil
}

func (a *Tasks) generateTaskName(id string) string {
	// `TASK_ID` can contain only letters ([A-Za-z]), numbers ([0-9]),
	// hyphens (-), or underscores (_). The maximum length is 500
	taskID := a.taskIDPrefix + delimiter + id
	return a.queuePath + "/tasks/" + taskID
}

func (a *Tasks) parseTaskName(taskName string) string {
	return strings.TrimPrefix(taskName, a.generateTaskName(""))
}

// Register registeres schedule events to cloud tasks.
func (a *Tasks) Register(ctx context.Context, events ...model.ScheduleEvent) error {
	requests := make([]*taskspb.CreateTaskRequest, 0, len(events))
	for _, event := range events {
		var body []byte
		if a.payload != nil {
			d, err := json.Marshal(a.payload)
			if err != nil {
				return err
			}
			body = d
		}
		if a.header == nil {
			a.header = http.Header{}
		}
		if body != nil && a.header.Get(contentTypeHeader) == "" {
			a.header.Set(contentTypeHeader, "application/json")
		}
		headers := make(map[string]string, len(a.header))
		for k, h := range a.header {
			if len(h) > 0 {
				headers[k] = h[len(h)-1]
			}
		}

		req := &taskspb.CreateTaskRequest{
			Parent: a.queuePath,
			Task: &taskspb.Task{
				Name: a.generateTaskName(event.ID(delimiter)),
				MessageType: &taskspb.Task_HttpRequest{
					HttpRequest: &taskspb.HttpRequest{
						HttpMethod: taskspb.HttpMethod_POST,
						Url:        a.url,
						Headers:    headers,
						Body:       body,
						AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
							OidcToken: &taskspb.OidcToken{
								ServiceAccountEmail: a.serviceAccountEmail,
							},
						},
					},
				},
				ScheduleTime: &timestamp.Timestamp{
					Seconds: event.ExecuteAt.Unix(),
					Nanos:   int32(event.ExecuteAt.Nanosecond()),
				},
			},
		}
		log.Println("[tasks action] register, task_name:", req.Task.Name)
		requests = append(requests, req)
	}
	return a.registerTasks(ctx, requests...)
}

func (a *Tasks) registerTasks(ctx context.Context, requests ...*taskspb.CreateTaskRequest) error {
	for _, req := range requests {
		_, err := a.cli.CreateTask(ctx, req)
		if err != nil {
			switch status.Code(err) {
			case codes.AlreadyExists:
				log.Printf("[tasks action] already exists, task_name: %v\n", req.Task.Name)
				continue
			default:
				return err
			}
		}
	}
	return nil
}

// Unregister unregisters schedule events from cloud tasks.
func (a *Tasks) Unregister(ctx context.Context, events ...model.ScheduleEvent) error {
	requests := make([]*taskspb.DeleteTaskRequest, 0, len(events))
	for _, event := range events {
		req := &taskspb.DeleteTaskRequest{
			Name: a.generateTaskName(event.ID(delimiter)),
		}
		log.Println("[tasks action] unregister, task_name:", req.Name)
		requests = append(requests, req)
	}
	return a.unregisterTasks(ctx, requests...)
}

func (a *Tasks) unregisterTasks(ctx context.Context, requests ...*taskspb.DeleteTaskRequest) error {
	for _, req := range requests {
		if err := a.cli.DeleteTask(ctx, req); err != nil {
			return err
		}
	}
	return nil
}
