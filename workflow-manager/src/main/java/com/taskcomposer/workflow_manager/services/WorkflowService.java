package com.taskcomposer.workflow_manager.services;


import com.taskcomposer.workflow_manager.repositories.WorkflowRepository;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import com.taskcomposer.workflow_manager.services.exceptions.WorkflowAlreadyExistsException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

@Service
public class WorkflowService {

    private final WorkflowRepository workflowRepository;


    @Autowired
    public WorkflowService(WorkflowRepository workflowRepository) {
        this.workflowRepository = workflowRepository;
    }

    public Optional<Workflow> getWorkflowByName(String name) {
        return workflowRepository.findByWorkflowName(name);
    }
    public List<Workflow> getAllWorkflows() {
        return workflowRepository.findAll();
    }
    public Workflow saveWorkflow(Workflow workflow) {
        if (workflowRepository.existsByWorkflowName(workflow.getWorkflowName())) {
            throw new WorkflowAlreadyExistsException(workflow.getWorkflowName());
        }
        return workflowRepository.save(workflow);
    }

    public void deleteWorkflowByName(String name) {
        Optional<Workflow> workflow = workflowRepository.findByWorkflowName(name);
        workflow.ifPresent(workflowRepository::delete);
    }
}
