package saga

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.example/saga/pkg/jsonmap"
)

type SagaState struct {
	ID          uuid.UUID
	Version     int8
	Type        string
	Payload     jsonmap.JSONMap
	CurrentStep SagaStep
	StepStatus  jsonmap.JSONMap
	SagaStatus  SagaStatus
}

// Repository
type Repository interface {
	Persist(ctx context.Context, tx *sql.Tx, ss SagaState) error
	Update(ctx context.Context, tx *sql.Tx, ss SagaState) error
	QueryByID(ctx context.Context, tx *sql.Tx, ID string) (*SagaState, error)
}

func NewSaga(sagaType string, payload jsonmap.JSONMap, currentStep SagaStep) SagaState {
	return SagaState{
		ID:          uuid.New(),
		Version:     1,
		Type:        sagaType,
		Payload:     payload,
		CurrentStep: currentStep,
		StepStatus:  jsonmap.JSONMap{string(currentStep): SagaStepStatusStarted},
		SagaStatus:  SagaStatusStarted,
	}
}

// NextSagaStatus evaluate current SagaStepStatuses and set SagaStatus
func (s *SagaState) NextSagaStatus() {
	ss := map[string]bool{}
	for _, v := range s.StepStatus {
		ss[fmt.Sprintf("%v", v)] = true
	}

	if ss[SagaStepStatusSucceeded] && len(ss) == 1 {
		s.SagaStatus = SagaStatusCompleted
	} else if (ss[SagaStepStatusStarted] && len(ss) == 1) || (ss[SagaStepStatusSucceeded] && ss[SagaStepStatusStarted] && len(ss) == 2) {
		s.SagaStatus = SagaStatusStarted
	} else if !ss[SagaStepStatusCompensating] {
		s.SagaStatus = SagaStatusAborted
	} else {
		s.SagaStatus = SagaStatusAborting
	}
}

// IncrementVersion
func (s *SagaState) IncrementVersion() {
	s.Version++
}

// SagaStatus represents the saga status based on steps status
type SagaStatus string

// SagaStatus type
const (
	SagaStatusStarted   = "STARTED"
	SagaStatusAborting  = "ABORTING"
	SagaStatusAborted   = "ABORTED"
	SagaStatusCompleted = "COMPLETED"
)

// SagaStepStatus represent current saga step status
type SagaStepStatus string

// SagaStepStatus type
const (
	SagaStepStatusStarted      = "STARTED"
	SagaStepStatusFailed       = "FAILED"
	SagaStepStatusSucceeded    = "SUCCEEDED"
	SagaStepStatusCompensating = "COMPENSATING"
	SagaStepStatusCompensated  = "COMPENSATED"
)

// SagaStep define saga service step in order to follow
type SagaStep string

// NextSagaStep find saga next step from provided steps and current saga step
func NextSagaStep(steps []SagaStep, currentStep SagaStep) SagaStep {
	if currentStep == "" {
		return steps[0]
	}

	curr := -1
	for i := 0; i < len(steps); i++ {
		if steps[i] == currentStep {
			curr = i
			break
		}
	}

	if curr == -1 || curr+1 == len(steps) {
		return ""
	}

	return steps[curr+1]
}

// PrevSagaStep find saga previous step from provided steps and current saga step
func PrevSagaStep(steps []SagaStep, currentStep SagaStep) SagaStep {
	curr := -1
	for i := 0; i < len(steps); i++ {
		if steps[i] == currentStep {
			curr = i
			break
		}
	}

	if curr == -1 || curr-1 == -1 {
		return ""
	}

	return steps[curr-1]
}
