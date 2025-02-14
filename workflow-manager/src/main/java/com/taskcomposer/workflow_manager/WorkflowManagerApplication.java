package com.taskcomposer.workflow_manager;

import io.opentelemetry.sdk.autoconfigure.spi.AutoConfigurationCustomizerProvider;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.actuate.autoconfigure.security.servlet.ManagementWebSecurityAutoConfiguration;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.domain.EntityScan;
import org.springframework.boot.autoconfigure.security.servlet.SecurityAutoConfiguration;
import org.springframework.cloud.consul.ConsulAutoConfiguration;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;

@SpringBootApplication(exclude = {
		ConsulAutoConfiguration.class,
		SecurityAutoConfiguration.class,
		ManagementWebSecurityAutoConfiguration.class
})
@ComponentScan(basePackages = {
		"com.taskcomposer.workflow_manager.controllers",
		"com.taskcomposer.workflow_manager.services",
		"com.taskcomposer.workflow_manager.config",
		"com.taskcomposer.workflow_manager.repositories",
		"com.taskcomposer.workflow_manager.grpc_services"
})
@EntityScan( basePackages = "com.taskcomposer.workflow_manager.repositories.models")
public class WorkflowManagerApplication {

	public static void main(String[] args) {
		SpringApplication.run(WorkflowManagerApplication.class, args);
	}
}
