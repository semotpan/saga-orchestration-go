package model

import (
	"github.com/google/uuid"
	"go.example/saga/pkg/jsonmap"
	"go.example/saga/pkg/saga"
	"go.example/saga/pkg/store/postgres"
	"time"
)

// Reservation defines an individual reservation created by a guest for hotel's room.
type Reservation struct {
	ID           uuid.UUID         `json:"reservationId"`
	HotelID      int64             `json:"hotelId"`
	RoomID       int64             `json:"roomId"`
	StartDate    string            `json:"startDate"`
	EndDate      string            `json:"endDate"`
	Status       ReservationStatus `json:"status"`
	GuestID      int64             `json:"guestId"`
	PaymentDue   int64             `json:"paymentDue"`
	CreditCardNO string            `json:"creditCardNo"`
}

// NewReservation creates a new reservation
func NewReservation(HotelID, RoomID, GuestID, PaymentDue int64, StartDate, EndDate, CreditCardNO string) *Reservation {
	return &Reservation{
		ID:           uuid.New(),
		HotelID:      HotelID,
		RoomID:       RoomID,
		StartDate:    StartDate,
		EndDate:      EndDate,
		GuestID:      GuestID,
		PaymentDue:   PaymentDue,
		CreditCardNO: CreditCardNO,
		Status:       ReservationStatusPending,
	}
}

// ReservationStatus defines the status of the reservation
type ReservationStatus string

// ReservationStatus type.
const (
	ReservationStatusPending   = "PENDING"
	ReservationStatusSucceed   = "SUCCEED"
	ReservationStatusFailed    = "FAILED"
	ReservationStatusCancelled = "CANCELLED"
	ReservationStatusRefund    = "REFUND"
)

type ReservationCmd struct {
	HotelID      int64  `json:"hotelId"`
	RoomID       int64  `json:"roomId"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	GuestID      int64  `json:"guestId"`
	PaymentDue   int64  `json:"paymentDue"`
	CreditCardNO string `json:"creditCardNo"`
}

type ReservationView struct {
	ID      uuid.UUID         `json:"reservationId"`
	HotelID int64             `json:"hotelId"`
	RoomID  int64             `json:"roomId"`
	GuestID int64             `json:"guestId"`
	Status  ReservationStatus `json:"status"`
}

func (r *Reservation) ToJSONMap() jsonmap.JSONMap {
	return map[string]interface{}{
		"reservationId": r.ID,
		"hotelId":       r.HotelID,
		"roomId":        r.RoomID,
		"startDate":     r.StartDate,
		"endDate":       r.EndDate,
		"status":        r.Status,
		"guestId":       r.GuestID,
		"paymentDue":    r.PaymentDue,
		"creditCardNo":  r.CreditCardNO,
		"type":          postgres.RequestEventType,
	}
}

type Event[T Payload] struct {
	EventID   string
	MsgID     string
	Timestamp time.Time
	Payload   T
}

type Payload interface {
	BookingEventPayload | PaymentEventPayload

	SagaStepStatus() saga.SagaStepStatus
}

// room-booking events
type (
	// BookingEventPayload JSON payload
	BookingEventPayload struct {
		Status BookingStatus `json:"status"`
	}

	// BookingStatus defines the booking status event response
	BookingStatus string
)

// BookingStatus type.
const (
	BookingStatusBooked    = "BOOKED"
	BookingStatusRejected  = "REJECTED"
	BookingStatusCancelled = "CANCELLED"
)

// SagaStepStatus defines the mapping - BookingStatus to SagaStepStatus
func (r BookingEventPayload) SagaStepStatus() saga.SagaStepStatus {
	switch r.Status {
	case BookingStatusBooked:
		return saga.SagaStepStatusSucceeded
	case BookingStatusRejected:
		return saga.SagaStepStatusFailed
	case BookingStatusCancelled:
		return saga.SagaStepStatusCompensated
	}

	return ""
}

// payment events
type (
	PaymentStatus string

	// PaymentEventPayload defines the payment event response status
	PaymentEventPayload struct {
		Status PaymentStatus `json:"status"`
	}
)

// PaymentStatus type.
const (
	PaymentStatusRequested = "REQUESTED"
	PaymentStatusCancelled = "CANCELLED"
	PaymentStatusFailed    = "FAILED"
	PaymentStatusCompleted = "COMPLETED"
)

// SagaStepStatus defines the mapping - PaymentStatus to SagaStepStatus
func (p PaymentEventPayload) SagaStepStatus() saga.SagaStepStatus {
	switch p.Status {
	case PaymentStatusRequested, PaymentStatusCompleted:
		return saga.SagaStepStatusSucceeded
	case PaymentStatusFailed:
		return saga.SagaStepStatusFailed
	case PaymentStatusCancelled:
		return saga.SagaStepStatusCompensated
	}

	return ""
}
