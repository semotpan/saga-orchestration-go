package postgres

import (
	"context"
	"database/sql"
	"go.example/saga/hotel/pkg/model"
)

// Repository defines a Postgres-based hotel repository.
type Repository struct {
}

// New repo constructor
func New() *Repository {
	return &Repository{}
}

// IsRoomAvailable check if provided RoomID is available inside the provided TX
func (r Repository) IsRoomAvailable(ctx context.Context, tx *sql.Tx, roomID model.RoomID) (bool, error) {
	var available bool
	row := tx.QueryRowContext(ctx, "SELECT available FROM room WHERE id=$1", roomID)
	if err := row.Scan(&available); err != nil {
		return false, err
	}

	return available, nil
}

// BookRoom book the provided roomID inside the provided TX
func (r Repository) BookRoom(ctx context.Context, tx *sql.Tx, roomID model.RoomID) error {
	_, err := tx.ExecContext(ctx, "UPDATE room SET available=false WHERE id=$1", roomID)
	return err
}

// ReleaseRoom release the provided roomID inside the provided TX
func (r Repository) ReleaseRoom(ctx context.Context, tx *sql.Tx, roomID model.RoomID) error {
	_, err := tx.ExecContext(ctx, "UPDATE room SET available=true WHERE id=$1", roomID)
	return err
}
