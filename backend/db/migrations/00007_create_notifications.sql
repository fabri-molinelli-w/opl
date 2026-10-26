-- +goose Up
CREATE TABLE notifications (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id),
    event_type text NOT NULL
               CHECK (event_type IN ('service.down', 'incident.opened', 'incident.assigned', 'document.updated')),
    title      text NOT NULL,
    body       text NOT NULL DEFAULT '',
    link       text NOT NULL DEFAULT '',
    dedupe_key text NOT NULL,
    read_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notifications_user_created_idx ON notifications (user_id, created_at DESC);
CREATE INDEX notifications_created_at_idx ON notifications (created_at);
-- RN-2: la misma notificación no se repite a un usuario mientras tenga una igual sin leer.
CREATE UNIQUE INDEX notifications_unread_dedupe_uq
    ON notifications (user_id, dedupe_key)
    WHERE read_at IS NULL;

-- +goose Down
DROP TABLE notifications;
