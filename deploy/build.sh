#!/bin/bash
eval $(minikube docker-env)
docker build -t workflow-manager-service:latest ../workflow-manager