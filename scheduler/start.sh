#!/bin/sh

echo "Connecting to $KAFKA_HOST:$KAFKA_PORT"

./wait-for-it.sh $KAFKA_HOST:$KAFKA_PORT --timeout=30 --strict -- echo "Broker is up"
exec ./app