-- +goose Up
-- CU-06: certificados SSL/TLS a vigilar.
CREATE TABLE certificates (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    domain          text NOT NULL,
    port            integer NOT NULL DEFAULT 443 CHECK (port BETWEEN 1 AND 65535),
    issuer          text NOT NULL DEFAULT '',
    not_after       timestamptz,
    days_left       integer,
    status          text NOT NULL DEFAULT 'unknown'
                    CHECK (status IN ('unknown', 'ok', 'warning', 'critical')),
    error           text,
    last_checked_at timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX certificates_domain_port_uq ON certificates (lower(domain), port);

-- CU-07: historial de deploys (manuales o por webhook).
CREATE TABLE deploys (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id  uuid NOT NULL REFERENCES services(id),
    version     text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('prod', 'staging', 'dev')),
    author      text NOT NULL DEFAULT '',
    result      text NOT NULL CHECK (result IN ('success', 'failed', 'rollback')),
    notes       text NOT NULL DEFAULT '',
    source      text NOT NULL CHECK (source IN ('manual', 'webhook')),
    deployed_at timestamptz NOT NULL DEFAULT now(),
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX deploys_service_deployed_idx ON deploys (service_id, deployed_at DESC);
CREATE INDEX deploys_deployed_at_idx ON deploys (deployed_at DESC);

-- +goose Down
DROP TABLE deploys;
DROP TABLE certificates;
