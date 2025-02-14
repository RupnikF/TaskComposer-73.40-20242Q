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
        if(body.getParameters() != null && !body.getParameters().isEmpty()){
            //Validate Parameters
            var params = body.getParameters();
            if(params.get("cronDefinition") != null){
                //Check if the cron definition is valid, Ignore Intellisense redundancy
                if(!params.get("cronDefinition").matches("^(\\*|((\\*\\/)?[1-5]?[0-9])) (\\*|((\\*\\/)?[1-5]?[0-9])) (\\*|((\\*\\/)?(1?[0-9]|2[0-3]))) (\\*|((\\*\\/)?([1-9]|[12][0-9]|3[0-1]))) (\\*|((\\*\\/)?([1-9]|1[0-2]))) (\\*|((\\*\\/)?[0-6]))$")){
                    return ResponseEntity.badRequest().body("Invalid cron definition");
                }
            }
            if(params.get("delayed") != null){
                //Check if it's an integer larger than 0
                if (!params.get("delayed").matches("^\\d+$")){
                    return ResponseEntity.badRequest().body("Invalid delayed value");
                }
            }
        }
        Span span = tracer.spanBuilder("Start execution")
                .setAttribute("execution-workflow", body.getWorkflowName())
                .startSpan();
        log.info("Starting execution {}", body.getWorkflowName());
        Workflow workflow = workflowOptional.get();
        try {
            String executionUUID = triggerService.triggerWorkflow(workflow, body.getTags(), body.getParameters(), body.getArgs());
            return ResponseEntity.ok(executionUUID);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body("Missing required arguments. I do not tell you bc I'm lazy.");
        } finally {
            span.end();
        }
    }
}
