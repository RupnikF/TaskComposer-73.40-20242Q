package com.taskcomposer.workflow_manager.controllers.dtos;

import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import lombok.Getter;

import java.util.List;
import java.util.Map;

@Getter
public class ExecutionSubmissionDTO {
    private String workflowName;
    private Long workflowID;
    private List<String> tags;
    private Map<String, String> parameters;
    private Map<String, String> args;
    private List<StepDTO> steps;

    public ExecutionSubmissionDTO(Workflow workflow, List<String> tags, Map<String, String> parameters, Map<String, String> args) {
        this.workflowName = workflow.getWorkflowName();
        this.workflowID = workflow.getId();
        this.tags = tags;
        this.parameters = parameters;
        this.steps = workflow.getSteps().stream().map(StepDTO::fromStep).toList();
    }
}
