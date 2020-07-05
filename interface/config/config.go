package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ww24/calendar-notifier/domain/model"
)

const (
	defaultInterval = 1 * time.Minute
)

// Config represents a config.yml.
type Config struct {
	Version    string                      `yaml:"version"`
	Mode       model.RunningMode           `yaml:"mode"`
	Interval   time.Duration               `yaml:"interval"`
	CalendarID string                      `yaml:"calendar_id"`
	Handler    map[string]EventHandler     `yaml:"handler"`
	Action     map[model.ActionName]Action `yaml:"action"`
}

// EventHandler is event handler which contains action names.
type EventHandler struct {
	Start []model.ActionName `yaml:"start"`
	End   []model.ActionName `yaml:"end"`
}

// Action is action definition.
type Action struct {
	Type              model.ActionType `yaml:"type"`
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
func Parse(configPath string) (*Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cnf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	if err := yaml.Unmarshal(cnf, conf); err != nil {
		return nil, err
	}
	// set default running mode
	if conf.Mode == "" {
		conf.Mode = model.ModeResident
	}
	if err := conf.validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	return conf, nil
}

func (c *Config) validate() error {
	if c.Version == "" {
		return errors.New("version is required")
	}
	switch c.Mode {
	case model.ModeResident:
	case model.ModeOnDemand:
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
		if c.Mode == model.ModeOnDemand && a.Type != model.ActionTasks {
			return fmt.Errorf("unsupported action type with ondemand running mode: %s", a.Type)
		}
		switch a.Type {
		case model.ActionHTTP:
		case model.ActionPubSub:
		case model.ActionTasks:
		default:
			return fmt.Errorf("unsupported action type: %s", a.Type)
		}
	}
	return nil
}

// ActionNames returns action names from event schedule.
func (c *Config) ActionNames(event model.ScheduleEvent) ([]model.ActionName, bool) {
	eh, ok := c.Handler[strings.TrimSpace(event.Summary)]
	if !ok {
		return nil, false
	}
	switch event.EventType {
	case model.Start:
		return eh.Start, true
	case model.End:
		return eh.End, true
	}
	return nil, false
}

// ActionConfigMap returns action config map.
func (c *Config) ActionConfigMap() map[model.ActionName]model.ActionConfig {
	acm := make(map[model.ActionName]model.ActionConfig, len(c.Action))
	for an, action := range c.Action {
		acm[an] = c.toActionConfig(action, an)
	}
	return acm
}

func (c *Config) toActionConfig(a Action, an model.ActionName) model.ActionConfig {
	ac := model.ActionConfig{
		Name:    an,
		Type:    a.Type,
		Payload: a.Payload,
	}
	switch ac.Type {
	case model.ActionHTTP:
		ac.HTTPRequestAction = model.HTTPRequestAction(a.HTTPRequestAction)
	case model.ActionPubSub:
		ac.CloudPubSubAction = model.CloudPubSubAction(a.CloudPubSubAction)
	case model.ActionTasks:
		ac.CloudTasksAction = model.CloudTasksAction(a.CloudTasksAction)
		ac.HTTPRequestAction = model.HTTPRequestAction(a.HTTPRequestAction)
	}
	return ac
}

// RunningMode returns running mode.
func (c *Config) RunningMode() model.RunningMode {
	return c.Mode
}

// SyncInterval retruns sync interval for resident mode.
func (c *Config) SyncInterval() time.Duration {
	if c.Interval == 0 {
		return defaultInterval
	}
	return c.Interval
}

// Calendar returns google calendar id.
func (c *Config) Calendar() string {
	return c.CalendarID
}
