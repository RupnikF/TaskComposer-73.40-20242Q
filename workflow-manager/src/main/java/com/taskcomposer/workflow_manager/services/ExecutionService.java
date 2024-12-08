package com.taskcomposer.workflow_manager.services;

import com.taskcomposer.workflow_manager.controllers.ExecutionController;
import com.taskcomposer.workflow_manager.controllers.dtos.ExecutionSubmissionDTO;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import jakarta.inject.Inject;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

@Service
public class ExecutionService {
    private final Logger log = LogManager.getLogger(ExecutionService.class);
    private final KafkaTemplate<String, Object> kafkaTemplate;
    private final Tracer tracer;

    @Value("${spring.kafka.template.default-topic}")
    private String kafkaTopic;

    @Inject
    public ExecutionService(KafkaTemplate<String, Object> kafkaTemplate, OpenTelemetry openTelemetry) {
        this.kafkaTemplate = kafkaTemplate;
        this.tracer = openTelemetry.tracerBuilder(ExecutionController.class.getName()).build();
    }
//
    public String executeWorkflow(Workflow workflow, List<String> tags, Map<String, String> parameters, Map<String, String> args) {
        ExecutionSubmissionDTO submissionDTO = new ExecutionSubmissionDTO(workflow, tags, parameters, args);
        // TODO: send submissionDTO to queue of submissions
        Span span = this.tracer.spanBuilder("Execution Service").startSpan();
        span.setAttribute("execution-topic", kafkaTopic);
        var completableFuture = this.kafkaTemplate.send(kafkaTopic , submissionDTO);
        try {
            var result = completableFuture.get(1000, TimeUnit.MILLISECONDS);
            log.debug(result.toString());
        } catch (InterruptedException | TimeoutException | ExecutionException e) {
            log.warn(e.getMessage());
        }
        span.end();
        return submissionDTO.getExecutionUUID();
    }
//
//    @KafkaListener(topics = "executions", groupId = "pepe")
//    public void listenGroupFoo(String message) {
//        System.out.println("Received Message in group pepe: " + message);
//    }

}
