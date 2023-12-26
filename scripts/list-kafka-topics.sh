#!/bin/sh

echo ">> Kafka Topics:"
docker exec -t zookeeper kafka-topics --list --bootstrap-server kafka:9092
