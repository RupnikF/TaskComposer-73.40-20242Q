package com.taskcomposer.workflow_manager.repositories;

import com.taskcomposer.workflow_manager.repositories.model.Service;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.Set;

@Repository
public class ServiceRepositoryImpl implements ServiceRepository{
    @Override
    public Optional<Service> getServiceByName(String name) {
        if (name.equals("native")) {
            return Optional.of(new Service("native", Set.of("if", "abort")));
        }
        else if (name.equals("s3_service")) {
            return Optional.of(new Service("s3_service", Set.of("upload", "download")));
        }
        return Optional.of(new Service("echo", Set.of("echo")));
    }
}
