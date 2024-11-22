package com.taskcomposer.workflow_manager.controllers.dtos;

import lombok.Getter;

import java.util.List;
import java.util.Map;

@Getter
public class WorkflowDTO {
    private String name;
    private List<String> tags;
    private List<Map<String, Map<String, String>>> args;

    private List<Map<String, StepDTO>> steps;

    @Getter
    public static class StepDTO {
        private String service;
        private String task;
        private Map<String, String> input;
    }
 }
