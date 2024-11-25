package com.taskcomposer.workflow_manager.controllers.dtos;

import com.taskcomposer.workflow_manager.repositories.model.Step;
import com.taskcomposer.workflow_manager.repositories.model.StepInput;
import lombok.Getter;
import lombok.Setter;

import java.util.HashMap;
import java.util.Map;

@Getter
@Setter
public class StepDTO {
    private String service;
    private String task;
    private Map<String, String> input;

    public static StepDTO fromStep(Step step) {
        StepDTO stepDTO = new StepDTO();
        Map<String, String> stepInputs = new HashMap<>();
        for (StepInput stepInput : step.getStepInputs()) {
            stepInputs.put(stepInput.getInputKey(), stepInput.getInputValue());
        }
        stepDTO.setInput(stepInputs);
        stepDTO.setTask(step.getTask());
        stepDTO.setService(step.getService());
        return stepDTO;
    }
}
