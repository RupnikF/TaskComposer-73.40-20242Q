package com.taskcomposer.workflow_manager.controllers;

import com.taskcomposer.workflow_manager.controllers.dtos.ExecutionRequestDTO;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import com.taskcomposer.workflow_manager.services.ExecutionService;
import com.taskcomposer.workflow_manager.services.WorkflowService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Optional;

@RestController
@RequestMapping("/executions")
public class ExecutionController {

    private final ExecutionService executionService;
    private final WorkflowService workflowService;

    @Autowired
    public ExecutionController(ExecutionService executionService, WorkflowService workflowService) {
        this.executionService = executionService;
        this.workflowService = workflowService;
    }

    @PostMapping(
        consumes = {"application/json"}
    )
    public ResponseEntity<String> startExecution(@RequestBody final ExecutionRequestDTO body) {
        Optional<Workflow> workflowOptional = workflowService.getWorkflowByName(body.getWorkflowName());
        if (workflowOptional.isEmpty()) {
            return ResponseEntity.badRequest().body("Workflow not found");
        }
        Workflow workflow = workflowOptional.get();
        //executionService.executeWorkflow(workflow);
        return ResponseEntity.ok().build();
    }
}
