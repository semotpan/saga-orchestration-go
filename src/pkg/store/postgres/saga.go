package postgres

import (
	"context"
	"database/sql"
	"errors"
	"go.example/saga/pkg/saga"
	"log"
)

type SagaRepository struct {
}

func NewSagaRepository() *SagaRepository {
	return &SagaRepository{}
}

func (sr SagaRepository) Persist(ctx context.Context, tx *sql.Tx, ss saga.SagaState) error {
	qss := "INSERT INTO sagastate(id, version, type, payload, current_step, step_status, saga_status) VALUES ($1,$2,$3,$4,$5,$6,$7)"
	_, err := tx.ExecContext(ctx, qss, ss.ID, ss.Version, ss.Type, ss.Payload, ss.CurrentStep, ss.StepStatus, ss.SagaStatus)
	return err
}

func (sr SagaRepository) Update(ctx context.Context, tx *sql.Tx, ss saga.SagaState) error {
	q := "UPDATE sagastate SET version=$1, payload=$2, current_step=$3, step_status=$4, saga_status=$5 WHERE id=$6"
	_, err := tx.ExecContext(ctx, q, ss.Version, ss.Payload, ss.CurrentStep, ss.StepStatus, ss.SagaStatus, ss.ID)
	return err
}

func (sr SagaRepository) QueryByID(ctx context.Context, tx *sql.Tx, ID string) (*saga.SagaState, error) {
	var ss saga.SagaState
	row := tx.QueryRowContext(ctx, "SELECT id, version, type, payload, current_step, step_status, saga_status FROM sagastate WHERE id=$1", ID)
	err := row.Scan(&ss.ID, &ss.Version, &ss.Type, &ss.Payload, &ss.CurrentStep, &ss.StepStatus, &ss.SagaStatus)
	if err != nil || errors.Is(err, sql.ErrNoRows) {
		log.Printf("failed to fetch saga state %v", err)
		return nil, err
	}

	return &ss, nil
}
