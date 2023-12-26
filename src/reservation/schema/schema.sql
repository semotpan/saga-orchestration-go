CREATE TABLE IF NOT EXISTS reservation
(
    id             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    timestamp      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    hotel_id       BIGINT      NOT NULL,
    room_id        INT         NOT NULL,
    start_date     DATE        NOT NULL,
    end_date       DATE        NOT NULL,
    status         VARCHAR(20) NOT NULL,
    guest_id       BIGINT      NOT NULL,
    payment_due    BIGINT      NOT NULL,
    credit_card_no VARCHAR(16) NOT NULL
);

-- Infrastructure tables
CREATE TABLE IF NOT EXISTS sagastate
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version      int8         NOT NULL,
    type         VARCHAR(100) NOT NULL,
    payload      JSONB        NOT NULL,
    current_step VARCHAR(100),
    step_status  JSONB,
    saga_status  VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS eventlog
(
    event_id  UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    issued_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS outboxevent
(
    id            UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    timestamp     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    aggregatetype VARCHAR(100) NOT NULL,
    aggregateid   VARCHAR(100) NOT NULL,
    type          VARCHAR(100) NOT NULL,
    payload       JSONB        NOT NULL
);

ALTER TABLE outboxevent
    REPLICA IDENTITY FULL;
