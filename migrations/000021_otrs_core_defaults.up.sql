-- ============================================
-- OTRS Core Default Data
-- Additional default data that OTRS typically ships with
-- Based on OTRS standard installation
-- ============================================

-- ============================================
-- Communication Channels
-- ============================================
DELETE FROM communication_channel WHERE id > 0;
ALTER SEQUENCE communication_channel_id_seq RESTART WITH 1;

INSERT INTO communication_channel (name, module, package_name, valid_id, create_by, change_by) VALUES
('Email', 'Kernel::System::CommunicationChannel::Email', 'Framework', 1, 1, 1),
('Phone', 'Kernel::System::CommunicationChannel::Phone', 'Framework', 1, 1, 1),
('Internal', 'Kernel::System::CommunicationChannel::Internal', 'Framework', 1, 1, 1),
('Chat', 'Kernel::System::CommunicationChannel::Chat', 'Framework', 1, 1, 1),
('SMS', 'Kernel::System::CommunicationChannel::SMS', 'Framework', 1, 1, 1);

-- ============================================
-- Auto Response Types
-- ============================================
DELETE FROM auto_response_type WHERE id > 0;
ALTER SEQUENCE auto_response_type_id_seq RESTART WITH 1;

INSERT INTO auto_response_type (name, comments, valid_id, create_by, change_by) VALUES
('auto reply', 'Automatic reply sent on new ticket creation', 1, 1, 1),
('auto reject', 'Automatic rejection for follow-ups on closed tickets', 1, 1, 1),
('auto follow up', 'Automatic response for follow-ups', 1, 1, 1),
('auto reply/new ticket', 'Automatic reply for new tickets', 1, 1, 1),
('auto remove', 'Automatic removal notification', 1, 1, 1);

-- ============================================
-- Auto Responses
-- ============================================
DELETE FROM auto_response WHERE id > 0;
ALTER SEQUENCE auto_response_id_seq RESTART WITH 1;

INSERT INTO auto_response (name, text0, text1, type_id, system_address_id, valid_id, comments, create_by, change_by) VALUES
('default reply (after new ticket has been created)', 
'Thank you for your email.

You wrote:
<OTRS_CUSTOMER_EMAIL[50]>

Your email will be answered by a human agent as soon as possible.

Do not reply to this email. This mailbox is not monitored and you will not receive a response.

<OTRS_CONFIG_NotificationSenderName>',
'RE: <OTRS_CUSTOMER_SUBJECT>',
1, 1, 1, 'Default auto reply for new tickets', 1, 1),

('default reject (after follow-up and rejected of a closed ticket)', 
'Your previous ticket [<OTRS_TICKET_TicketNumber>] is closed.

Unfortunately we could not detect a valid ticket number in your subject, so this email cannot be processed.

Please create a new ticket via the customer portal.

<OTRS_CONFIG_NotificationSenderName>',
'RE: <OTRS_CUSTOMER_SUBJECT>',
2, 1, 1, 'Default rejection for follow-ups on closed tickets', 1, 1);

-- ============================================
-- Roles (if not already populated)
-- ============================================
-- Check if roles exist, if not add default ones
INSERT INTO roles (name, comments, valid_id, create_by, change_by) 
SELECT 'Agent', 'Default agent role', 1, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Agent');

INSERT INTO roles (name, comments, valid_id, create_by, change_by) 
SELECT 'Admin', 'Administrator role with full access', 1, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Admin');

INSERT INTO roles (name, comments, valid_id, create_by, change_by) 
SELECT 'Customer', 'Customer portal access role', 1, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Customer');

-- ============================================
-- Services (more defaults if needed)
-- ============================================
-- Only add if no services exist
INSERT INTO service (name, valid_id, comments, create_by, change_by)
SELECT 'IT Support', 1, 'General IT support services', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM service WHERE name = 'IT Support');

INSERT INTO service (name, valid_id, comments, create_by, change_by)
SELECT 'IT Support::Hardware', 1, 'Hardware support services', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM service WHERE name = 'IT Support::Hardware');

