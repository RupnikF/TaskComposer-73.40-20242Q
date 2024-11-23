package com.taskcomposer.workflow_manager.controllers.dtos;

import lombok.Getter;

import java.util.List;
import java.util.Map;

@Getter
public class ExecutionRequestDTO {
    private String workflowName;
    private List<String> tags;
    private Map<String, String> parameters;
    private Map<String, String> args;
}
