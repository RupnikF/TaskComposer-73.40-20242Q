package com.taskcomposer.workflow_manager.repositories;

import com.taskcomposer.workflow_manager.repositories.model.Service;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.Set;

@Repository
public class ServiceRepositoryImpl implements ServiceRepository{
    @Override
    public Optional<Service> getServiceByName(String name) {
        return Optional.of(new Service("echo", Set.of("echo")));
    }
}
