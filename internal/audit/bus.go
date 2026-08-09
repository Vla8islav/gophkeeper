// Package audit
package audit

import (
	"context"
)

// Event describes an audited metrics operation
// generate:reset
type Event struct {
	Time       UnixTime `json:"ts"`
	Metrics    []string `json:"metrics"`
	RemoteAddr string   `json:"ip_address"`
	Operation  string   `json:"operation"`
}

// Sink writes audit events to a destination
type Sink interface {
	Write(ctx context.Context, event Event) error
}

// Publisher sends audit events to one or more sinks
type Publisher struct {
	sinks []Sink
}

// NewPublisher creates a Publisher that writes events to sinks
func NewPublisher(sinks ...Sink) *Publisher {
	return &Publisher{sinks: sinks}
}

// Publish writes event to each configured sink
func (p *Publisher) Publish(ctx context.Context, event Event) error {
	for _, sink := range p.sinks {
		if err := sink.Write(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
