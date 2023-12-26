# Microservices Go Saga Orchestration - Use Case

###### Transactional Outbox + Change Data Capture with Debezium

---

A simple PoC of SAGA orchestration.

The implementation of the services are done with Golang and Apache Kafka.
* `Hotel Service`
* `Payment Service`
* `Reservation Service`

### Running services

Start the docker compose (`docker-compose.yaml`)

`src % docker-compose up`

Submit from e2e three scenarios:  

`e2e % ./room-reservation.sh`
