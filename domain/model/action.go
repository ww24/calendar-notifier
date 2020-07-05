package model

import "net/http"

// ActionType represents action type.
type ActionType string

const (
	// ActionNone is uncategorized action type.
	ActionNone ActionType = ""
	// ActionHTTP is action type for HTTP action.
	ActionHTTP ActionType = "http"
	// ActionPubSub is action type for Cloud Pub/Sub action.
	ActionPubSub ActionType = "pubsub"
	// ActionTasks is action type for Cloud Tasks action.
	ActionTasks ActionType = "tasks"
)

// ActionName represents action name.
type ActionName string

// ActionConfig is action configuration.
type ActionConfig struct {
	Name ActionName
	Type ActionType
	HTTPRequestAction
	CloudPubSubAction
	CloudTasksAction
	Payload map[string]interface{}
}

// HTTPRequestAction is parameter of HTTP action.
type HTTPRequestAction struct {
	Method string
	Header http.Header
	URL    string
}

// CloudPubSubAction is parameter of Cloud Pub/Sub action.
type CloudPubSubAction struct {
	Topic string
}

// CloudTasksAction is parameter of Cloud Tasks action.
type CloudTasksAction struct {
	Location            string
	Queue               string
	TaskIDPrefix        string
	ServiceAccountEmail string
}
