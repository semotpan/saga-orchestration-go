package hotel

import (
	"context"
	"database/sql"
	"go.example/saga/hotel/pkg/model"
	"go.example/saga/pkg/store/postgres"
	"log"
)

// roomBookIngester defines the interface for ingesting room booking events.
type roomBookIngester interface {
	Ingest(ctx context.Context) (chan model.RoomBookingEvent, error)
}

// eventLogger defines the interface for ensuring the exact once event consuming as part of current tx
type eventLogger interface {
	IsConsumed(ctx context.Context, tx *sql.Tx, eventID string) bool
	Consume(ctx context.Context, tx *sql.Tx, eventID string) error
}

// repository
type repository interface {
	IsRoomAvailable(ctx context.Context, tx *sql.Tx, roomID model.RoomID) (bool, error)
	BookRoom(ctx context.Context, tx *sql.Tx, roomID model.RoomID) error
	ReleaseRoom(ctx context.Context, tx *sql.Tx, roomID model.RoomID) error
}

// Controller is responsible for handling room booking events.
type Controller struct {
	ingester    roomBookIngester
	store       *postgres.Store
	eventLogger eventLogger
	repository  repository
}

// New creates a new instance of the hotel service controller.
func New(ingester roomBookIngester, store *postgres.Store, eventLogger eventLogger, repository repository) *Controller {
	return &Controller{ingester, store, eventLogger, repository}
}

// StartIngestion starts the ingestion of room booking events.
func (c *Controller) StartIngestion(ctx context.Context) error {
	// Ingest room booking events through the provided ingester.
	ch, err := c.ingester.Ingest(ctx)
	if err != nil {
		return err
	}

	// Process each room booking event received from the ingester channel.
	for e := range ch {
		log.Printf("on RoomBookingEvent: %d eventType: %s payload: %v", e.Payload.RoomID, e.Payload.Type, e)
		if _, err := c.onEvent(ctx, e); err != nil {
			log.Printf("Failed to process message key %s eventID %s: %v", e.MsgID, e.EventID, err)
		}
	}

	return nil
}

// onEvent processes a room booking event, updating the room availability and publishing an outbox event.
func (c *Controller) onEvent(ctx context.Context, e model.RoomBookingEvent) (interface{}, error) { // Perform the transaction using the Datasource.
	return c.store.Transact(ctx, func(tx *sql.Tx) (interface{}, error) {
		// ensure idempotence (at least once semantic)
		if c.eventLogger.IsConsumed(ctx, tx, e.EventID) {
			return nil, nil
		}

		// Process the room booking event and get its status.
		status, _ := c.handle(ctx, tx, e)
		outboxEvent := postgres.NewEvent(e.MsgID, "room-booking", "RoomUpdated", status.ToJSONMap())
		if err := outboxEvent.Persist(ctx, tx); err != nil {
			return nil, err
		}

		// consume the event
		if err := c.eventLogger.Consume(ctx, tx, e.EventID); err != nil {
			return nil, err
		}

		return status, nil
	})
}

// handle processes a room booking event and updates the room availability.
func (c *Controller) handle(ctx context.Context, tx *sql.Tx, e model.RoomBookingEvent) (model.BookingStatus, error) {
	available, err := c.repository.IsRoomAvailable(ctx, tx, e.Payload.RoomID)
	if err != nil {
		return model.BookingStatusRejected, err // in case of failures
	}

	var status model.BookingStatus
	if e.Payload.Type == postgres.RequestEventType {
		status = model.BookingStatusRejected
		if available {
			if err := c.repository.BookRoom(ctx, tx, e.Payload.RoomID); err == nil {
				status = model.BookingStatusBooked
			}
		}
	} else {
		// Release the room and publish a cancellation event.
		status = model.BookingStatusCancelled
		if err := c.repository.ReleaseRoom(ctx, tx, e.Payload.RoomID); err != nil {
			status = model.BookingStatusRejected
		}
	}
	return status, nil
}
