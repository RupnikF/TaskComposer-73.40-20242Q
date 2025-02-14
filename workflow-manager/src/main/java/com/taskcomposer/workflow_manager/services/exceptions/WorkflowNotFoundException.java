package com.taskcomposer.workflow_manager.services.exceptions;

import com.taskcomposer.workflow_manager.repositories.model.Workflow;

public class WorkflowNotFoundException extends RuntimeException {
    public WorkflowNotFoundException(Workflow workflow) {
        super(String.format("Workflow of id %d not found", workflow.getId()));
    }
}
