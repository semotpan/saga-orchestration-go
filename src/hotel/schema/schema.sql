CREATE TABLE IF NOT EXISTS hotel
(
    id            SERIAL PRIMARY KEY,
    creation_time TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name          VARCHAR(255) NOT NULL,
    address       VARCHAR(255) NOT NULL,
    location      VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS room
(
    id            SERIAL PRIMARY KEY,
    creation_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name          VARCHAR(255),
    number        INT       NOT NULL,
    floor         INT       NOT NULL,
    available     BOOL               DEFAULT TRUE,
    hotel_id      INT       NOT NULL,
    FOREIGN KEY (hotel_id) REFERENCES hotel (id) ON UPDATE CASCADE ON DELETE CASCADE
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

-- DEMO data
INSERT INTO hotel(id, name, address, location)
VALUES (1, 'Bristol Central Park Hotel', 'str. Puskin 32, 2012', 'Chişinău');

INSERT INTO room(id, name, number, floor, hotel_id, available)
VALUES (1, 'Twin with view', 38, 5, 1, true),
       (2, 'Deluxe', 25, 3, 1, false),
       (3, 'Twin Deluxe', 27, 2, 1, true);
