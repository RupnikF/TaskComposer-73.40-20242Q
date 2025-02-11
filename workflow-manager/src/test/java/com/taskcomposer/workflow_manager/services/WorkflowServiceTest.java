package com.taskcomposer.workflow_manager.services;

import com.taskcomposer.workflow_manager.repositories.ServiceRepository;
import com.taskcomposer.workflow_manager.repositories.WorkflowRepository;
import com.taskcomposer.workflow_manager.repositories.model.Step;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import com.taskcomposer.workflow_manager.services.exceptions.ServiceNotFoundException;
import com.taskcomposer.workflow_manager.services.exceptions.TaskNotFoundException;
import com.taskcomposer.workflow_manager.services.exceptions.WorkflowAlreadyExistsException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;

import java.util.List;
import java.util.Optional;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@Tag("unit")
class WorkflowServiceTest {

    @Mock
    private WorkflowRepository workflowRepository;

    @Mock
    private ServiceRepository serviceRepository;

    @InjectMocks
    private WorkflowService workflowService;

    @BeforeEach
    void setUp() {
        MockitoAnnotations.openMocks(this);
    }

    @Test
    void testGetWorkflowById() {
        Long workflowId = 1L;
        Workflow workflow = new Workflow();
        workflow.setId(workflowId);

        when(workflowRepository.findById(workflowId)).thenReturn(Optional.of(workflow));

        Optional<Workflow> result = workflowService.getWorkflowById(workflowId);

        assertTrue(result.isPresent());
        assertEquals(workflowId, result.get().getId());
    }

    @Test
    void testGetWorkflowByName() {
        String workflowName = "TestWorkflow";
        Workflow workflow = new Workflow();
        workflow.setWorkflowName(workflowName);

        when(workflowRepository.findByWorkflowName(workflowName)).thenReturn(Optional.of(workflow));

        Optional<Workflow> result = workflowService.getWorkflowByName(workflowName);

        assertTrue(result.isPresent());
        assertEquals(workflowName, result.get().getWorkflowName());
    }

    @Test
    void testSaveWorkflow() throws WorkflowAlreadyExistsException {
        Workflow workflow = new Workflow();
        workflow.setWorkflowName("NewWorkflow");

        when(workflowRepository.existsByWorkflowName(workflow.getWorkflowName())).thenReturn(false);
        when(workflowRepository.save(workflow)).thenReturn(workflow);

        Workflow result = workflowService.saveWorkflow(workflow);

        assertNotNull(result);
        assertEquals("NewWorkflow", result.getWorkflowName());
    }

    @Test
    void testSaveWorkflowAlreadyExists() {
        Workflow workflow = new Workflow();
        workflow.setWorkflowName("ExistingWorkflow");

        when(workflowRepository.existsByWorkflowName(workflow.getWorkflowName())).thenReturn(true);

        assertThrows(WorkflowAlreadyExistsException.class, () -> {
            workflowService.saveWorkflow(workflow);
        });
    }

    @Test
    void testSaveWorkflowServiceNotFound() {
        Workflow workflow = new Workflow();
        workflow.setWorkflowName("NewWorkflow");
        Step step = new Step();
        step.setService("NonExistentService");
        workflow.setSteps(List.of(step));

        when(workflowRepository.existsByWorkflowName(workflow.getWorkflowName())).thenReturn(false);
        when(serviceRepository.getServiceByName(step.getService())).thenReturn(Optional.empty());

        assertThrows(ServiceNotFoundException.class, () -> {
            workflowService.saveWorkflow(workflow);
        });
    }

    @Test
    void testSaveWorkflowTaskNotFound() {
        Workflow workflow = new Workflow();
        workflow.setWorkflowName("NewWorkflow");
        Step step = new Step();
        step.setService("ExistingService");
        step.setTask("NonExistentTask");
        workflow.setSteps(List.of(step));

        com.taskcomposer.workflow_manager.repositories.model.Service service = new com.taskcomposer.workflow_manager.repositories.model.Service("ExistingService", Set.of("ExistingTask"));
        when(workflowRepository.existsByWorkflowName(workflow.getWorkflowName())).thenReturn(false);
        when(serviceRepository.getServiceByName(step.getService())).thenReturn(Optional.of(service));

        assertThrows(TaskNotFoundException.class, () -> {
            workflowService.saveWorkflow(workflow);
        });
    }

    @Test
    void testDeleteWorkflowById() {
        Long workflowId = 1L;
        Workflow workflow = new Workflow();
        workflow.setId(workflowId);

        when(workflowRepository.findById(workflowId)).thenReturn(Optional.of(workflow));

        boolean result = workflowService.deleteWorkflowById(workflowId);

        assertTrue(result);
        verify(workflowRepository, times(1)).delete(workflow);
    }
}
