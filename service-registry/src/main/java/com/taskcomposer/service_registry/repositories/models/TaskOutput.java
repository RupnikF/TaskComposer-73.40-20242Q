package com.taskcomposer.service_registry.repositories.models;

import com.taskcomposer.service_registry.repositories.models.idclass.TaskId;
import jakarta.persistence.*;

@Entity
@Table(name = "task_outputs")
public class TaskOutput {
    @EmbeddedId
    private TaskId taskId;

    @Column
    private String key;

    @Column
    private String type;

    @ManyToOne(fetch = FetchType.LAZY)
    @MapsId("taskId")
    @JoinColumn(name = "name")
    private Task task;
}
