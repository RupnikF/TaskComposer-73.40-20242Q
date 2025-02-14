package com.taskcomposer.workflow_manager.repositories.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Setter;
import lombok.Getter;

import java.util.ArrayList;
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

    @OneToMany(mappedBy = "workflow", cascade = CascadeType.ALL, orphanRemoval = true, fetch = FetchType.EAGER)
    private List<Argument> args;

    @OneToMany(mappedBy = "workflow", cascade = CascadeType.ALL, orphanRemoval = true, fetch = FetchType.EAGER)
    @OrderColumn(name = "step_order")
    private List<Step> steps;

    public void setArgs(List<Argument> newArgs) {
        if (this.args == null) {
            this.args = new ArrayList<>();
        } else {
            this.args.clear();
        }
        if (newArgs != null) {
            for (Argument arg : newArgs) {
                arg.setWorkflow(this);  // Maintain consistency
                this.args.add(arg);
            }
        }
    }

    public void setSteps(List<Step> newSteps) {
        if (this.steps == null) {
            this.steps = new ArrayList<>();
        } else {
            this.steps.clear();
        }
        if (newSteps != null) {
            for (Step step : newSteps) {
                step.setWorkflow(this);  // Maintain consistency
                this.steps.add(step);
            }
        }
    }

}

