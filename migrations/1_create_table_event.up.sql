CREATE TABLE events (
    uuid CHAR(36) PRIMARY KEY,
    name VARCHAR(32) NOT NULL,
    retry_count INT NOT NULL,
    transaction_id VARCHAR(255),
    initiator_legal_entity VARCHAR(255) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    consent_id CHAR(36),
    payload TEXT NOT NULL,
    error TEXT
);
