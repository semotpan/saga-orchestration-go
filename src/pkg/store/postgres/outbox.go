package postgres

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"go.example/saga/pkg/jsonmap"
	"time"
)

// OutboxEvent represent outbox representation expected by debezium strimzi connect
type OutboxEvent struct {
	ID            uuid.UUID
	Timestamp     time.Time
	AggregateID   string
	AggregateType string
	Type          string
	Payload       jsonmap.JSONMap
}

// NewEvent factory method for building an event
func NewEvent(aggregateID, aggregateType, eventType string, payload jsonmap.JSONMap) OutboxEvent {
	return OutboxEvent{
		ID:            uuid.New(),
		Timestamp:     time.Now(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Type:          eventType,
		Payload:       payload,
	}
}

// EventType defines the payload event type
type EventType string

const (
	RequestEventType = "REQUEST"
	CancelEventType  = "CANCEL"
)

// FIXME refactor to a proper implementaiton
// Persist the outbox event within the provided Transaction and Context
func (oe *OutboxEvent) Persist(ctx context.Context, tx *sql.Tx) error {
	q := "INSERT INTO outboxevent(id, timestamp, aggregatetype, aggregateid, type, payload) VALUES ($1,$2,$3,$4,$5,$6)"
	if _, err := tx.ExecContext(ctx, q, oe.ID, oe.Timestamp, oe.AggregateType, oe.AggregateID, oe.Type, oe.Payload); err != nil {
		return err
	}

	return nil
}
