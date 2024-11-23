package com.taskcomposer.workflow_manager.controllers;

import com.taskcomposer.workflow_manager.controllers.dtos.WorkflowDefinitionDTO;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import com.taskcomposer.workflow_manager.services.WorkflowService;
import com.taskcomposer.workflow_manager.services.exceptions.WorkflowAlreadyExistsException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.support.ServletUriComponentsBuilder;

import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/workflows")
public class WorkflowController {

    private final WorkflowService workflowService;

    @Autowired
    public WorkflowController(final WorkflowService workflowService) {
        this.workflowService = workflowService;
    }

    @PostMapping(
        consumes = {"application/yaml", "application/x-yaml"},
        produces = {"application/json"}
    )
    public ResponseEntity<Workflow> uploadWorkflow(@RequestBody final WorkflowDefinitionDTO body) {
        try {
            Workflow createdWorkflow = workflowService.saveWorkflow(body.toWorkflow());
            URI location = ServletUriComponentsBuilder.fromCurrentRequest().path("/{id}").buildAndExpand(createdWorkflow.getId()).toUri();
            return ResponseEntity.created(location).body(createdWorkflow);
        } catch (WorkflowAlreadyExistsException _) {
            return new ResponseEntity<>(HttpStatus.CONFLICT);
        }
    }

    @GetMapping(
        produces = {"application/json"}
    )
    public ResponseEntity<List<Workflow>> getAllWorkflows(@RequestParam(name = "name", required = false) Optional<String> name) {
        List<Workflow> workflows;
        if (name.isPresent()) {
            Optional<Workflow> workflow = workflowService.getWorkflowByName(name.get());
            workflows = new ArrayList<>();
            workflow.ifPresent(workflows::add);
        }
        else {
            workflows = workflowService.getAllWorkflows();
        }

        if (workflows.isEmpty()) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.ok().body(workflows);
    }

    @GetMapping(
        path = "/{id}",
        produces = {"application/json"}
    )
    public ResponseEntity<Workflow> getWorkflow(@PathVariable(name = "id") final Long id) {
        Optional<Workflow> workflow = workflowService.getWorkflowById(id);

        if (workflow.isEmpty()) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok().body(workflow.get());
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteWorkflow(@PathVariable(name = "id") final Long id) {
        boolean deleted = workflowService.deleteWorkflowById(id);
        if (deleted) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }
}
