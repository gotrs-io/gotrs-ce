-- Auto-generated test data - DO NOT COMMIT TO GIT
-- Generated: 2025-08-19T11:26:23Z
-- This file is gitignored and should never be committed

DO $$
BEGIN
    IF current_setting('app.env', true) = 'production' THEN
        RAISE EXCEPTION 'Test data migration cannot be run in production';
    END IF;
END $$;

-- Test customer companies
INSERT INTO customer_company (customer_id, name, street, city, country, valid_id, create_by, change_by) VALUES
('COMP1', 'Acme Corporation', '123 Main St', 'New York', 'USA', 1, 1, 1),
('COMP2', 'TechStart Inc', '456 Tech Ave', 'San Francisco', 'USA', 1, 1, 1),
('COMP3', 'Global Services Ltd', '789 Business Park', 'London', 'UK', 1, 1, 1)
ON CONFLICT (customer_id) DO NOTHING;

-- Test agents (dynamically generated passwords)
INSERT INTO users (login, pw, first_name, last_name, valid_id, create_by, change_by) VALUES
('admin', '$2a$12$Xt4vBgG2aAmBKE37mzAKr.bX3bXNKxtCwA8nT4oztCf5iuPTSzmYC', 'Admin', 'User', 1, 1, 1),
('agent.smith', '$2a$12$Z1ZvZKLw8hIsW9QnpREI5uhBYXQQNmo2/tvhwFoAONmfXzGEktR56', 'Agent', 'Smith', 1, 1, 1),
('agent.jones', '$2a$12$Mvmv3gE7s5OY9nYY4jiLJOpdPuC3Dy8u4f9VWnbNhngz07ie0uVFe', 'Agent', 'Jones', 1, 1, 1)
ON CONFLICT (login) DO NOTHING;

-- Test customers (dynamically generated passwords)
INSERT INTO customer_user (login, email, customer_id, pw, first_name, last_name, phone, valid_id, create_by, change_by) VALUES
('john.customer', 'john@acme.com', 'COMP1', '$2a$12$X.FXxRhIAKGyuuVl./HQXusDhdmJBGCOedv3TiUvLp8vyR.4js/IG', 'John', 'Customer', '555-0101', 1, 1, 1),
('jane.customer', 'jane@techstart.com', 'COMP2', '$2a$12$Bf8.AgvlSC6aENaeeCDSbO01l6pRXVLCsA7JLxL8/NyhASRYbDWLa', 'Jane', 'Customer', '555-0102', 1, 1, 1),
('bob.customer', 'bob@global.com', 'COMP3', '$2a$12$bgyeBsZZAlyW883BZa1eSuf8mzE82KPiAShGijSg5eWitlxjgPf/i', 'Bob', 'Customer', '555-0103', 1, 1, 1)
ON CONFLICT (login) DO NOTHING;

-- Add agents to users group
INSERT INTO user_groups (user_id, group_id, permission_key, permission_value, create_by, change_by) 
SELECT id, 2, 'rw', 1, 1, 1 FROM users WHERE login IN ('admin', 'agent.smith', 'agent.jones')
ON CONFLICT DO NOTHING;

-- Sample tickets (use subqueries for user_id references)
INSERT INTO ticket (tn, title, queue_id, ticket_lock_id, user_id, ticket_priority_id, ticket_state_id, customer_id, customer_user_id, create_by, change_by) VALUES
('2025010000001', 'Cannot login to system', 4, 1, (SELECT id FROM users WHERE login = 'agent.smith'), 3, 1, 'COMP1', 'john.customer', 1, 1),
('2025010000002', 'Request for new feature', 4, 1, (SELECT id FROM users WHERE login = 'agent.jones'), 2, 2, 'COMP2', 'jane.customer', 1, 1),
('2025010000003', 'System running slow', 4, 1, (SELECT id FROM users WHERE login = 'agent.smith'), 4, 2, 'COMP1', 'john.customer', 1, 1),
('2025010000004', 'Password reset needed', 4, 1, (SELECT id FROM users WHERE login = 'agent.jones'), 3, 1, 'COMP3', 'bob.customer', 1, 1),
('2025010000005', 'API documentation request', 4, 1, (SELECT id FROM users WHERE login = 'agent.smith'), 2, 3, 'COMP2', 'jane.customer', 1, 1)
ON CONFLICT (tn) DO NOTHING;

-- Sample articles for tickets
INSERT INTO article (ticket_id, article_sender_type_id, communication_channel_id, is_visible_for_customer, create_by, change_by)
SELECT id, 3, 1, 1, 1, 1 FROM ticket
ON CONFLICT DO NOTHING;

-- Add article content (a_body is bytea, incoming_time is integer Unix timestamp)
INSERT INTO article_data_mime (article_id, a_from, a_to, a_subject, a_body, incoming_time, create_by, change_by)
SELECT 
    a.id,
    cu.email,
    'support@example.com',
    t.title,
    CAST(('Initial ticket description for: ' || t.title) AS bytea),
    EXTRACT(EPOCH FROM t.create_time)::integer,
    1,
    1
FROM article a
JOIN ticket t ON a.ticket_id = t.id
JOIN customer_user cu ON t.customer_user_id = cu.login
ON CONFLICT DO NOTHING;
