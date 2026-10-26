-- +goose Up
CREATE TABLE services (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id            uuid NOT NULL REFERENCES servers(id),
    name                 text NOT NULL,
    type                 text NOT NULL CHECK (type IN ('web', 'db', 'queue', 'other')),
    port                 integer NOT NULL CHECK (port BETWEEN 1 AND 65535),
    check_type           text NOT NULL CHECK (check_type IN ('http', 'tcp')),
    check_target         text NOT NULL,
    expected_status      integer NOT NULL DEFAULT 200
                         CHECK (expected_status BETWEEN 100 AND 599),
    interval_seconds     integer NOT NULL DEFAULT 30 CHECK (interval_seconds > 0),
    latency_threshold_ms integer NOT NULL DEFAULT 1000 CHECK (latency_threshold_ms > 0),
    current_status       text NOT NULL DEFAULT 'UNKNOWN'
                         CHECK (current_status IN ('UP', 'DOWN', 'DEGRADED', 'UNKNOWN')),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);

CREATE INDEX services_server_id_idx ON services (server_id);
CREATE INDEX services_status_idx ON services (current_status) WHERE deleted_at IS NULL;
CREATE INDEX services_type_idx ON services (type) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE services;
