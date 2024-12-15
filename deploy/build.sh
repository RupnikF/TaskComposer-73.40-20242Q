#!/bin/bash
eval $(minikube docker-env)
docker build -t workflow-manager-service:latest ../workflow-manager
docker build -t scheduler-service:latest ../scheduler
docker build -t echo-service:latest ../echo-service