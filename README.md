# Saga Orchestration Pattern: Use Case Implementation

###### Transactional Outbox + Change Data Capture with Debezium

###### _Tech Stack: Golang and Apache Kafka_

---

This use case demonstrates how to implement the [**SAGA Pattern**](https://microservices.io/patterns/data/saga.html) to achieve distributed transactions across multiple microservices, assuming the implementation utilizes a [**Database per Service**](https://microservices.io/patterns/data/database-per-service.html) approach. The solution employs the [**Transactional Outbox Pattern**](https://microservices.io/patterns/data/transactional-outbox.html) in conjunction with Change Data Capture (CDC) and [**Debezium**](https://debezium.io/documentation/reference/stable/transformations/outbox-event-router.html) on top of Kafka Connect, and, [**Apache Kafka**](https://www.confluent.io/what-is-apache-kafka/) serves as the messaging backbone for the overall architecture.

### Context:

> **Use case**: _As a guest, I want to make a hotel reservation_

There are 3 microservices involved:
* `Reservation Service` - initiator and orchestrator of the Saga
* `Hotel Service` - ensure the hotel's room requested is available and book it or rejects it if unavailable
* `Payment Service` - executes the payment associated to an incoming reservation

And we have the following service configuration:
![Context Overview](assets/contextoverview.png)

### Task: 
Use the [**SAGA Orchestration Pattern**](https://microservices.io/post/sagas/2019/12/12/developing-sagas-part-4.html) to ensure data consistency between services.

### Implementation
The reservation process is as follows:
![Reservation State](assets/reservationstate.png)

The SAGA framework state machine happy path:
![SagaStateMachine](assets/sagastatemachine.png)

A detailed component workflow:
![Use Case Overview](assets/usecaseoverview.png)

A happy path for making a reservation is illustrated in the following _sequence diagram_:
![Happy Path](assets/statediagramhappypath.png)

An unhappy path of making a reservation, when the payment is rejected, and the _compensation_ step is involved you can see in the following _sequence diagram_:
![Unhappy Path](assets/statediagramunhappypath.png)

## Running the Use Case

Start the docker compose (`docker-compose.yaml`)
```console
/src % docker-compose up --build
```
During the build process there is a _kafka-setup_ service which creates kafka topics and submit the debezium connectors

Make a reservation
```console
% http POST http://localhost:8080/api/v1/reservations < e2e/room-reservation-placement.json

HTTP/1.1 202
Location: http://localhost:8080/api/v1/reservations/6656d80e-edf0-42ab-b114-ced902bafa36
Retry-After: 0.5
```

Examine the emitted event for `room-booking.inbox.events` in Apache Kafka:
```console
% docker run --tty --rm \
    --network go-saga-orchestration-network \
    quay.io/debezium/tooling:1.2 \
    kafkacat -b kafka:9092 -C -o beginning -q \
    -f "{\"key\":%k, \"headers\":\"%h\"}\n%s\n" \
    -t room-booking.inbox.events 
    
{"key":057ceada-02b3-4a65-beb3-3de54d6e29f3, "headers":"id=aed1882e-c6d2-402c-985f-e40d19d4df43,eventType=REQUEST"}
{"type":"REQUEST","roomId":1,"endDate":"2023-12-17","guestId":10000001,"hotelId":1,"startDate":"2023-12-16","paymentDue":1702632793441,"creditCardNo":"************7999","reservationId":"6656d80e-edf0-42ab-b114-ced902bafa36"}
```

and, the emitted event for `payment.inbox.events` in Apache Kafka:
```console
% docker run --tty --rm \
    --network go-saga-orchestration-network \
    quay.io/debezium/tooling:1.2 \
    kafkacat -b kafka:9092 -C -o beginning -q \
    -f "{\"key\":%k, \"headers\":\"%h\"}\n%s\n" \
    -t payment.inbox.events 
    
{"key":057ceada-02b3-4a65-beb3-3de54d6e29f3, "headers":"id=e3280316-16b8-4114-98ee-9443d523d2ea,eventType=REQUEST"}
{"type":"REQUEST","roomId":1,"endDate":"2023-12-17","guestId":10000001,"hotelId":1,"startDate":"2023-12-16","paymentDue":1702632793441,"creditCardNo":"************7999","reservationId":"6656d80e-edf0-42ab-b114-ced902bafa36"}
```

Query the saga state in the reservation service's database:
```console
% docker run --tty --rm -i \
        --network go-saga-orchestration-network \
        quay.io/debezium/tooling:1.2 \
        bash -c 'pgcli postgresql://reservationuser:secret@reservation-db:5432/reservationdb'

SELECT * FROM public.sagastate;

+--------------------------------------+-----------+------------------+---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+----------------+-------------------------------------------------------+---------------+
| id                                   | version   | type             | payload                                                                                                                                                                                                                                           | current_step   | step_status                                           | saga_status   |
|--------------------------------------+-----------+------------------+---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+----------------+-------------------------------------------------------+---------------|
| 057ceada-02b3-4a65-beb3-3de54d6e29f3 | 3         | room-reservation | {"type": "REQUEST", "roomId": 1, "endDate": "2023-12-17", "guestId": 10000001, "hotelId": 1, "startDate": "2023-12-16", "paymentDue": 1702632793441, "creditCardNo": "************7999", "reservationId": "6656d80e-edf0-42ab-b114-ced902bafa36"} |                | {"payment": "SUCCEEDED", "room-booking": "SUCCEEDED"} | COMPLETED     |
+--------------------------------------+-----------+------------------+---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+----------------+-------------------------------------------------------+---------------+
```

And, also, query the location URL:
```console
% http GET http://localhost:8080/api/v1/reservations/6656d80e-edf0-42ab-b114-ced902bafa36 
HTTP/1.1 200 
Content-Type: application/json

{
    "guestId": 10000001,
    "hotelId": 1,
    "reservationId": "6656d80e-edf0-42ab-b114-ced902bafa36",
    "roomId": 1,
    "status": "SUCCEED"
}
```

#### Checkout `e2e` folder with some unhappy scenarios
