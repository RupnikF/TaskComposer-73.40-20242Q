package repository

import (
	"gorm.io/gorm"
	"log"
)

type ExecutionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) *ExecutionRepository {
	return &ExecutionRepository{db}
}

func (r *ExecutionRepository) CreateExecution(execution *Execution) uint {
	tx := r.db.Create(execution)
	if tx.Error != nil {
		log.Printf("Failed to create execution: %v", tx.Error)
	}
	return execution.ID
}

func (r *ExecutionRepository) GetExecutionById(id uint) *Execution {
	execution := Execution{}
	tx := r.db.First(&execution, id)
	if tx.Error != nil {
		log.Printf("Failed to get execution: %v", tx.Error)
	}
	return &execution
}

func (r *ExecutionRepository) UpdateState(state *State) {
	tx := r.db.Save(state)
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
