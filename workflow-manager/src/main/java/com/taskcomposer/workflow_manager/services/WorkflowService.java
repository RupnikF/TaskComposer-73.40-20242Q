package com.taskcomposer.workflow_manager.services;


import com.taskcomposer.workflow_manager.repositories.ServiceRepository;
import com.taskcomposer.workflow_manager.repositories.WorkflowRepository;
import com.taskcomposer.workflow_manager.repositories.model.Step;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import com.taskcomposer.workflow_manager.services.exceptions.ServiceNotFoundException;
import com.taskcomposer.workflow_manager.services.exceptions.TaskNotFoundException;
import com.taskcomposer.workflow_manager.services.exceptions.WorkflowAlreadyExistsException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

@Service
public class WorkflowService {

    private final WorkflowRepository workflowRepository;
    private final ServiceRepository serviceRepository;


    @Autowired
    public WorkflowService(WorkflowRepository workflowRepository, ServiceRepository serviceRepository) {
        this.workflowRepository = workflowRepository;
        this.serviceRepository = serviceRepository;
    }

    public Optional<Workflow> getWorkflowByName(String name) {
        return workflowRepository.findByWorkflowName(name);
    }

    public Optional<Workflow> getWorkflowById(Long id) {
        return workflowRepository.findById(id);
    }

    public List<Workflow> getAllWorkflows() {
        return workflowRepository.findAll();
    }

    public Workflow saveWorkflow(Workflow workflow) throws WorkflowAlreadyExistsException, ServiceNotFoundException, TaskNotFoundException {
        if (workflowRepository.existsByWorkflowName(workflow.getWorkflowName())) {
            throw new WorkflowAlreadyExistsException(workflow.getWorkflowName());
        }
        if (workflow.getSteps() != null) {
            for (Step step : workflow.getSteps()) {
                Optional<com.taskcomposer.workflow_manager.repositories.model.Service> service = serviceRepository.getServiceByName(step.getService());
                if (service.isEmpty()) {
                    throw new ServiceNotFoundException(step.getService());
                }
                if (!service.get().isTaskInService(step.getTask())) {
                    throw new TaskNotFoundException(step.getService(), step.getTask());
                }
            }
        }
        return workflowRepository.save(workflow);
    }

    public boolean deleteWorkflowById(Long id) {
        Optional<Workflow> workflow = workflowRepository.findById(id);
        if (workflow.isPresent()) {
            workflowRepository.delete(workflow.get());
            return true;
        }
        return false;
    }
}
