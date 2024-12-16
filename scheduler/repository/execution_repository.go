package repository

import (
	"context"
	"log"

	"gorm.io/gorm"
)

type ExecutionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) *ExecutionRepository {
	return &ExecutionRepository{db}
}

func (r *ExecutionRepository) CreateExecution(ctx context.Context, execution *Execution) uint {
	tx := r.db.WithContext(ctx).Create(execution)
	if tx.Error != nil {
		log.Printf("Failed to create execution: %v", tx.Error)
	}
	return execution.ID
}

func (r *ExecutionRepository) GetExecutionById(id uint) *Execution {
	execution := Execution{}
	tx := r.db.Preload("State").Preload("Steps").Preload("Steps.Inputs").Preload("State.Outputs").First(&execution, id)
	if tx.Error != nil {
		log.Printf("Failed to get execution: %v", tx.Error)
	}
	return &execution
}
func (r *ExecutionRepository) GetExecutionByUUID(uuid string) *Execution {
	execution := Execution{}
	tx := r.db.Preload("State").Preload("Steps").Preload("State.Outputs").Where("execution_uuid = ?", uuid).First(&execution)
	if tx.Error != nil {
		log.Printf("Failed to get execution: %v", tx.Error)
	}
	return &execution
}

func (r *ExecutionRepository) UpdateState(ctx context.Context, state *State) {
	tx := r.db.WithContext(ctx).Save(state)
	if tx.Error != nil {
		log.Printf("Failed to update execution: %v", tx.Error)
	}
}

func (r *ExecutionRepository) GetStateByExecutionID(executionID uint) *State {
	state := State{}
	tx := r.db.Where("execution_id = ?", executionID).Preload("Arguments").Preload("Outputs").First(&state)
	if tx.Error != nil {
		log.Printf("Failed to get state: %v", tx.Error)
	}
	return &state
}
