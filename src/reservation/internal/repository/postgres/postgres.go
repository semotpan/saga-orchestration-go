package postgres

import (
	"context"
	"database/sql"
	"errors"
	"go.example/saga/reservation/internal/repository"
	"go.example/saga/reservation/pkg/model"
	"log"
)

// Repository defines a Postgres-based hotel repository.
type Repository struct {
}

// New repo constructor
func New() *Repository {
	return &Repository{}
}

// Add a new reservation into DB
func (rp Repository) Add(ctx context.Context, tx *sql.Tx, r *model.Reservation) error {
	// persist reservation
	qr := "INSERT INTO reservation(id, hotel_id, room_id, start_date, end_date, status, guest_id, payment_due, credit_card_no) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)"
	_, err := tx.ExecContext(ctx, qr, r.ID, r.HotelID, r.RoomID, r.StartDate, r.EndDate, r.Status, r.GuestID, r.PaymentDue, r.CreditCardNO)
	return err
}

// UpdateStatus update the status of the provided reservation ID
func (rp Repository) UpdateStatus(ctx context.Context, tx *sql.Tx, ID string, status model.ReservationStatus) error {
	// persist reservation
	_, err := tx.ExecContext(ctx, "UPDATE reservation SET status=$1 WHERE id=$2", status, ID)
	return err
}

func (rp Repository) QueryByID(ctx context.Context, tx *sql.Tx, ID string) (*model.ReservationView, error) {
	var r model.ReservationView
	row := tx.QueryRowContext(ctx, "SELECT id, status, hotel_id, guest_id, room_id FROM reservation WHERE id=$1", ID)
	err := row.Scan(&r.ID, &r.Status, &r.HotelID, &r.GuestID, &r.RoomID)
	if err != nil || errors.Is(err, sql.ErrNoRows) {
		log.Printf("failed to fetch saga state %v", err)
		return nil, repository.ErrNotFound
	}

	return &r, nil
}
