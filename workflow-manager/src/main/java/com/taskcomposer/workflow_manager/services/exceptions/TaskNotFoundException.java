package com.taskcomposer.workflow_manager.services.exceptions;

public class TaskNotFoundException extends RuntimeException {
    public TaskNotFoundException(String serviceName, String taskName) {
        super(String.format("Task with the name %s could not be found in service %s", taskName, serviceName));
    }
}
