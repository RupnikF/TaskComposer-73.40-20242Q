#!/bin/bash

cd ../../

docker build . -t taskcomposer/ubuntu-service-unit-test -f Dockerfile.unit
docker run --rm taskcomposer/ubuntu-service-unit-test