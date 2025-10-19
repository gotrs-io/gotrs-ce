-- Auto-generated test data - DO NOT COMMIT TO GIT
-- Generated: 2025-08-27T13:32:03Z
-- This file is gitignored and should never be committed

DO $$
BEGIN
    IF current_setting('app.env', true) = 'production' THEN
        RAISE EXCEPTION 'Test data migration cannot be run in production';
    END IF;
END $$;

-- Test customer companies
INSERT INTO customer_company (customer_id, name, street, city, country, valid_id, create_time, create_by, change_time, change_by) VALUES
('COMP1', 'Acme Corporation', '123 Main St', 'New York', 'USA', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('COMP2', 'TechStart Inc', '456 Tech Ave', 'San Francisco', 'USA', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('COMP3', 'Global Services Ltd', '789 Business Park', 'London', 'UK', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1)
ON CONFLICT (customer_id) DO NOTHING;

-- Test agents (dynamically generated passwords)
-- || root@localhost / X2xSOca5gGy3-1Lq!1
-- || agent.smith / TRCzvGXJyGZJUf9s!1
-- || agent.jones / TS7GvNVOVmS9A7fF!1
INSERT INTO users (login, pw, first_name, last_name, valid_id, create_time, create_by, change_time, change_by) VALUES
('root@localhost', '$2a$12$iiVLeCJ5iIBHSGQP5KezReA83QYcF2F9sMylXirZNYsbMAheqVrVK', 'Admin', 'OTRS', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('agent.smith', '$2a$12$zipdGEyx4n1NEPmPg1kOxO9f4iXEx047dzBmouAO2L2afHgrjm6fa', 'Agent', 'Smith', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('agent.jones', '$2a$12$uV98Swzd4eQPpY1dCKUZBO9KDm3tjFvoklXTxFfrJHVRDQseqsJN2', 'Agent', 'Jones', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1)
ON CONFLICT (login) DO NOTHING;

-- Test customers (dynamically generated passwords)
-- || john.customer / Yq2PuMbRjW4JLQQK!1
-- || jane.customer / xUKsHdIZQdPrX9BP!1
-- || bob.customer / -QJ_XSXPl5Mrm1-5!1
INSERT INTO customer_user (login, email, customer_id, pw, first_name, last_name, phone, valid_id, create_time, create_by, change_time, change_by) VALUES
('john.customer', 'john@acme.com', 'COMP1', '$2a$12$GOnBz1NZApuDpBSIQ.veJ.38SUqrFhdTFPLBjA.ai3b8Qef8LEgDm', 'John', 'Customer', '555-0101', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('jane.customer', 'jane@techstart.com', 'COMP2', '$2a$12$VCx9Xt7LI7agRCthknqAC.0NAEB00TvTG29Ny66gGh.Eb62POXd4G', 'Jane', 'Customer', '555-0102', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('bob.customer', 'bob@global.com', 'COMP3', '$2a$12$rRx0BpPbyVDwPKsB4yOlUOqKM//KLwCBqySpPEUYCuH8dlWwZOsqS', 'Bob', 'Customer', '555-0103', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1)
ON CONFLICT (login) DO NOTHING;

-- Add agents to users group
INSERT INTO group_user (user_id, group_id, permission_key, create_time, create_by, change_time, change_by) 
SELECT id, 2, 'rw', CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1 FROM users WHERE login IN ('root@localhost', 'agent.smith', 'agent.jones')
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
