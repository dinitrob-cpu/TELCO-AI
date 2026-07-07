CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS metrics (
    time        TIMESTAMPTZ       NOT NULL,
    node_id     TEXT,
    edge_id     TEXT,
    metric_name TEXT              NOT NULL,
    value       DOUBLE PRECISION  NOT NULL,
    tags        JSONB
);
SELECT create_hypertable('metrics', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS anomaly_events (
    time            TIMESTAMPTZ NOT NULL,
    correlation_id  TEXT        NOT NULL,
    node_ids        TEXT[],
    edge_ids        TEXT[],
    severity        DOUBLE PRECISION,
    score           DOUBLE PRECISION,
    attribution     JSONB
);
SELECT create_hypertable('anomaly_events', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS faults (
    id                TEXT PRIMARY KEY,
    opened_at         TIMESTAMPTZ NOT NULL,
    closed_at         TIMESTAMPTZ,
    root_cause_node_id TEXT,
    hypothesis        JSONB,
    status            TEXT NOT NULL DEFAULT 'open'
);

CREATE TABLE IF NOT EXISTS qkd_health (
    time             TIMESTAMPTZ NOT NULL,
    link_id          TEXT NOT NULL,
    qber             DOUBLE PRECISION,
    key_rate_bps     DOUBLE PRECISION,
    sifting_ratio    DOUBLE PRECISION,
    drift            DOUBLE PRECISION
);
SELECT create_hypertable('qkd_health', 'time', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS topology_snapshots (
    version  BIGINT PRIMARY KEY,
    time     TIMESTAMPTZ NOT NULL,
    graph    JSONB NOT NULL
);
