package com.taskcomposer.workflow_manager.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.io.FileReader;
import java.util.Map;
import java.util.List;

@Configuration
public class ServiceConfig {
    private final static String SERVICE_DEFINITION_PATH = "/data/services.json";

    @Bean
    public ServiceContainer getServices() {
        ObjectMapper objectMapper = new ObjectMapper();
        try (var file = new FileReader(SERVICE_DEFINITION_PATH)) {
            return objectMapper.readValue(file, ServiceContainer.class);
        } catch (Exception e) {
            e.printStackTrace();
            return new ServiceContainer(Map.of());
        }
    }


    public record ServiceContainer(Map<String, ServiceDefinition> services) {
    }

    public record ServiceDefinition(String name, String server, String inputTopic, String outputTopic,
                                    List<String> tasks) {
    }
}
