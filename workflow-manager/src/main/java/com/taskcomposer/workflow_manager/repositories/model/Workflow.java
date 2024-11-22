package com.taskcomposer.workflow_manager.repositories.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Setter;
import lombok.Getter;

import java.util.List;

@Builder
@Entity
@Getter
@Setter
@Table(name = "workflows")
@AllArgsConstructor
public class Workflow {
    public Workflow() {

    }

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Getter
    @Column(name = "workflow_name", unique = true, nullable = false)
    private String workflowName;

    @OneToMany(mappedBy = "workflow", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Argument> args;

    @OneToMany(mappedBy = "workflow", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Step> steps;

}

