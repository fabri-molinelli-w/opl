-- +goose Up
CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         text NOT NULL,
    name          text NOT NULL,
    password_hash text NOT NULL,
    role          text NOT NULL CHECK (role IN ('admin', 'sysadmin', 'soporte')),
    active        boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_uq ON users (lower(email));

-- +goose Down
DROP TABLE users;
