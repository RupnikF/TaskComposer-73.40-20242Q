#!/bin/bash

# Run
cd ../../

./mvnw dependency:go-offline -B

./mvnw test -Dgroup=functional -DKAFKA_HOST=localhost -DKAFKA_TOPIC=test -DPOSTGRES_HOST=postgres -DPOSTGRES_PORT=5432 -DPOSTGRES_DB=workflow -DPOSTGRES_USER=workflow -DPOSTGRES_PASSWORD=workflow