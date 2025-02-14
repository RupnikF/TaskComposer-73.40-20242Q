package com.taskcomposer.workflow_manager.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.taskcomposer.workflow_manager.controllers.WorkflowController;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.io.FileNotFoundException;
import java.io.FileReader;
import java.io.IOException;
import java.util.Map;
import java.util.List;

@Configuration
public class ServiceConfig {
    @Value("${service.definition.path:/data/services.json}")
    private String SERVICE_DEFINITION_PATH;
    private static final Logger log = LogManager.getLogger(ServiceConfig.class);

    @Bean
    public ServiceContainer getServices() {
        ObjectMapper objectMapper = new ObjectMapper();
        try (var file = new FileReader(SERVICE_DEFINITION_PATH)) {
            return objectMapper.readValue(file, ServiceContainer.class);
        } catch (IOException e) {
            log.warn("File /data/services.json not found");
            return new ServiceContainer(Map.of());
        }
    }


    public record ServiceContainer(Map<String, ServiceDefinition> services) {
    }

    public record ServiceDefinition(String name, String server, String inputTopic, String outputTopic,
                                    List<String> tasks) {
    }
}
