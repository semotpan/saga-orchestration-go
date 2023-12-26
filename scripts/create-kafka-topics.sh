#!/usr/bin/env bash

echo
echo ">> hotel service"
echo "> Create topic 'room-booking.inbox.events'"
echo "----------------------------------------"
docker exec -t zookeeper kafka-topics --create --bootstrap-server kafka:9092 --replication-factor 1 --partitions 1 --topic room-booking.inbox.events

echo
echo "> Create topic 'room-booking.outbox.events'"
echo "-----------------------------------------"
docker exec -t zookeeper kafka-topics --create --bootstrap-server kafka:9092 --replication-factor 1 --partitions 1 --topic room-booking.outbox.events

echo
echo ">> payment service"
echo "> Create topic 'payment.inbox.events'"
echo "----------------------------------------"
docker exec -t zookeeper kafka-topics --create --bootstrap-server kafka:9092 --replication-factor 1 --partitions 1 --topic payment.inbox.events

echo
echo "> Create topic 'payment.outbox.events'"
echo "-----------------------------------------"
docker exec -t zookeeper kafka-topics --create --bootstrap-server kafka:9092 --replication-factor 1 --partitions 1 --topic payment.outbox.events
