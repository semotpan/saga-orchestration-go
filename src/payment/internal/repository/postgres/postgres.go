package postgres

import (
	"context"
	"database/sql"
	"go.example/saga/payment/pkg/model"
)

// Repository defines a Postgres-based hotel repository.
type Repository struct {
}

// New repo constructor
func New() *Repository {
	return &Repository{}
}

// Add a new payment to db
func (r Repository) Add(ctx context.Context, tx *sql.Tx, p model.Payment) error {
	// Insert payment
	if _, err := tx.ExecContext(ctx, "INSERT INTO payment(reservation_id, guest_id, payment_due, credit_card_no, type) VALUES ($1,$2,$3,$4,$5)",
		p.ID, p.GuestID, p.PaymentDue, p.CreditCardNO, p.Type); err != nil {
		return err
	}

	return nil
}
