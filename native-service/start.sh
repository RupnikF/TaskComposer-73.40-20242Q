#!/bin/sh
./wait-for-it.sh broker:9092 --timeout=30 --strict -- echo "Broker is up"
exec ./app