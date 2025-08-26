-- ============================================
-- Lookup Table Defaults
-- Default data for salutations, signatures, system addresses, and other lookups
-- Based on OTRS standard configuration
-- ============================================

-- ============================================
-- Salutations
-- ============================================
-- Clear existing test data first
DELETE FROM salutation WHERE id > 0;

-- Reset sequence
ALTER SEQUENCE salutation_id_seq RESTART WITH 1;

-- Insert OTRS-compatible salutations
INSERT INTO salutation (name, text, content_type, comments, valid_id, create_time, create_by, change_time, change_by) VALUES
('system standard salutation (en)', 'Dear <OTRS_UserFirstname>,

Thank you for your inquiry.', 'text/plain; charset=utf-8', 'Standard salutation for English communication', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('Formal (Mr/Ms)', 'Dear Mr./Ms. <OTRS_UserLastname>,

Thank you for contacting us.', 'text/plain; charset=utf-8', 'Formal salutation using last name', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('Informal', 'Hi <OTRS_UserFirstname>,

Thanks for reaching out!', 'text/plain; charset=utf-8', 'Informal salutation for casual communication', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('No salutation', '', 'text/plain; charset=utf-8', 'Empty salutation for automated messages', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1);

-- ============================================
-- Signatures
-- ============================================
-- Clear existing test data first
DELETE FROM signature WHERE id > 0;

-- Reset sequence
ALTER SEQUENCE signature_id_seq RESTART WITH 1;

-- Insert OTRS-compatible signatures
INSERT INTO signature (name, text, content_type, comments, valid_id, create_time, create_by, change_time, change_by) VALUES
('system standard signature (en)', 'Your Ticket-Team

<OTRS_Agent_UserFirstname> <OTRS_Agent_UserLastname>

--
Super Support - Waterford Business Park
5201 Blue Lagoon Drive - 8th Floor & 9th Floor - Miami, 33126 USA
Email: hot@example.com - Web: http://www.example.com/', 'text/plain; charset=utf-8', 'Standard signature for English communication', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('Support Team', 'Best regards,
Support Team

--
<OTRS_Config_Organization>
<OTRS_Config_CustomerHeadline>', 'text/plain; charset=utf-8', 'Generic support team signature', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('Technical Support', 'Kind regards,
Technical Support Department

<OTRS_Agent_UserFirstname> <OTRS_Agent_UserLastname>
Phone: <OTRS_Agent_UserPhone>

--
<OTRS_Config_Organization>
Technical Support Division', 'text/plain; charset=utf-8', 'Technical support team signature', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('No signature', '', 'text/plain; charset=utf-8', 'Empty signature for automated messages', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1);

-- ============================================
-- System Email Addresses
-- ============================================
-- Clear existing test data first
DELETE FROM system_address WHERE id > 0;

-- Reset sequence
ALTER SEQUENCE system_address_id_seq RESTART WITH 1;

-- Insert OTRS-compatible system addresses
-- Note: queue_id is required, using 1 as default (Raw queue typically)
INSERT INTO system_address (value0, value1, value2, value3, comments, valid_id, queue_id, create_time, create_by, change_time, change_by) VALUES
('otrs@localhost', 'otrs@localhost', 'OTRS System', '', 'Default system email address', 1, 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('support@example.com', 'support@example.com', 'Support Team', '', 'General support email address', 1, 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('noreply@example.com', 'noreply@example.com', 'No Reply', '', 'No-reply address for automated messages', 1, 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('helpdesk@example.com', 'helpdesk@example.com', 'Help Desk', '', 'Help desk email address', 1, 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('sales@example.com', 'sales@example.com', 'Sales Department', '', 'Sales department email', 1, 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1),
('billing@example.com', 'billing@example.com', 'Billing Department', '', 'Billing department email', 1, 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1);

-- ============================================
-- Valid Status (if not already populated)
-- ============================================
-- This is usually populated by the schema, but ensure it exists
INSERT INTO valid (id, name, create_time, create_by, change_time, change_by) 
SELECT 1, 'valid', CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1
WHERE NOT EXISTS (SELECT 1 FROM valid WHERE id = 1);

INSERT INTO valid (id, name, create_time, create_by, change_time, change_by) 
SELECT 2, 'invalid', CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1
WHERE NOT EXISTS (SELECT 1 FROM valid WHERE id = 2);

INSERT INTO valid (id, name, create_time, create_by, change_time, change_by) 
SELECT 3, 'invalid-temporarily', CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP, 1
WHERE NOT EXISTS (SELECT 1 FROM valid WHERE id = 3);

-- ============================================
-- Follow-up Options (if table exists)
-- These are typically hardcoded in OTRS but we'll add them if the table exists
-- ============================================
-- Note: OTRS typically uses hardcoded values 1=possible, 2=reject, 3=new ticket
-- These are referenced by ID in the queue configuration

-- ============================================
-- Update sequences to ensure next values are correct
-- ============================================
SELECT setval('salutation_id_seq', (SELECT MAX(id) FROM salutation), true);
SELECT setval('signature_id_seq', (SELECT MAX(id) FROM signature), true);
SELECT setval('system_address_id_seq', (SELECT MAX(id) FROM system_address), true);

-- ============================================
-- Link default items to existing queues
-- ============================================
-- Set default salutation and signature for existing queues that don't have them
UPDATE queue 
SET salutation_id = 1, signature_id = 1
WHERE salutation_id IS NULL OR signature_id IS NULL;

-- Set default system address for existing queues that don't have one
UPDATE queue 
SET system_address_id = 1
WHERE system_address_id IS NULL;