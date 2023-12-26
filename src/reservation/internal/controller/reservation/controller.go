package reservation

import (
	"context"
	"database/sql"
	"fmt"
	"go.example/saga/pkg/saga"
	"go.example/saga/pkg/store/postgres"
	"go.example/saga/reservation/pkg/model"
	"log"
)

const roomReservationSaga = "room-reservation"

const (
	roomBookingStep = "room-booking"
	paymentStep     = "payment"
)

// sagaSteps provides the service order steps to complete a reservation(SUCCESS/FAILED)
var sagaSteps = []saga.SagaStep{roomBookingStep, paymentStep}

// eventLogger defines the interface for ensuring the exact once event consuming as part of current tx
type eventLogger interface {
	IsConsumed(ctx context.Context, tx *sql.Tx, eventID string) bool
	Consume(ctx context.Context, tx *sql.Tx, eventID string) error
}

type repository interface {
	Add(ctx context.Context, tx *sql.Tx, r *model.Reservation) error
	UpdateStatus(ctx context.Context, tx *sql.Tx, ID string, status model.ReservationStatus) error
	QueryByID(ctx context.Context, tx *sql.Tx, ID string) (*model.ReservationView, error)
}

type ingester[T model.Payload] interface {
	Ingest(ctx context.Context) (chan model.Event[T], error)
}

// Controller defines a Reservation service controller.
type Controller struct {
	store           *postgres.Store
	eventLogger     eventLogger
	repository      repository
	sagaRepository  saga.Repository
	bookingIngester ingester[model.BookingEventPayload]
	paymentIngester ingester[model.PaymentEventPayload]
}

// New creates a reservation service controller.
func New(store *postgres.Store,
	eventLogger eventLogger,
	repository repository,
	sagaRepository saga.Repository,
	bookingIngester ingester[model.BookingEventPayload],
	paymentIngester ingester[model.PaymentEventPayload]) *Controller {
	return &Controller{store, eventLogger, repository, sagaRepository, bookingIngester, paymentIngester}
}

// PostReservation create the reservation in PENDING state and starts the saga process to complete the reservation
func (c *Controller) PostReservation(ctx context.Context, cmd model.ReservationCmd) (*model.Reservation, error) {
	// make reservation
	r := model.NewReservation(cmd.HotelID, cmd.RoomID, cmd.GuestID, cmd.PaymentDue, cmd.StartDate, cmd.EndDate, cmd.CreditCardNO)

	// alert kafka using type events
	// FIXME: implement properly transactional script pattern
	if _, err := c.store.Transact(ctx, func(tx *sql.Tx) (interface{}, error) {

		// persist reservation
		if err := c.repository.Add(ctx, tx, r); err != nil {
			return nil, err
		}

		// Start SAGA
		payload := r.ToJSONMap()
		currStep := saga.NextSagaStep(sagaSteps, "")
		sagaState := saga.NewSaga(roomReservationSaga, payload, currStep)

		if err := c.sagaRepository.Persist(ctx, tx, sagaState); err != nil {
			return nil, err
		}

		// publish outbox event to debezium
		outboxEvent := postgres.NewEvent(sagaState.ID.String(), string(currStep), postgres.RequestEventType, payload)
		if err := outboxEvent.Persist(ctx, tx); err != nil {
			return nil, err
		}

		log.Printf("Started Saga for reservationID %s sagaID %s", r.ID, sagaState.ID)

		return r, nil
	}); err != nil {
		return nil, err
	}

	return r, nil
}

func (c Controller) GetReservation(ctx context.Context, ID string) (interface{}, error) {
	r, err := c.store.Transact(ctx, func(tx *sql.Tx) (interface{}, error) {
		r, err := c.repository.QueryByID(ctx, tx, ID)
		return r, err
	})

	return r, err
}

// StartBookingIngestion starts the ingestion of room booking events.
func (c *Controller) StartBookingIngestion(ctx context.Context) error {
	// Ingest room booking events through the provided ingester.
	ch, err := c.bookingIngester.Ingest(ctx)
	if err != nil {
		return err
	}

	// Process each room booking event received from the ingester channel.
	for e := range ch {
		log.Printf("On RoomBookingEvent  key %s eventID %s payload %v", e.MsgID, e.EventID, e.Payload)
		if _, err := c.onStepEvent(ctx, e.MsgID, e.EventID, e.Payload.SagaStepStatus()); err != nil {
			log.Printf("Failed to process message key %s eventID %s: %v", e.MsgID, e.EventID, err)
		}
	}
	return nil
}

