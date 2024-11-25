package com.taskcomposer.workflow_manager.repositories;

import com.taskcomposer.workflow_manager.repositories.model.Service;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface ServiceRepository {
    Optional<Service> getServiceByName(String name);
}
