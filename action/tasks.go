package action

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/ww24/calendar-notifier"
	"github.com/ww24/calendar-notifier/config"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	tasksClient      *cloudtasks.Client
	tasksClientMutex sync.Mutex
)

// TasksAction implements action for tasks.
type TasksAction struct {
	queuePath           string
	taskIDPrefix        string
	serviceAccountEmail string
	method              string
	url                 string
	header              http.Header
	payload             map[string]interface{}
}

// NewTasks returns an action for cloud tasks.
func NewTasks(locationID, queueID, taskIDPrefix, serviceAccountEmail, method, url string, header http.Header, payload map[string]interface{}, an config.ActionName) (Action, error) {
	action, err := newTasksAction(locationID, queueID, taskIDPrefix, serviceAccountEmail, method, url, header, payload, an)
	if err != nil {
		return nil, err
	}
	return wrapAction(action), nil
}

// newTasksAction returns a new tasks action.
func newTasksAction(locationID, queueID, taskIDPrefix, serviceAccountEmail, method, url string, header http.Header, payload map[string]interface{}, an config.ActionName) (*TasksAction, error) {
	tasksClientMutex.Lock()
	defer tasksClientMutex.Unlock()
	if tasksClient == nil {
		ctx := context.Background()
		cli, err := cloudtasks.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		tasksClient = cli
	}
	return &TasksAction{
		queuePath:           fmt.Sprintf("projects/%s/locations/%s/queues/%s", defaultProjectID, locationID, queueID),
		taskIDPrefix:        taskIDPrefix + string(an),
		serviceAccountEmail: serviceAccountEmail,
		method:              method,
		url:                 url,
		payload:             payload,
	}, nil
}

// ExecOnSchedule adds a scheduled task to cloud tasks.
func (a *TasksAction) ExecOnSchedule(ctx context.Context, e *calendar.EventItem) error {
	var scheduleTime time.Time
	switch e.EventType {
	case calendar.Start:
		scheduleTime = e.StartAt
	case calendar.End:
		scheduleTime = e.EndAt
	}

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

	// `TASK_ID` can contain only letters ([A-Za-z]), numbers ([0-9]),
	// hyphens (-), or underscores (_). The maximum length is 500
	taskID := fmt.Sprintf("%s_%d",
		a.taskIDPrefix,
		scheduleTime.Unix(),
	)
	taskName := fmt.Sprintf("%s/tasks/%s", a.queuePath, taskID)
	req := &taskspb.CreateTaskRequest{
		Parent: a.queuePath,
		Task: &taskspb.Task{
			Name: taskName,
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
				Seconds: scheduleTime.Unix(),
			},
		},
	}
	resp, err := tasksClient.CreateTask(ctx, req)
	if err != nil {
		switch status.Code(err) {
		case codes.AlreadyExists:
			log.Printf("Already exists, task_name: %v\n", taskName)
			return nil
		default:
			return err
		}
	}

	log.Printf("Scheduled, task_name: %v\n", resp.Name)
	return nil
}
