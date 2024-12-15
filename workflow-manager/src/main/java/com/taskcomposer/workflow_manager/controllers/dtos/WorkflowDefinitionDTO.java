package com.taskcomposer.workflow_manager.controllers.dtos;

import com.taskcomposer.workflow_manager.repositories.model.Argument;
import com.taskcomposer.workflow_manager.repositories.model.Step;
import com.taskcomposer.workflow_manager.repositories.model.StepInput;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import lombok.Getter;

import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

@Getter
public class WorkflowDefinitionDTO {
    private String name;
    private List<Map<String, Map<String, String>>> args;

    private List<Map<String, StepDTO>> steps;

    public List<Step> toSteps(Workflow workflow) {
        AtomicInteger stepCounter = new AtomicInteger(0);
        return steps.stream().map((step) -> {
            var stepBuilder = Step.builder();
            var stepEntry = step.entrySet().stream().findFirst().orElseThrow(() -> new IllegalArgumentException("No keys for step map"));
            var stepDto = stepEntry.getValue();
            stepBuilder.workflow(workflow)
                    .stepName(stepEntry.getKey())
                    .service(stepDto.getService())
                    .task(stepDto.getTask())
                    .stepOrder(stepCounter.getAndIncrement());
            Step stepModel = stepBuilder.build();

            var inputMap = stepDto.getInput();
            if (inputMap != null) {
                stepModel.setStepInputs(
                    stepDto.getInput().entrySet()
                        .stream()
                        .map((input) ->
                                StepInput.builder()
                                        .inputKey(input.getKey())
                                        .inputValue(input.getValue())
                                        .step(stepModel)
                                        .build()
                        ).toList()
                );
            }

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
