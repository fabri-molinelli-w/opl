-- +goose Up
-- CU-06 RN-5: evento certificate.critical para las notificaciones internas (CU-08).
ALTER TABLE notifications DROP CONSTRAINT notifications_event_type_check;
ALTER TABLE notifications ADD CONSTRAINT notifications_event_type_check
    CHECK (event_type IN ('service.down', 'incident.opened', 'incident.assigned', 'document.updated', 'certificate.critical'));

-- +goose Down
DELETE FROM notifications WHERE event_type = 'certificate.critical';
ALTER TABLE notifications DROP CONSTRAINT notifications_event_type_check;
ALTER TABLE notifications ADD CONSTRAINT notifications_event_type_check
    CHECK (event_type IN ('service.down', 'incident.opened', 'incident.assigned', 'document.updated'));
