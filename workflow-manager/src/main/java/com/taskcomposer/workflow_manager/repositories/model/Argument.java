package com.taskcomposer.workflow_manager.repositories.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.*;

@Builder
@NoArgsConstructor
@AllArgsConstructor
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

    @JsonIgnore
    @ManyToOne
    @JoinColumn(name = "workflow_id", nullable = false)
    private Workflow workflow;

    @Column(name = "arg_key", nullable = false)
    private String argKey;

    @Column(name = "default_value")
    private String defaultValue;

}

