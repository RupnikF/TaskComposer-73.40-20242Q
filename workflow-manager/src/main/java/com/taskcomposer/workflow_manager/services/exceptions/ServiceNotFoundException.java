package com.taskcomposer.workflow_manager.services.exceptions;

public class ServiceNotFoundException extends RuntimeException {
    public ServiceNotFoundException(String name) {
        super(String.format("Service with the name %s could not be found", name));
    }
}
