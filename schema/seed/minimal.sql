-- Minimal seed data for development environment
-- This creates just enough data to have a working system

BEGIN;

-- Admin user (password: admin)
-- Password hash for 'admin': SHA256 hash
INSERT INTO users (id, login, pw, first_name, last_name, valid_id, create_time, create_by, change_time, change_by) VALUES
(1, 'admin@localhost', '8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918', 'System', 'Administrator', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1)
ON CONFLICT (id) DO NOTHING;

-- Grant admin permissions
INSERT INTO group_user (user_id, group_id, permission_key, permission_value, create_time, create_by, change_time, change_by) VALUES
(1, 1, 'rw', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
(1, 2, 'rw', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
(1, 3, 'rw', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1)
ON CONFLICT DO NOTHING;

-- Set sequence starting points
SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1));
SELECT setval('groups_id_seq', COALESCE((SELECT MAX(id) FROM groups), 3));
SELECT setval('queue_id_seq', COALESCE((SELECT MAX(id) FROM queue), 4));
SELECT setval('ticket_priority_id_seq', COALESCE((SELECT MAX(id) FROM ticket_priority), 5));
SELECT setval('ticket_state_id_seq', COALESCE((SELECT MAX(id) FROM ticket_state), 5));
SELECT setval('ticket_type_id_seq', COALESCE((SELECT MAX(id) FROM ticket_type), 5));

COMMIT;