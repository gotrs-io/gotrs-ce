-- ============================================
-- OTRS Configuration Defaults
-- Additional configuration and reference data
-- ============================================

-- ============================================
-- Time Accounting Settings (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'time_accounting_action') THEN
        DELETE FROM time_accounting_action WHERE id > 0;
        
        INSERT INTO time_accounting_action (name, valid_id, create_by, change_by) VALUES
        ('Phone Call', 1, 1, 1),
        ('Email', 1, 1, 1),
        ('Research', 1, 1, 1),
        ('Meeting', 1, 1, 1),
        ('Documentation', 1, 1, 1),
        ('Testing', 1, 1, 1),
        ('Development', 1, 1, 1),
        ('Other', 1, 1, 1);
    END IF;
END $$;

-- ============================================
-- Process Management States (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'pm_process_state') THEN
        INSERT INTO pm_process_state (id, name, create_by, change_by) 
        SELECT 'S1', 'Active', 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM pm_process_state WHERE name = 'Active');
        
        INSERT INTO pm_process_state (id, name, create_by, change_by) 
        SELECT 'S2', 'Inactive', 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM pm_process_state WHERE name = 'Inactive');
        
        INSERT INTO pm_process_state (id, name, create_by, change_by) 
        SELECT 'S3', 'FadeAway', 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM pm_process_state WHERE name = 'FadeAway');
    END IF;
END $$;

-- ============================================
-- Standard Templates (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'standard_template') THEN
        DELETE FROM standard_template WHERE id > 0;
        
        INSERT INTO standard_template (name, content_type, text, template_type, valid_id, create_by, change_by) VALUES
        ('empty answer', 'text/plain; charset=utf-8', '', 'Answer', 1, 1, 1),
        ('Thank you for your email', 'text/plain; charset=utf-8', 
'Thank you for your email.

We are currently processing your request and will respond as soon as possible.

Best regards,
Your Support Team', 'Answer', 1, 1, 1),
        ('Your request has been closed', 'text/plain; charset=utf-8',
'Your request has been successfully resolved and closed.

If you have any further questions, please don''t hesitate to contact us again.

Best regards,
Your Support Team', 'Answer', 1, 1, 1);
    END IF;
END $$;

-- ============================================
-- Notification Event Types (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'notification_event_type') THEN
        INSERT INTO notification_event_type (name, valid_id, create_by, change_by) 
        SELECT 'Ticket', 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM notification_event_type WHERE name = 'Ticket');
        
        INSERT INTO notification_event_type (name, valid_id, create_by, change_by) 
        SELECT 'Article', 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM notification_event_type WHERE name = 'Article');
        
        INSERT INTO notification_event_type (name, valid_id, create_by, change_by) 
        SELECT 'Appointment', 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM notification_event_type WHERE name = 'Appointment');
    END IF;
END $$;

-- ============================================
-- Web Service Configuration (basic example)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gi_webservice_config') THEN
        INSERT INTO gi_webservice_config (name, config, config_md5, valid_id, create_by, change_by) 
        SELECT 'Example Web Service', '{}', MD5('{}'), 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM gi_webservice_config WHERE name = 'Example Web Service');
    END IF;
END $$;

-- ============================================
-- Calendar Configuration (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'calendar') THEN
        INSERT INTO calendar (name, group_id, salt_string, color, valid_id, create_by, change_by) 
        SELECT 'Default', 1, MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT), '#3A87AD', 1, 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM calendar WHERE name = 'Default');
    END IF;
END $$;

-- ============================================
-- System Data - Core System Settings
-- ============================================
INSERT INTO system_data (data_key, data_value, create_by, change_by) 
SELECT 'SystemID', '10', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM system_data WHERE data_key = 'SystemID');

INSERT INTO system_data (data_key, data_value, create_by, change_by) 
SELECT 'FQDN', 'gotrs.local', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM system_data WHERE data_key = 'FQDN');

INSERT INTO system_data (data_key, data_value, create_by, change_by) 
SELECT 'Organization', 'Your Organization', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM system_data WHERE data_key = 'Organization');

INSERT INTO system_data (data_key, data_value, create_by, change_by) 
SELECT 'AdminEmail', 'admin@gotrs.local', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM system_data WHERE data_key = 'AdminEmail');

INSERT INTO system_data (data_key, data_value, create_by, change_by) 
SELECT 'NotificationSenderName', 'GOTRS Notification System', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM system_data WHERE data_key = 'NotificationSenderName');

INSERT INTO system_data (data_key, data_value, create_by, change_by) 
SELECT 'NotificationSenderEmail', 'noreply@gotrs.local', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM system_data WHERE data_key = 'NotificationSenderEmail');

-- ============================================
-- Process Management Configuration (if tables exist)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'pm_entity_sync') THEN
        INSERT INTO pm_entity_sync (entity_type, entity_id, sync_state) 
        SELECT 'Process', '1', 'not_sync'
        WHERE NOT EXISTS (SELECT 1 FROM pm_entity_sync WHERE entity_type = 'Process');
    END IF;
END $$;

-- ============================================
-- ACL Configuration (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'acl') THEN
        -- Default ACL entries would go here
        -- But ACLs are usually complex and site-specific
        NULL;
    END IF;
END $$;

-- ============================================
-- Generic Agent Jobs (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'generic_agent_jobs') THEN
        -- Example generic agent job for auto-closing tickets
        INSERT INTO generic_agent_jobs (job_name, job_data, create_by, change_by) 
        SELECT 'close-tickets-after-7-days', '{"Description": "Close tickets in pending auto close state after 7 days", "TicketSearch": {"StateType": ["pending auto close"]}, "TicketChange": {"State": "closed successful"}}', 1, 1
        WHERE NOT EXISTS (SELECT 1 FROM generic_agent_jobs WHERE job_name = 'close-tickets-after-7-days');
    END IF;
END $$;

-- ============================================
-- Mail Account Configuration (if table exists)
-- ============================================
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'mail_account') THEN
        -- Note: Mail accounts are usually configured per-installation
        -- This is just a placeholder example
        NULL;
    END IF;
END $$;