package com.taskcomposer.workflow_manager.services.exceptions;

public class WorkflowNotFoundException extends RuntimeException {
    public WorkflowNotFoundException(String name){
        super(String.format("Workflow with the name %s could not be found", name));
    }
}
