package com.taskcomposer.service_registry.repositories.models;

import com.taskcomposer.service_registry.repositories.models.idclass.TaskId;
import jakarta.persistence.*;
import java.util.List;

@Entity
@Table(name = "tasks", uniqueConstraints = {
        @UniqueConstraint(columnNames = { "service_name", "name" })
})
@IdClass(TaskId.class)
public class Task {
    @Id
    @Column(name = "service_name")
    private String serviceName;

    @Id
    @Column(name = "name")
    private String name;

    @ManyToOne
    @JoinColumn(name = "service_name", nullable = false)
    private Service service;

    @OneToMany(mappedBy = "task", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<TaskInput> inputs;

    @OneToMany(mappedBy = "task", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<TaskOutput> ouputs;
}
