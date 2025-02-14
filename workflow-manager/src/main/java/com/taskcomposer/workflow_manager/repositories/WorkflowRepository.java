package com.taskcomposer.workflow_manager.repositories;

import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface WorkflowRepository extends JpaRepository<Workflow,Long> {
    // Find Workflow by its unique name
    Optional<Workflow> findByWorkflowName(String workflowName);

    // Check if a Workflow exists by its name
    boolean existsByWorkflowName(String workflowName);
}
