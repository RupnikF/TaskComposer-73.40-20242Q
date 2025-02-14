package com.taskcomposer.workflow_manager.repositories.model;

import lombok.Getter;

import java.util.Set;

@Getter
public class Service {
    private String name;
    private Set<String> tasks;

    public Service(String name, Set<String> tasks) {
        this.name = name;
        this.tasks = tasks;
    }

    public boolean isTaskInService(String taskName) {
        return this.getTasks().contains(taskName);
    }
}
