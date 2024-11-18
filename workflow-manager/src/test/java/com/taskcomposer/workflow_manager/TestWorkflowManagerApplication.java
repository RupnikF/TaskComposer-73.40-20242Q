package com.taskcomposer.workflow_manager;

import org.springframework.boot.SpringApplication;

public class TestWorkflowManagerApplication {

	public static void main(String[] args) {
		SpringApplication.from(WorkflowManagerApplication::main).with(TestcontainersConfiguration.class).run(args);
	}

}
