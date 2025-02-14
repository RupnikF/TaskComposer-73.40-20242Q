package com.taskcomposer.workflow_manager.services;

import com.taskcomposer.workflow_manager.controllers.dtos.ExecutionSubmissionDTO;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.context.propagation.TextMapPropagator;
import io.opentelemetry.context.propagation.TextMapSetter;
import jakarta.inject.Inject;
import org.apache.kafka.clients.producer.ProducerRecord;
import org.apache.kafka.common.header.Header;
import org.apache.kafka.common.header.internals.RecordHeader;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import javax.annotation.Nullable;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

@Service
public class TriggerService {
    private final Logger log = LogManager.getLogger(com.taskcomposer.workflow_manager.services.TriggerService.class);
    private final KafkaTemplate<String, Object> kafkaTemplate;
    private final Tracer tracer;
    private final OpenTelemetry openTelemetry;

    @Value("${spring.kafka.template.default-topic}")
    private String kafkaTopic;

    @Inject
    public TriggerService(KafkaTemplate<String, Object> kafkaTemplate, OpenTelemetry openTelemetry) {
        this.kafkaTemplate = kafkaTemplate;
        this.tracer = openTelemetry.tracerBuilder(com.taskcomposer.workflow_manager.controllers.TriggerController.class.getName()).build();
        this.openTelemetry = openTelemetry;
    }
//

    private boolean validateWorkflow(Workflow workflow, Map<String, String> args) {
        boolean valid = true;
        for (var arg : workflow.getArgs()) {
            if (arg.getDefaultValue() == null) {
                valid = valid && args.containsKey(arg.getArgKey());
            }
        }
        return valid;
    }

    public String triggerWorkflow(Workflow workflow, List<String> tags, Map<String, String> parameters, Map<String, String> args) throws IllegalArgumentException {
        if (!validateWorkflow(workflow, args)) {
            throw new IllegalArgumentException("Invalid workflow");
        }
        ExecutionSubmissionDTO submissionDTO = new ExecutionSubmissionDTO(workflow, tags, parameters, args);
        Span span = this.tracer.spanBuilder("Trigger Service").startSpan();
        span.setAttribute("execution-topic", kafkaTopic);
        span.setAttribute("tags", String.valueOf(tags));
        List<Header> headers = new ArrayList<>();

        openTelemetry.getPropagators().getTextMapPropagator().inject(io.opentelemetry.context.Context.current().with(span), headers, (headers1, s, s1) -> {
            if (headers1 == null) return;
            headers1.add(new RecordHeader(s, s1.getBytes()));
        });

        ProducerRecord<String, Object> producerRecord = new ProducerRecord<String, Object>(kafkaTopic, null, null, submissionDTO, headers);
        var completableFuture = this.kafkaTemplate.send(producerRecord);
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
