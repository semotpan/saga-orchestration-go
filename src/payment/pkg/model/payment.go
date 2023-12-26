package model

import (
	"github.com/google/uuid"
	"go.example/saga/pkg/jsonmap"
	"go.example/saga/pkg/store/postgres"
	"strings"
	"time"
)

type Payment struct {
	ID           uuid.UUID          `json:"reservationId"`
	CreationTime time.Time          `json:"creationTime"`
	GuestID      int64              `json:"guestId"`
	PaymentDue   int64              `json:"paymentDue"`
	CreditCardNO string             `json:"creditCardNo"`
	Type         postgres.EventType `json:"type"`
}

// PaymentStatus simulate the payment status
func (p Payment) PaymentStatus() PaymentStatus {
	if p.Type == "" || p.CreditCardNO == "" {
		return PaymentStatusFailed
	}

	var status PaymentStatus
	if p.Type == postgres.RequestEventType {
		if strings.HasSuffix(p.CreditCardNO, "9999") { //FIXME: demo purpose
			status = PaymentStatusFailed
		} else {
			status = PaymentStatusRequested
		}
	} else {
		status = PaymentStatusCancelled
	}

	return status
}

// PaymentStatus the status of payment processing
type PaymentStatus string

// PaymentStatus type.
const (
	PaymentStatusRequested = "REQUESTED"
	PaymentStatusCancelled = "CANCELLED"
	PaymentStatusFailed    = "FAILED"
	PaymentStatusCompleted = "COMPLETED"
)

// ToJSONMap convert status to Json format
func (status PaymentStatus) ToJSONMap() jsonmap.JSONMap {
	return jsonmap.JSONMap{"status": string(status)}
}

// PaymentEvent incoming payment event request
type PaymentEvent struct {
	EventID   string
	MsgID     string
	Timestamp time.Time
	Payload   Payment
}
