# Workflow Manager
Workflow manager that manages CRUD of workflows, validates and submits new executions of those workflows.

## Requirements
- Maven
- Java 23

## Endpoints:
```http
GET /workflows 
POST /workflows + Accepts:YAML
POST /triggers + Accepts:JSON
```

## Documentation
The examples folder has the basic format of a workflow to be submitted

The documentation folder has some http calls to be used for some cool workflows we created.