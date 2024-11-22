package com.taskcomposer.workflow_manager.repositories.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.*;

@Builder
@AllArgsConstructor
@NoArgsConstructor
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

    @JsonIgnore
    @ManyToOne
    @JoinColumn(name = "step_id", nullable = false)
    private Step step;

    @Column(name = "input_key", nullable = false)
    private String inputKey;

    @Column(name = "input_value", nullable = false)
    private String inputValue;

}

