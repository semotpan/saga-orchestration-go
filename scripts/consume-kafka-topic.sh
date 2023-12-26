#!/bin/sh

echo
echo "Topic 'room-booking.inbox.events' messages:"
docker exec -t zookeeper  kafka-console-consumer --topic room-booking.inbox.events --from-beginning --bootstrap-server kafka:9092

echo
echo "Topic 'room-book.outbox.events' messages:"
docker exec -t zookeeper  kafka-console-consumer --topic room-booking.outbox.events --from-beginning --bootstrap-server kafka:9092

echo
echo "Topic 'payment.inbox.events' messages:"
docker exec -t zookeeper  kafka-console-consumer --topic payment.outbox.events --from-beginning --bootstrap-server kafka:9092

echo
echo "Topic 'payment.outbox.events' messages:"
docker exec -t zookeeper  kafka-console-consumer --topic payment.outbox.events --from-beginning --bootstrap-server kafka:9092
