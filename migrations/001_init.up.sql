CREATE TABLE IF NOT EXISTS max_values (
    uuid  VARCHAR(36) PRIMARY KEY,
    ts          TIMESTAMP NOT NULL,
    max_value   BIGINT    NOT NULL
    );

CREATE INDEX IF NOT EXISTS max_values_ts_idx ON max_values (ts);