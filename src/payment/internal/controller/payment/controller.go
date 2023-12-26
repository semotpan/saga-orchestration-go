package payment

import (
	"context"
	"database/sql"
	"go.example/saga/payment/pkg/model"
	"go.example/saga/pkg/store/postgres"
	"log"
)

// eventLogger defines the interface for ensuring the exact once event consuming as part of current tx
type eventLogger interface {
	IsConsumed(ctx context.Context, tx *sql.Tx, eventID string) bool
	Consume(ctx context.Context, tx *sql.Tx, eventID string) error
}

// repository
type repository interface {
	Add(ctx context.Context, tx *sql.Tx, p model.Payment) error
}

// roomBookIngester defines the interface for ingesting room booking events.
type ingester interface {
	Ingest(ctx context.Context) (chan model.PaymentEvent, error)
}

// Controller is responsible for handling room booking events.
type Controller struct {
	store       *postgres.Store
	repository  repository
	ingester    ingester
	eventLogger eventLogger
}

// New creates a new instance of the hotel service controller.
func New(store *postgres.Store, repository repository, ingester ingester, eventLogger eventLogger) *Controller {
	return &Controller{store, repository, ingester, eventLogger}
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
		log.Printf("on PaymentEvent: %d eventType: %s payload: %v", e.Payload.ID, e.Payload.Type, e)

		c.store.Transact(ctx, func(tx *sql.Tx) (interface{}, error) {
			// ensure idempotence (at least once semantic)
			if c.eventLogger.IsConsumed(ctx, tx, e.EventID) {
				return nil, nil
			}

			if err := c.repository.Add(ctx, tx, e.Payload); err != nil {
				return nil, err
			}

			// publish outbox event to debezium
			status := e.Payload.PaymentStatus()
			outboxEvent := postgres.NewEvent(e.MsgID, "payment", "PaymentUpdated", status.ToJSONMap())
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

	return nil
}
