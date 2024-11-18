package com.taskcomposer.service_registry;

import org.springframework.boot.SpringApplication;

public class TestServiceRegistryApplication {

	public static void main(String[] args) {
		SpringApplication.from(ServiceRegistryApplication::main).with(TestcontainersConfiguration.class).run(args);
	}

}
