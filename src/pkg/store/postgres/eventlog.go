package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// EventLog keeps the consumed events from kafka ingester
// kafka follows the at least once semantic, message log ensure consumed events are tracked
type EventLog struct {
	EventID  string
	IssuedOn time.Time
}

// EventLogs defines the data access type for kafka consumed events
type EventLogs struct {
}

// NewEventLogs constructor
func NewEventLogs() *EventLogs {
	return &EventLogs{}
}

// IsConsumed check if provided eventId is already exit into the DB in the current TX
func (el EventLogs) IsConsumed(ctx context.Context, tx *sql.Tx, eventID string) bool {
	var consumed bool
	row := tx.QueryRowContext(ctx, "SELECT count(event_id)=1 FROM eventlog WHERE event_id=$1", eventID)
	if err := row.Scan(&consumed); err == nil && consumed {
		log.Printf("Event %s already consumed", eventID)
		return true
	}
	return false
}

// Consume insert the provided eventId into consumed message array in the current TX
func (el EventLogs) Consume(ctx context.Context, tx *sql.Tx, eventID string) error {
	// consume the event
	_, err := tx.ExecContext(ctx, "INSERT INTO eventlog(event_id, issued_on) VALUES ($1,$2)", eventID, time.Now())
	return err
}
