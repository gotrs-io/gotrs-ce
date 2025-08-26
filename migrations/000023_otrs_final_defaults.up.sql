-- ============================================
-- OTRS Final Default Data
-- Remaining default data for complete OTRS compatibility
-- ============================================

-- ============================================
-- Link Types (ticket relationships)
-- ============================================
DELETE FROM link_type WHERE id > 0;
ALTER SEQUENCE link_type_id_seq RESTART WITH 1;

INSERT INTO link_type (name, valid_id, create_by, change_by) VALUES
('Normal', 1, 1, 1),
('ParentChild', 1, 1, 1),
('DependsOn', 1, 1, 1),
('RelevantTo', 1, 1, 1),
('AlternativeTo', 1, 1, 1),
('ConnectedTo', 1, 1, 1),
('DuplicateOf', 1, 1, 1);

-- ============================================
-- Article Flags (for marking articles)
-- ============================================
-- Note: Article flags are usually created per-article dynamically
-- The table uses a composite primary key, no sequence
-- This is typically populated when articles are created/viewed

-- ============================================
-- Notification Events (if not already populated)
-- ============================================
DELETE FROM notification_event WHERE id > 0;
ALTER SEQUENCE notification_event_id_seq RESTART WITH 1;

-- Basic notification events that OTRS typically includes
INSERT INTO notification_event (name, valid_id, comments, create_by, change_by) VALUES
('Ticket::NewTicket', 1, 'Notification for new tickets', 1, 1),
('Ticket::FollowUp', 1, 'Notification for ticket follow-ups', 1, 1),
('Ticket::QueueUpdate', 1, 'Notification when ticket queue changes', 1, 1),
('Ticket::StateUpdate', 1, 'Notification when ticket state changes', 1, 1),
('Ticket::OwnerUpdate', 1, 'Notification when ticket owner changes', 1, 1),
('Ticket::ResponsibleUpdate', 1, 'Notification when ticket responsible changes', 1, 1),
('Ticket::PriorityUpdate', 1, 'Notification when ticket priority changes', 1, 1),
('Ticket::CustomerUpdate', 1, 'Notification when ticket customer changes', 1, 1),
('Ticket::PendingTimeUpdate', 1, 'Notification when ticket pending time changes', 1, 1),
('Ticket::LockUpdate', 1, 'Notification when ticket lock status changes', 1, 1),
('Ticket::ArchiveFlagUpdate', 1, 'Notification when ticket archive flag changes', 1, 1),
('Ticket::EscalationTimeStart', 1, 'Notification when ticket escalation starts', 1, 1),
('Ticket::EscalationTimeStop', 1, 'Notification when ticket escalation stops', 1, 1),
('Ticket::EscalationTimeNotifyBefore', 1, 'Notification before ticket escalation', 1, 1),
('Article::NewArticle', 1, 'Notification for new articles', 1, 1),
('Article::ArticleSend', 1, 'Notification when article is sent', 1, 1);

-- ============================================
-- Default Admin User (if not exists)
-- ============================================
-- Ensure there's at least one admin user
-- Note: Password should be changed immediately after installation
INSERT INTO users (login, pw, first_name, last_name, valid_id, create_by, change_by) 
SELECT 'root@localhost', '$2a$10$EzSiBKx1VrvVcECFKsGog.W1OcYvhfKQBm3mh2kLO5qwpULIAYf4a', 'Admin', 'OTRS', 1, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM users WHERE login = 'root@localhost');

-- Link admin user to admin group if exists
INSERT INTO group_user (user_id, group_id, permission_key, permission_value, create_by, change_by) 
SELECT u.id, g.id, 'rw', 1, 1, 1
FROM users u, groups g
WHERE u.login = 'root@localhost' 
AND g.name = 'admin'
AND NOT EXISTS (
    SELECT 1 FROM group_user ug 
    WHERE ug.user_id = u.id 
    AND ug.group_id = g.id 
    AND ug.permission_key = 'rw'
);

-- ============================================
-- Default Groups (ensure core groups exist)
-- ============================================
INSERT INTO groups (name, comments, valid_id, create_by, change_by) 
SELECT 'admin', 'Admin Group', 1, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM groups WHERE name = 'admin');

INSERT INTO groups (name, comments, valid_id, create_by, change_by) 
SELECT 'stats', 'Stats Group', 1, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM groups WHERE name = 'stats');

-- ============================================
-- Process States (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'process_state') THEN
        INSERT INTO process_state (entity_id, name, state_type, create_by, change_by) 
        SELECT 'S1', 'Active', 'Active', 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM process_state WHERE entity_id = 'S1');
        
        INSERT INTO process_state (entity_id, name, state_type, create_by, change_by) 
        SELECT 'S2', 'Inactive', 'Inactive', 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM process_state WHERE entity_id = 'S2');
        
        INSERT INTO process_state (entity_id, name, state_type, create_by, change_by) 
        SELECT 'S3', 'FadeAway', 'FadeAway', 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM process_state WHERE entity_id = 'S3');
    END IF;
END $$;

-- ============================================
-- Dynamic Field Object Types (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'dynamic_field_obj_type') THEN
        INSERT INTO dynamic_field_obj_type (name, display_name, valid_id, create_by, change_by) 
        SELECT 'Ticket', 'Ticket', 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM dynamic_field_obj_type WHERE name = 'Ticket');
        
        INSERT INTO dynamic_field_obj_type (name, display_name, valid_id, create_by, change_by) 
        SELECT 'Article', 'Article', 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM dynamic_field_obj_type WHERE name = 'Article');
        
        INSERT INTO dynamic_field_obj_type (name, display_name, valid_id, create_by, change_by) 
        SELECT 'CustomerUser', 'Customer User', 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM dynamic_field_obj_type WHERE name = 'CustomerUser');
        
        INSERT INTO dynamic_field_obj_type (name, display_name, valid_id, create_by, change_by) 
        SELECT 'CustomerCompany', 'Customer Company', 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM dynamic_field_obj_type WHERE name = 'CustomerCompany');
    END IF;
END $$;

-- ============================================
-- Update sequences
-- ============================================
SELECT setval('link_type_id_seq', (SELECT COALESCE(MAX(id), 1) FROM link_type), true);
SELECT setval('notification_event_id_seq', (SELECT COALESCE(MAX(id), 1) FROM notification_event), true);
SELECT setval('groups_id_seq', (SELECT COALESCE(MAX(id), 1) FROM groups), true);
SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 1) FROM users), true);