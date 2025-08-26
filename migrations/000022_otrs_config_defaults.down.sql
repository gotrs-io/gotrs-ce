-- ============================================
-- Rollback OTRS Configuration Defaults
-- ============================================

-- Remove system data entries
DELETE FROM system_data WHERE data_key IN (
    'SystemID', 'FQDN', 'Organization', 'AdminEmail', 
    'NotificationSenderName', 'NotificationSenderEmail'
);

-- Conditional removals for tables that might exist
DO $$ 
BEGIN
    -- Remove time accounting actions
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'time_accounting_action') THEN
        DELETE FROM time_accounting_action WHERE name IN (
            'Phone Call', 'Email', 'Research', 'Meeting', 
            'Documentation', 'Testing', 'Development', 'Other'
        );
    END IF;
    
    -- Remove process states
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'pm_process_state') THEN
        DELETE FROM pm_process_state WHERE name IN ('Active', 'Inactive', 'FadeAway');
    END IF;
    
    -- Remove standard templates
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'standard_template') THEN
        DELETE FROM standard_template WHERE name IN (
            'empty answer', 'Thank you for your email', 'Your request has been closed'
        );
    END IF;
    
    -- Remove notification event types
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'notification_event_type') THEN
        DELETE FROM notification_event_type WHERE name IN ('Ticket', 'Article', 'Appointment');
    END IF;
    
    -- Remove web service config
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gi_webservice_config') THEN
        DELETE FROM gi_webservice_config WHERE name = 'Example Web Service';
    END IF;
    
    -- Remove default calendar
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'calendar') THEN
        DELETE FROM calendar WHERE name = 'Default';
    END IF;
    
    -- Remove process sync
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'pm_entity_sync') THEN
        DELETE FROM pm_entity_sync WHERE entity_type = 'Process' AND entity_id = '1';
    END IF;
    
    -- Remove generic agent job
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'generic_agent_jobs') THEN
        DELETE FROM generic_agent_jobs WHERE job_name = 'close-tickets-after-7-days';
    END IF;
END $$;