package com.taskcomposer.workflow_manager.controllers;

import com.taskcomposer.workflow_manager.controllers.dtos.ExecutionRequestDTO;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import com.taskcomposer.workflow_manager.services.ExecutionService;
import com.taskcomposer.workflow_manager.services.WorkflowService;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.logs.LoggerProvider;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.slf4j.LoggerFactory;
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
    private final Tracer tracer;
    private final Logger log = LogManager.getLogger(ExecutionController.class);

    @Autowired
    public ExecutionController(ExecutionService executionService, WorkflowService workflowService, OpenTelemetry openTelemetry) {
        this.executionService = executionService;
        this.workflowService = workflowService;
        this.tracer = openTelemetry.tracerBuilder(ExecutionController.class.getName()).build();
    }

    @PostMapping(
        consumes = {"application/json"}
    )
    public ResponseEntity<String> startExecution(@RequestBody final ExecutionRequestDTO body) {
        Optional<Workflow> workflowOptional = workflowService.getWorkflowByName(body.getWorkflowName());
        if (workflowOptional.isEmpty()) {
            return ResponseEntity.badRequest().body("Workflow not found");
        }
        Span span = tracer.spanBuilder("Start execution")
                .setAttribute("execution-workflow", body.getWorkflowName())
                .startSpan();
        log.info("Starting execution {}", body.getWorkflowName());
        Workflow workflow = workflowOptional.get();
        String executionUUID = executionService.executeWorkflow(workflow, body.getTags(), body.getParameters(), body.getArgs());
        span.end();
        return ResponseEntity.ok(executionUUID);
    }
}
