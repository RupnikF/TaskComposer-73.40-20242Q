package com.taskcomposer.workflow_manager.services;

import com.taskcomposer.workflow_manager.controllers.dtos.ExecutionSubmissionDTO;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

@Service
public class ExecutionService {

//    private final KafkaTemplate<String, String> kafkaTemplate;
//
//    public ExecutionService(KafkaTemplate<String, String> kafkaTemplate) {
//        this.kafkaTemplate = kafkaTemplate;
//    }
//
    public void executeWorkflow(Workflow workflow, List<String> tags, Map<String, String> parameters, Map<String, String> args) {
        ExecutionSubmissionDTO submissionDTO = new ExecutionSubmissionDTO(workflow, tags, parameters, args);
        // TODO: send submissionDTO to queue of submissions
    }
//
//    @KafkaListener(topics = "executions", groupId = "pepe")
//    public void listenGroupFoo(String message) {
//        System.out.println("Received Message in group pepe: " + message);
//    }

}
