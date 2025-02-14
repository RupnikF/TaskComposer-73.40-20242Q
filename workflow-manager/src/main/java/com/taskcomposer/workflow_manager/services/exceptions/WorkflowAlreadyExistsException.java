package com.taskcomposer.workflow_manager.services.exceptions;

public class WorkflowAlreadyExistsException extends RuntimeException {
    public WorkflowAlreadyExistsException(String name){
        super(String.format("Workflow with the name %s already exists", name));
    }
}
