package com.taskcomposer.workflow_manager.services;

import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

@Service
public class ExecutionService {

//    private final KafkaTemplate<String, String> kafkaTemplate;
//
//    public ExecutionService(KafkaTemplate<String, String> kafkaTemplate) {
//        this.kafkaTemplate = kafkaTemplate;
//    }
//
//    public void executeWorkflow() {
//        try {
//            System.out.println(kafkaTemplate.send("executions", "hola").get());
//        }catch (Exception e){
//            e.printStackTrace();
//        }
//    }
//
//    @KafkaListener(topics = "executions", groupId = "pepe")
//    public void listenGroupFoo(String message) {
//        System.out.println("Received Message in group pepe: " + message);
//    }

}
