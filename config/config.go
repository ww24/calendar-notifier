package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

// RunningMode represents running mode.
type RunningMode string

const (
	// ModeNone is uncategorized running mode.
	ModeNone RunningMode = ""
	// ModeResident is resident running mode.
	ModeResident RunningMode = "resident"
	// ModeOnDemand is on-demand running mode.
	ModeOnDemand RunningMode = "ondemand"
)

// ActionName represents action name.
type ActionName string

// Config represents a config.yml.
type Config struct {
	Version    string                  `yaml:"version"`
	Mode       RunningMode             `yaml:"mode"`
	CalendarID string                  `yaml:"calendar_id"`
	Handler    map[string]EventHandler `yaml:"handler"`
	Action     map[ActionName]Action   `yaml:"action"`
}

// EventHandler is event handler which contains action names.
type EventHandler struct {
	Start []ActionName `yaml:"start"`
	End   []ActionName `yaml:"end"`
}

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

// Action is action definition.
type Action struct {
	Type              ActionType `yaml:"type"`
	HTTPRequestAction `yaml:",inline"`
	CloudPubSubAction `yaml:",inline"`
	CloudTasksAction  `yaml:",inline"`
	Payload           map[string]interface{} `yaml:"payload"`
}

// HTTPRequestAction is configuration of HTTP action.
type HTTPRequestAction struct {
	Method string      `yaml:"method"`
	Header http.Header `yaml:"header"`
	URL    string      `yaml:"url"`
}

// CloudPubSubAction is configuration of Cloud Pub/Sub action.
type CloudPubSubAction struct {
	Topic string `yaml:"topic"`
}

// CloudTasksAction is configuration of Cloud Tasks action.
type CloudTasksAction struct {
	Location            string `yaml:"location"`
	Queue               string `yaml:"queue"`
	TaskIDPrefix        string `yaml:"task_id_prefix"`
	ServiceAccountEmail string `yaml:"service_account_email"`
}

// Parse parses config file and returns config data.
func Parse(configPath string) (Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	cnf, err := ioutil.ReadAll(f)
	if err != nil {
		return Config{}, err
	}
	conf := Config{}
	if err := yaml.Unmarshal(cnf, &conf); err != nil {
		return Config{}, err
	}
	// set default running mode
	if conf.Mode == "" {
		conf.Mode = ModeResident
	}
	if err := conf.validate(); err != nil {
		return Config{}, fmt.Errorf("validation error: %w", err)
	}
	return conf, nil
}

func (c *Config) validate() error {
	if c.Version == "" {
		return errors.New("version is required")
	}
	switch c.Mode {
	case ModeResident:
	case ModeOnDemand:
	default:
		return fmt.Errorf("unsupported running mode: %s", c.Mode)
	}
	if c.CalendarID == "" {
		return errors.New("calendar_id is required")
	}
	if len(c.Handler) == 0 {
		return errors.New("handler should be defined one or more")
	}
	if len(c.Action) == 0 {
		return errors.New("action should be defined one or more")
	}
	for _, h := range c.Handler {
		for _, action := range append(h.Start, h.End...) {
			if action == "" {
				return errors.New("action name should not be empty")
			}
			// TODO: validate action name
			if _, ok := c.Action[action]; !ok {
				return fmt.Errorf("action (%s) is not defined", action)
			}
		}
	}
	for _, a := range c.Action {
		if c.Mode == ModeOnDemand && a.Type != ActionTasks {
			return fmt.Errorf("unsupported action type with ondemand running mode: %s", a.Type)
		}
		switch a.Type {
		case ActionHTTP:
		case ActionPubSub:
		case ActionTasks:
		case ActionNone:
		default:
			return fmt.Errorf("unsupported action type: %s", a.Type)
		}
	}
	return nil
}
