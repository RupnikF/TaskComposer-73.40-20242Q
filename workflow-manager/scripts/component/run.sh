#!/bin/bash

# Run
cd ../../

mvn test -Dgroup=component -DKAFKA_HOST=localhost -DKAFKA_TOPIC=test -DPOSTGRES_HOST=postgres -DPOSTGRES_PORT=5432 -DPOSTGRES_DB=workflow -DPOSTGRES_USER=workflow -DPOSTGRES_PASSWORD=workflow