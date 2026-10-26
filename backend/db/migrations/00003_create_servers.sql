-- +goose Up
CREATE TABLE servers (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hostname    text NOT NULL,
    ip          inet NOT NULL,
    environment text NOT NULL CHECK (environment IN ('prod', 'staging', 'dev')),
    os          text NOT NULL DEFAULT '',
    location    text NOT NULL DEFAULT 'onprem' CHECK (location IN ('onprem', 'cloud')),
    owner       text NOT NULL DEFAULT '',
    tags        text[] NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX servers_hostname_lower_uq
    ON servers (lower(hostname))
    WHERE deleted_at IS NULL;

CREATE INDEX servers_tags_gin ON servers USING gin (tags);
CREATE INDEX servers_environment_idx ON servers (environment) WHERE deleted_at IS NULL;
CREATE INDEX servers_os_idx ON servers (os) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE servers;
