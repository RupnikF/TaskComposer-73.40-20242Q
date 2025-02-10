#!/bin/bash

cd ../../

docker build . -t taskcomposer/scheduler-unit-test -f Dockerfile.unit
docker run --rm taskcomposer/scheduler-unit-test