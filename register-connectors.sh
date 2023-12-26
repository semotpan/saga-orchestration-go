#!/bin/sh

http PUT http://localhost:8083/connectors/reservation-outbox-connector/config < connectors/reservation-outbox-connector.json
http PUT http://localhost:8083/connectors/hotel-outbox-connector/config < connectors/hotel-outbox-connector.json
http PUT http://localhost:8083/connectors/payment-outbox-connector/config < connectors/payment-outbox-connector.json
