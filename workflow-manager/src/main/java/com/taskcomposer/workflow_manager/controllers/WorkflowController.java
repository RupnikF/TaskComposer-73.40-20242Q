package com.taskcomposer.workflow_manager.controllers;

import com.taskcomposer.workflow_manager.controllers.dtos.WorkflowDefinitionDTO;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import com.taskcomposer.workflow_manager.services.WorkflowService;
import com.taskcomposer.workflow_manager.services.exceptions.ServiceNotFoundException;
import com.taskcomposer.workflow_manager.services.exceptions.TaskNotFoundException;
import com.taskcomposer.workflow_manager.services.exceptions.WorkflowAlreadyExistsException;
import com.taskcomposer.workflow_manager.services.exceptions.WorkflowNotFoundException;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
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
    private final Logger log = LogManager.getLogger(WorkflowController.class);
    private final WorkflowService workflowService;

    @Autowired
    public WorkflowController(final WorkflowService workflowService) {
        this.workflowService = workflowService;
    }

    @PostMapping(
        consumes = {"application/yaml", "application/x-yaml"},
        produces = {"application/json"}
    )
    public ResponseEntity<Object> uploadWorkflow(@RequestBody final WorkflowDefinitionDTO body) {
        try {
            log.info("Uploading workflow of name {}", body.getName());
            Workflow createdWorkflow = workflowService.saveWorkflow(body.toWorkflow());
            URI location = ServletUriComponentsBuilder.fromCurrentRequest().path("/{id}").buildAndExpand(createdWorkflow.getId()).toUri();
            return ResponseEntity.created(location).body(createdWorkflow);
        } catch (WorkflowAlreadyExistsException e) {
            log.warn("Workflow {} already exists", body.getName());
            return ResponseEntity.status(HttpStatus.CONFLICT).body(e.getMessage());
        } catch (ServiceNotFoundException | TaskNotFoundException e) {
            log.info("Service or task not found");
            return ResponseEntity.badRequest().body(e.getMessage());
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

        return workflow.map(value -> ResponseEntity.ok().body(value)).orElseGet(() -> ResponseEntity.notFound().build());
    }

    @PutMapping(path = "/{id}", produces = {"application/json"}, consumes = {"application/yaml", "application/x-yaml"})
    public ResponseEntity<Object> putWorkflow(@PathVariable(name = "id") final Long id, @RequestBody final WorkflowDefinitionDTO body) {
        Workflow createdWorkflow = body.toWorkflow();
        try {
            return ResponseEntity.ok(this.workflowService.updateWorkflow(id, createdWorkflow));
        } catch (WorkflowNotFoundException e) {
            return ResponseEntity.notFound().build();
        } catch (ServiceNotFoundException | TaskNotFoundException e) {
            return ResponseEntity.badRequest().body(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteWorkflow(@PathVariable(name = "id") final Long id) {
        boolean deleted = workflowService.deleteWorkflowById(id);
        if (deleted) {
            log.info("Workflow {} deleted", id);
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }
}
