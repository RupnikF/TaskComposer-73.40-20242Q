package com.taskcomposer.service_registry.repositories.models.idclass;

import jakarta.persistence.Column;
import jakarta.persistence.Embeddable;

import java.io.Serializable;
import java.util.Objects;

@Embeddable
public class TaskId implements Serializable {
    @Column(name = "service_name")
    private String serviceName;

    @Column(name = "name")
    private String name;


    public TaskId() {

    }

    @Override
    public boolean equals(Object o) {
        if (o == null || getClass() != o.getClass()) return false;
        TaskId taskId = (TaskId) o;
        return Objects.equals(serviceName, taskId.serviceName) && Objects.equals(name, taskId.name);
    }

    @Override
    public int hashCode() {
        return Objects.hash(serviceName, name);
    }
}