CREATE TABLE events (
    uuid CHAR(36) PRIMARY KEY,
    name VARCHAR(32) NOT NULL,
    retry_count INT NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    consent_id CHAR(36),
    custodian VARCHAR(255) NOT NULL,
    payload TEXT NOT NULL,
    error TEXT
);