INSERT INTO service (name, valid_id, comments, create_by, change_by)
SELECT 'IT Support::Software', 1, 'Software support services', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM service WHERE name = 'IT Support::Software');

INSERT INTO service (name, valid_id, comments, create_by, change_by)
SELECT 'IT Support::Network', 1, 'Network support services', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM service WHERE name = 'IT Support::Network');

-- ============================================
-- SLA Templates (more defaults if needed)
-- ============================================
-- Only add basic SLAs if none exist
INSERT INTO sla (name, valid_id, comments, create_by, change_by)
SELECT 'Standard', 1, 'Standard service level - 8 hour response', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM sla WHERE name = 'Standard');

INSERT INTO sla (name, valid_id, comments, create_by, change_by)
SELECT 'Premium', 1, 'Premium service level - 4 hour response', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM sla WHERE name = 'Premium');

INSERT INTO sla (name, valid_id, comments, create_by, change_by)
SELECT 'Gold', 1, 'Gold service level - 2 hour response', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM sla WHERE name = 'Gold');

INSERT INTO sla (name, valid_id, comments, create_by, change_by)
SELECT 'Critical', 1, 'Critical service level - 1 hour response', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM sla WHERE name = 'Critical');

-- ============================================
-- Article Types (if table exists and is empty)
-- Note: article_type table may not exist in all OTRS versions
-- ============================================
-- Check if article_type table exists and populate if empty
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'article_type') THEN
        DELETE FROM article_type WHERE id > 0;
        
        INSERT INTO article_type (id, name, comments, valid_id, create_by, change_by) VALUES
        (1, 'email-external', 'Email to external customer', 1, 1, 1),
        (2, 'email-internal', 'Internal email between agents', 1, 1, 1),
        (3, 'email-notification-ext', 'External notification email', 1, 1, 1),
        (4, 'email-notification-int', 'Internal notification email', 1, 1, 1),
        (5, 'phone', 'Phone call article', 1, 1, 1),
        (6, 'fax', 'Fax article', 1, 1, 1),
        (7, 'sms', 'SMS article', 1, 1, 1),
        (8, 'webrequest', 'Web request article', 1, 1, 1),
        (9, 'note-internal', 'Internal note', 1, 1, 1),
        (10, 'note-external', 'External note', 1, 1, 1),
        (11, 'note-report', 'Report note', 1, 1, 1),
        (12, 'chat-external', 'External chat', 1, 1, 1),
        (13, 'chat-internal', 'Internal chat', 1, 1, 1);
    END IF;
END $$;

-- ============================================
-- Link Auto Responses to Queues (examples)
-- ============================================
-- Link default auto responses to Postmaster queue
INSERT INTO queue_auto_response (queue_id, auto_response_id, auto_response_type_id, create_by, change_by)
SELECT 1, 1, 1, 1, 1  -- Postmaster queue, default reply, auto reply type
WHERE NOT EXISTS (
    SELECT 1 FROM queue_auto_response 
    WHERE queue_id = 1 AND auto_response_type_id = 1
);

INSERT INTO queue_auto_response (queue_id, auto_response_id, auto_response_type_id, create_by, change_by)
SELECT 1, 2, 2, 1, 1  -- Postmaster queue, default reject, auto reject type
WHERE NOT EXISTS (
    SELECT 1 FROM queue_auto_response 
    WHERE queue_id = 1 AND auto_response_type_id = 2
);

-- ============================================
-- Update sequences
-- ============================================
SELECT setval('communication_channel_id_seq', (SELECT COALESCE(MAX(id), 1) FROM communication_channel), true);
SELECT setval('auto_response_type_id_seq', (SELECT COALESCE(MAX(id), 1) FROM auto_response_type), true);
SELECT setval('auto_response_id_seq', (SELECT COALESCE(MAX(id), 1) FROM auto_response), true);
SELECT setval('roles_id_seq', (SELECT COALESCE(MAX(id), 1) FROM roles), true);
SELECT setval('service_id_seq', (SELECT COALESCE(MAX(id), 1) FROM service), true);
SELECT setval('sla_id_seq', (SELECT COALESCE(MAX(id), 1) FROM sla), true);