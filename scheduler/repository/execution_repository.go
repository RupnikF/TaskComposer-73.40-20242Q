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

func (r *ExecutionRepository) GetExecutionById(ctx context.Context, id uint) *Execution {
	execution := Execution{}
	tx := r.db.WithContext(ctx).Preload("State").Preload("Steps").Preload("Steps.Inputs").Preload("State.Outputs").First(&execution, id)
	if tx.Error != nil {
		log.Printf("Failed to get execution: %v", tx.Error)
	}
	return &execution
}
func (r *ExecutionRepository) GetExecutionByUUID(ctx context.Context, uuid string) *Execution {
	execution := Execution{}
	tx := r.db.WithContext(ctx).Preload("State").Preload("Steps").Preload("State.Outputs").Where("execution_uuid = ?", uuid).First(&execution)
	if tx.Error != nil {
		log.Printf("Failed to get execution: %v", tx.Error)
	}
	return &execution
}

func (r *ExecutionRepository) GetExecutionsByJobID(ctx context.Context, jobID string) []*Execution {
	var executions []Execution
	tx := r.db.WithContext(ctx).Preload("State").Preload("Steps").Preload("State.Outputs").Where("job_id = ?", jobID).Find(&executions)
	if tx.Error != nil {
		log.Printf("Failed to get executions: %v", tx.Error)
	}
	output := make([]*Execution, len(executions))
	for i, execution := range executions {
		output[i] = &execution
	}
	return output
}

func (r *ExecutionRepository) UpdateState(ctx context.Context, state *State) {
	tx := r.db.WithContext(ctx).Save(state)
	if tx.Error != nil {
		log.Printf("Failed to update execution: %v", tx.Error)
	}
}

func (r *ExecutionRepository) GetStateByExecutionID(ctx context.Context, executionID uint) *State {
	state := State{}
	tx := r.db.WithContext(ctx).Where("execution_id = ?", executionID).Preload("Arguments").Preload("Outputs").First(&state)
	if tx.Error != nil {
		log.Printf("Failed to get state: %v", tx.Error)
	}
	return &state
}
func (r *ExecutionRepository) CancelExecution(ctx context.Context, execution *Execution) {
	execution.State.Status = CANCELLED
	r.UpdateState(ctx, execution.State)
}
func (r *ExecutionRepository) GetExecutionsByTags(ctx context.Context, tags []string) []*Execution {
	var executions []Execution
	tx := r.db.WithContext(ctx).Preload("State").Preload("Tags").Find(&executions)
	if tx.Error != nil {
		log.Printf("Failed to get executions: %v", tx.Error)
	}
	var outputExecutions []*Execution
	for _, execution := range executions {
		if execution.State.Status == PENDING || execution.State.Status == EXECUTING {
		tagsSearch:
			for _, tag := range tags {
				for _, executionTag := range execution.Tags {
					if executionTag.Tag == tag {
						outputExecutions = append(outputExecutions, &execution)
						break tagsSearch
					}
				}
			}
		}
	}
	return outputExecutions
}
