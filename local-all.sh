#!/bin/bash

cd scheduler
docker compose up --build &
cd ..
cd workflow-manager
docker compose up --build &
cd ..
cd echo-service
docker compose up --build &
cd ..
cd s3-service
docker compose up --build &
cd ..
cd ubuntu-service
docker compose up --build &
cd ..
