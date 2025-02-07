package com.taskcomposer.workflow_manager.grpc_services;

import io.grpc.stub.StreamObserver;
import io.grpc.Status;
import com.taskcomposer.workflow_manager.TriggerRequest;
import com.taskcomposer.workflow_manager.TriggerResponse;
import com.taskcomposer.workflow_manager.WorkflowTriggerServiceGrpc;
import com.taskcomposer.workflow_manager.services.WorkflowService;
import com.taskcomposer.workflow_manager.services.TriggerService;
import com.taskcomposer.workflow_manager.repositories.model.Workflow;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import net.devh.boot.grpc.server.service.GrpcService;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@GrpcService
public class TriggerServant extends WorkflowTriggerServiceGrpc.WorkflowTriggerServiceImplBase {
    private final TriggerService triggerService;
    private final WorkflowService workflowService;
    private final Tracer tracer;
    private final Logger log = LogManager.getLogger(TriggerServant.class);

    @Autowired
    public TriggerServant(TriggerService triggerService, WorkflowService workflowService, OpenTelemetry openTelemetry) {
        this.triggerService = triggerService;
        this.workflowService = workflowService;
        this.tracer = openTelemetry.tracerBuilder(com.taskcomposer.workflow_manager.controllers.TriggerController.class.getName()).build();
    }


    @Override
    public void triggerWorkflow(TriggerRequest request, StreamObserver<TriggerResponse> responseObserver) {
        Optional<Workflow> workflowOptional = workflowService.getWorkflowByName(request.getWorkflowName());
        if (workflowOptional.isEmpty()) {
            responseObserver.onError(Status.NOT_FOUND.withDescription(String.format("Workflow %s not found", request.getWorkflowName())).asRuntimeException());
            return;
        }
        Span span = tracer.spanBuilder("Start execution")
                .setAttribute("execution-workflow", request.getWorkflowName())
                .startSpan();
        log.info("Starting execution {}", request.getWorkflowName());
        Workflow workflow = workflowOptional.get();
        List<String> tagsList = request.getTagsList().stream().toList();
        Map<String, String> parametersMap = new HashMap<>();
        request.getParametersList().forEach((keyValue -> parametersMap.put(keyValue.getKey(), keyValue.getValue())));
        Map<String, String> argsMap = new HashMap<>();
        request.getArgsList().forEach((keyValue -> argsMap.put(keyValue.getKey(), keyValue.getValue())));
        try {
            String executionUUID = triggerService.triggerWorkflow(workflow, tagsList, parametersMap, argsMap);
            responseObserver.onNext(TriggerResponse.newBuilder().setUuid(executionUUID).build());
        } catch (IllegalArgumentException e) {
            responseObserver.onError(Status.INVALID_ARGUMENT.withDescription(e.getMessage()).asException());
        } finally {
            span.end();
            responseObserver.onCompleted();
        }
    }
}
