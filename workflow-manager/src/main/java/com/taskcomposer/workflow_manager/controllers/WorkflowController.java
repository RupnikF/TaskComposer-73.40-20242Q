package com.taskcomposer.workflow_manager.controllers;

import com.taskcomposer.workflow_manager.controllers.dtos.WorkflowDTO;
import com.taskcomposer.workflow_manager.services.WorkflowService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

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
    public Object uploadWorkflow(@RequestBody final WorkflowDTO body) {
        return workflowService.saveWorkflow(body.toWorkflow()).getId();
    }
}
