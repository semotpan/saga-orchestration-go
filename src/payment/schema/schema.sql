CREATE TABLE IF NOT EXISTS payment
(
    reservation_id UUID PRIMARY KEY,
    timestamp      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    guest_id       INT         NOT NULL,
    payment_due    BIGINT      NOT NULL,
    credit_card_no VARCHAR(16) NOT NULL,
    type           VARCHAR(20) NOT NULL
);

-- Infrastructure tables
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
