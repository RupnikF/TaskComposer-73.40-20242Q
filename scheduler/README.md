# Scheduler
Core of the task scheduling process, receives workflow submissions and queues each step to be executed.

It is able to scale independent of the amount of instances, and uses Kafka to push messages to the relevant services.

Stores the intermediate results of each services, injects relevant properties as inputs for services, executes conditionals (currently bugged) and termination.

Also runs cron jobs and delayed submissions, although we didn't implement automatic leadership reselection due to time constraints. Shouldn't be difficult, its just another thread checking the lock on the etcd.

## Requirements
- Go (1.23)
- Docker

## Installation
- go mod tidy
- go test

### Debugging in Golang
https://www.rookout.com/blog/golang-debugging-tutorial/