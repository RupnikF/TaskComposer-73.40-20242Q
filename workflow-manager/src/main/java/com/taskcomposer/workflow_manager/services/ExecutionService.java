package com.taskcomposer.workflow_manager.services;

import com.taskcomposer.workflow_manager.controllers.dtos.ExecutionSubmissionDTO;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;
import java.util.logging.Logger;

@Service
public class ExecutionService {
    private final Logger log = Logger.getLogger(ExecutionService.class.getName());
    private final KafkaTemplate<String, Object> kafkaTemplate;

    @Value("${spring.kafka.template.default-topic}")
    private String kafkaTopic;

    public ExecutionService(KafkaTemplate<String, Object> kafkaTemplate) {
        this.kafkaTemplate = kafkaTemplate;
    }
//
    public String executeWorkflow(Workflow workflow, List<String> tags, Map<String, String> parameters, Map<String, String> args) {
        ExecutionSubmissionDTO submissionDTO = new ExecutionSubmissionDTO(workflow, tags, parameters, args);
        // TODO: send submissionDTO to queue of submissions
        log.info(kafkaTopic);
        var completableFuture = this.kafkaTemplate.send(kafkaTopic , submissionDTO);
        try {
            var result = completableFuture.get(1000, TimeUnit.MILLISECONDS);
            log.info(result.toString());
        } catch (InterruptedException | TimeoutException | ExecutionException e) {
            log.warning(e.getMessage());
        }
        return submissionDTO.getExecutionUUID();
    }
//
//    @KafkaListener(topics = "executions", groupId = "pepe")
//    public void listenGroupFoo(String message) {
//        System.out.println("Received Message in group pepe: " + message);
//    }

}
