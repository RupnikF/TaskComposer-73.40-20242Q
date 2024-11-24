package repository

import "log"

func CreateExecution(execution *Execution) uint {
	tx := db.Create(execution)
	if tx.Error != nil {
		log.Printf("Failed to create execution: %v", tx.Error)
	}
	return execution.ID
}

func GetExecutionById(id uint) *Execution {
	execution := Execution{}
	tx := db.First(&execution, id)
	if tx.Error != nil {
		log.Printf("Failed to get execution: %v", tx.Error)
	}
	return &execution
}

func UpdateState(state *State) {
	tx := db.Save(state)
	if tx.Error != nil {
		log.Printf("Failed to update execution: %v", tx.Error)
	}
}

func GetStateByExecutionID(executionID uint) *State {
	state := State{}
	tx := db.Where("execution_id = ?", executionID).First(&state)
	if tx.Error != nil {
		log.Printf("Failed to get state: %v", tx.Error)
	}
	return &state
}
