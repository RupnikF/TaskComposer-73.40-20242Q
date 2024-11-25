package com.taskcomposer.workflow_manager.repositories;

import com.taskcomposer.workflow_manager.repositories.model.Service;

import java.util.Optional;
import java.util.Set;

public class ServiceRepositoryImpl implements ServiceRepository{
    @Override
    public Optional<Service> getServiceByName(String name) {
        return Optional.of(new Service("echo", Set.of("echo")));
    }
}
