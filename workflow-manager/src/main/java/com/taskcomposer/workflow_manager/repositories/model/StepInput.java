package com.taskcomposer.workflow_manager.repositories.model;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

@Setter
@Getter
@Entity
@Table(name = "step_inputs", uniqueConstraints = {
        @UniqueConstraint(columnNames = {"step_id", "input_key"})
})
public class StepInput {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne
    @JoinColumn(name = "step_id", nullable = false)
    private Step step;

    @Column(name = "input_key", nullable = false)
    private String inputKey;

    @Column(name = "input_value", nullable = false)
    private String inputValue;

}