// StartPaymentIngestion starts the ingestion of room booking events.
func (c *Controller) StartPaymentIngestion(ctx context.Context) error {
	// Ingest room booking events through the provided ingester.
	ch, err := c.paymentIngester.Ingest(ctx)
	if err != nil {
		return err
	}

	// Process each room booking event received from the ingester channel.
	for e := range ch {
		log.Printf("On PaymentEvent key %s eventID %s payload %v", e.MsgID, e.EventID, e.Payload)
		if _, err := c.onStepEvent(ctx, e.MsgID, e.EventID, e.Payload.SagaStepStatus()); err != nil {
			log.Printf("Failed to process message key %s eventID %s: %v", e.MsgID, e.EventID, err)
		}
	}
	return nil
}

// onStepEvent is invoked by the ingester on incoming event
// in one transaction it ensures saga moving to next/prev status and update the reservation status
func (c *Controller) onStepEvent(ctx context.Context, msgID string, eventID string, sagaStepStatus saga.SagaStepStatus) (interface{}, error) {
	return c.store.Transact(ctx, func(tx *sql.Tx) (interface{}, error) {
		// 1. check if already processed event
		if c.eventLogger.IsConsumed(ctx, tx, eventID) {
			return nil, nil
		}

		// 2. find the saga
		state, err := c.sagaRepository.QueryByID(ctx, tx, msgID)
		if err != nil {
			//log.Printf("failed to fetch saga state %v", err)
			return nil, nil
		}

		// 3. update sagaStepStatus.update existing
		state.StepStatus[string(state.CurrentStep)] = sagaStepStatus

		// 4. Check current sagaStep Status and decide
		if sagaStepStatus == saga.SagaStepStatusSucceeded {
			if err := advance(ctx, tx, state); err != nil {
				return nil, err
			}
		} else if sagaStepStatus == saga.SagaStepStatusFailed || sagaStepStatus == saga.SagaStepStatusCompensated {
			if err := goBack(ctx, tx, state); err != nil {
				return nil, err
			}
		}

		// change sagaStatus
		state.NextSagaStatus()

		// FIXME: use optimistic locking => change version
		// FIXME: use sequence for versioning
		state.IncrementVersion()

		// Saga Update
		if err := c.sagaRepository.Update(ctx, tx, *state); err != nil {
			return nil, err
		}

		// update reservation status
		if err := c.updateReservationStatus(tx, *state, ctx); err != nil {
			return nil, err
		}

		// mark as consumed
		err2 := c.eventLogger.Consume(ctx, tx, eventID)
		return nil, err2
	})
}

// advance move saga step to next step based on sagasteps and current step
// generate a new request outbox event
func advance(ctx context.Context, tx *sql.Tx, state *saga.SagaState) error {
	nextState := saga.NextSagaStep(sagaSteps, state.CurrentStep)
	if nextState == "" {
		state.CurrentStep = ""
	} else {
		state.StepStatus[string(nextState)] = saga.SagaStepStatusStarted
		state.CurrentStep = nextState

		// Outbox insert
		outboxEvent := postgres.NewEvent(state.ID.String(), string(nextState), postgres.RequestEventType, state.Payload)
		if err := outboxEvent.Persist(ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

// goBack move saga step to prev step based on sagasteps and current step
// generate a compensating request outbox event
func goBack(ctx context.Context, tx *sql.Tx, state *saga.SagaState) error {
	prevState := saga.PrevSagaStep(sagaSteps, state.CurrentStep)
	if prevState == "" {
		state.CurrentStep = ""
	} else {
		state.StepStatus[string(prevState)] = saga.SagaStepStatusCompensating
		state.CurrentStep = prevState

		// Outbox insert
		payload := state.Payload
		payload["type"] = postgres.CancelEventType
		outboxEvent := postgres.NewEvent(state.ID.String(), string(prevState), postgres.CancelEventType, state.Payload)
		if err := outboxEvent.Persist(ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

// updateReservationStatus change the status of reservation baed on sagaState
func (c *Controller) updateReservationStatus(tx *sql.Tx, state saga.SagaState, ctx context.Context) error {
	sagaID := fmt.Sprintf("%v", state.Payload["reservationId"])
	if state.SagaStatus == saga.SagaStatusCompleted {
		if err := c.repository.UpdateStatus(ctx, tx, sagaID, model.ReservationStatusSucceed); err != nil {
			return err
		}
	} else if state.SagaStatus == saga.SagaStatusAborted {
		if err := c.repository.UpdateStatus(ctx, tx, sagaID, model.ReservationStatusFailed); err != nil {
			return err
		}
	}
	return nil
}
