package com.taskcomposer.workflow_manager.controllers.dtos;

import com.taskcomposer.workflow_manager.repositories.model.Argument;
import com.taskcomposer.workflow_manager.repositories.model.Step;
import com.taskcomposer.workflow_manager.repositories.model.StepInput;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import lombok.Getter;

import java.util.List;
import java.util.Map;

@Getter
public class WorkflowDefinitionDTO {
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

    public List<Step> toSteps(Workflow workflow) {
        return steps.stream().map((step) -> {
            var stepBuilder = Step.builder();
            var stepEntry = step.entrySet().stream().findFirst().orElseThrow(() -> new IllegalArgumentException("No keys for step map"));
            var stepDto = stepEntry.getValue();
            stepBuilder.workflow(workflow)
                    .stepName(stepEntry.getKey())
                    .service(stepDto.service)
                    .task(stepDto.task);
            Step stepModel = stepBuilder.build();
            stepModel.setStepInputs(
                stepDto.input.entrySet()
                    .stream()
                    .map((input) ->
                        StepInput.builder()
                            .inputKey(input.getKey())
                            .inputValue(input.getValue())
                            .step(stepModel)
                            .build()
                    ).toList()
            );
            return stepModel;
        }).toList();
    }

    public Workflow toWorkflow() {
        var builder = Workflow.builder();
        builder.workflowName(name);
        Workflow workflow = builder.build();
        List<Argument> arguments = args.stream().map((argumentMap) -> {
            var argumentBuilder = Argument.builder();
            var defaultMap = argumentMap.entrySet().stream().findFirst().orElseThrow(() -> new IllegalArgumentException("No default argument"));
            argumentBuilder.argKey(defaultMap.getKey());
            if (!defaultMap.getValue().containsKey("default")) {
                throw new IllegalArgumentException("No default argument");
            }
            argumentBuilder.defaultValue(defaultMap.getValue().get("default"));
            argumentBuilder.workflow(workflow);
            return argumentBuilder.build();
        }).toList();

        workflow.setArgs(arguments);
        workflow.setSteps(toSteps(workflow));
        return workflow;
    }
 }
