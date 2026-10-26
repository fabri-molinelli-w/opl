-- +goose Up
-- CU-03: estado del scheduler por servicio.
ALTER TABLE services
    ADD COLUMN last_checked_at      timestamptz,
    ADD COLUMN consecutive_failures integer NOT NULL DEFAULT 0 CHECK (consecutive_failures >= 0),
    -- RN-10: un servicio con el monitoreo pausado (mantenimiento, datos de demo) no se chequea.
    ADD COLUMN check_enabled        boolean NOT NULL DEFAULT true;

CREATE TABLE health_checks (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id  uuid NOT NULL REFERENCES services(id),
    status      text NOT NULL CHECK (status IN ('UP', 'DOWN', 'DEGRADED')),
    response_ms integer,
    http_status integer,
    error       text,
    checked_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX health_checks_service_checked_idx ON health_checks (service_id, checked_at DESC);
CREATE INDEX health_checks_checked_at_idx ON health_checks (checked_at);

-- CU-10: incidentes y su línea de tiempo.
CREATE TABLE incidents (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title        text NOT NULL,
    description  text NOT NULL DEFAULT '',
    service_id   uuid REFERENCES services(id),
    severity     text NOT NULL CHECK (severity IN ('SEV1', 'SEV2', 'SEV3', 'SEV4')),
    status       text NOT NULL DEFAULT 'open'
                 CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')),
    assignee_id  uuid REFERENCES users(id),
    root_cause   text,
    auto_created boolean NOT NULL DEFAULT false,
    opened_at    timestamptz NOT NULL DEFAULT now(),
    resolved_at  timestamptz,
    closed_at    timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX incidents_status_idx ON incidents (status);
CREATE INDEX incidents_severity_idx ON incidents (severity);
CREATE INDEX incidents_service_id_idx ON incidents (service_id);
CREATE INDEX incidents_assignee_id_idx ON incidents (assignee_id);
-- RN-4: como máximo un incidente automático no cerrado por servicio (idempotencia del scheduler).
CREATE UNIQUE INDEX incidents_one_open_auto_per_service_uq
    ON incidents (service_id)
    WHERE auto_created AND status <> 'closed';

CREATE TABLE incident_events (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id),
    author_id   uuid REFERENCES users(id),
    kind        text NOT NULL CHECK (kind IN ('status_change', 'note', 'system')),
    body        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX incident_events_incident_idx ON incident_events (incident_id, created_at);
CREATE INDEX incident_events_author_idx ON incident_events (author_id);

-- +goose Down
DROP TABLE incident_events;
DROP TABLE incidents;
DROP TABLE health_checks;
ALTER TABLE services DROP COLUMN check_enabled, DROP COLUMN consecutive_failures, DROP COLUMN last_checked_at;
