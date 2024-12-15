package com.taskcomposer.workflow_manager.controllers;

import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import com.taskcomposer.workflow_manager.services.WorkflowService;
import com.taskcomposer.workflow_manager.services.TriggerService;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Optional;

@RestController
@RequestMapping("/triggers")
public class TriggerController {
    private final TriggerService triggerService;
    private final WorkflowService workflowService;
    private final Tracer tracer;
    private final Logger log = LogManager.getLogger(com.taskcomposer.workflow_manager.controllers.TriggerController.class);

    @Autowired
    public TriggerController(TriggerService triggerService, WorkflowService workflowService, OpenTelemetry openTelemetry) {
        this.triggerService = triggerService;
        this.workflowService = workflowService;
        this.tracer = openTelemetry.tracerBuilder(com.taskcomposer.workflow_manager.controllers.TriggerController.class.getName()).build();
    }

    @PostMapping(
        consumes = {"application/json"}
    )
    public ResponseEntity<String> triggerWorkflow(@RequestBody final com.taskcomposer.workflow_manager.controllers.dtos.TriggerRequestDTO body) {
        Optional<Workflow> workflowOptional = workflowService.getWorkflowByName(body.getWorkflowName());
        if (workflowOptional.isEmpty()) {
            return ResponseEntity.badRequest().body("Workflow not found");
        }
        Span span = tracer.spanBuilder("Start execution")
                .setAttribute("execution-workflow", body.getWorkflowName())
                .startSpan();
        log.info("Starting execution {}", body.getWorkflowName());
        Workflow workflow = workflowOptional.get();
        String executionUUID = triggerService.triggerWorkflow(workflow, body.getTags(), body.getParameters(), body.getArgs());
        span.end();
        return ResponseEntity.ok(executionUUID);
    }
}
