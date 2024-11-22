package com.taskcomposer.workflow_manager.repositories.model;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

@Setter
@Getter
@Entity
@Table(name = "args", uniqueConstraints = {
        @UniqueConstraint(columnNames = {"workflow_id", "arg_key"})
})
public class Argument {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne
    @JoinColumn(name = "workflow_id", nullable = false)
    private Workflow workflow;

    @Column(name = "arg_key", nullable = false)
    private String argKey;

    @Column(name = "default_value")
    private String defaultValue;

}

