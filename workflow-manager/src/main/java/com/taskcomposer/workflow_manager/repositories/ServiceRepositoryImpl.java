package com.taskcomposer.workflow_manager.repositories;

import com.taskcomposer.workflow_manager.config.ServiceConfig;
import com.taskcomposer.workflow_manager.repositories.model.Service;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Repository;

import java.util.HashSet;
import java.util.Optional;
import java.util.Set;

@Repository
public class ServiceRepositoryImpl implements ServiceRepository{
    private final ServiceConfig.ServiceContainer serviceContainer;

    @Autowired
    public ServiceRepositoryImpl(
            ServiceConfig.ServiceContainer serviceContainer
            ) {
        this.serviceContainer = serviceContainer;
    }

    @Override
    public Optional<Service> getServiceByName(String name) {
        var service = serviceContainer.services().get(name);
        if (service == null) {
            return Optional.empty();
        }
        return Optional.of(new Service(service.name(), new HashSet<>(service.tasks())));
    }
}
