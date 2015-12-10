package event

import (
	"encoding/json"
	"time"
)

type Event struct {
	Fqdn            string            `json:"fqdn,omitempty"`
	Time            int64             `json:"time,omitempty"`
	Service         string            `json:"service,omitempty"`
	ServiceTSDB     string            `json:"tsdb,omitempty"`
	ServiceGraphite string            `json:"graphite,omitempty"`
	ServiceZabbix   string            `json:"zabbix,omitempty"`
	Value           float64           `json:"value,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
}

type Events []*Event

func Parse(body []byte) ([]*Event, error) {
	events, err := parseEvents(body)
	if err == nil {
		return events, nil
	}
	event, err := parseEvent(body)
	if err == nil {
		events := make([]*Event, 0)
		events = append(events, event)
		return events, nil
	}
	return nil, err
}

func parseEvent(body []byte) (*Event, error) {
	event := New()
	if err := json.Unmarshal(body, event); err == nil {
		return event, nil
	} else {
		return nil, err
	}
}

func parseEvents(body []byte) ([]*Event, error) {
	events := make([]*Event, 0)
	if err := json.Unmarshal(body, events); err == nil {
		return events, nil
	} else {
		return nil, err
	}
}

func New() *Event {
	return &Event{Tags: make(map[string]string, 0)}
}

func (e *Event) Create() *Event {
	return New()
}

func (e *Event) GetFqdn() string {
	return e.Fqdn
}

func (e *Event) GetService() string {
	return e.Service
}

func (e *Event) GetServiceTSDB() string {
	return e.ServiceTSDB
}

func (e *Event) GetServiceGraphite() string {
	return e.ServiceGraphite
}

func (e *Event) GetServiceZabbix() string {
	return e.ServiceZabbix
}

func (e *Event) GetTime() int64 {
	if e.Time == 0 {
		return time.Now().Unix()
	}
	return e.Time
}

func (e *Event) GetValue() float64 {
	return e.Value
}

func (e *Event) GetTags() map[string]string {
	return e.Tags
}
