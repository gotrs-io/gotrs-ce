-- ============================================
-- System Configuration Defaults
-- Populate sysconfig_default table with GoTRS configuration options
-- Based on config/Config.yaml structure
-- ============================================

-- Clear existing configuration (for development - remove in production)
DELETE FROM sysconfig_default WHERE id > 0;
ALTER SEQUENCE sysconfig_default_id_seq RESTART WITH 1;

-- ============================================
-- Core System Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required, 
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value, 
    create_by, change_by
) VALUES
-- SystemID
('SystemID', 'The system ID used for ticket number generation (1-99)', 'Core::System', 
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"integer","default":10,"min":1,"max":99,"validation":"^[1-9][0-9]?$"}',
 '{"type":"integer","default":10,"min":1,"max":99}',
 'Config.yaml', '10', 1, 1),

-- FQDN
('FQDN', 'Fully Qualified Domain Name of the system', 'Core::System',
 0, 0, 1, 1, 100, 1, 0,
 '{"type":"string","default":"gotrs.local","validation":"^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](?:\\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$"}',
 '{"type":"string","default":"gotrs.local"}',
 'Config.yaml', 'gotrs.local', 1, 1),

-- Organization
('Organization', 'Organization name for email signatures and UI', 'Core::System',
 0, 0, 1, 1, 100, 1, 0,
 '{"type":"string","default":"Your Organization"}',
 '{"type":"string","default":"Your Organization"}',
 'Config.yaml', 'Your Organization', 1, 1),

-- AdminEmail
('AdminEmail', 'Email address of the system administrator', 'Core::System',
 0, 0, 1, 1, 100, 1, 0,
 '{"type":"email","default":"admin@gotrs.local","validation":"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"}',
 '{"type":"email","default":"admin@gotrs.local"}',
 'Config.yaml', 'admin@gotrs.local', 1, 1);

-- ============================================
-- Ticket Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- Ticket::NumberGenerator
('Ticket::NumberGenerator', 'Ticket number generator module', 'Core::Ticket',
 0, 0, 1, 1, 200, 0, 0,
 '{"type":"select","default":"Date","options":[{"value":"Date","label":"Date-based (YYYYMMDDhhmmss)"},{"value":"DateChecksum","label":"Date with checksum"},{"value":"Random","label":"Random number"},{"value":"Increment","label":"Incremental"}]}',
 '{"type":"select","default":"Date"}',
 'Config.yaml', 'Date', 1, 1),

-- Ticket::DefaultPriority
('Ticket::DefaultPriority', 'Default priority for new tickets', 'Core::Ticket',
 0, 0, 1, 1, 200, 1, 0,
 '{"type":"select","default":"3 normal","options":[{"value":"1 very low"},{"value":"2 low"},{"value":"3 normal"},{"value":"4 high"},{"value":"5 very high"}]}',
 '{"type":"select","default":"3 normal"}',
 'Config.yaml', '3 normal', 1, 1),

-- Ticket::DefaultQueue
('Ticket::DefaultQueue', 'Default queue for new tickets', 'Core::Ticket',
 0, 0, 1, 1, 200, 1, 0,
 '{"type":"string","default":"Raw"}',
 '{"type":"string","default":"Raw"}',
 'Config.yaml', 'Raw', 1, 1),

-- Ticket::ViewableStateType
('Ticket::ViewableStateType', 'State types that are viewable in customer interface', 'Core::Ticket',
 0, 0, 1, 1, 200, 0, 0,
 '{"type":"array","default":["new","open","pending reminder","pending auto"]}',
 '{"type":"array","default":["new","open","pending reminder","pending auto"]}',
 'Config.yaml', '["new","open","pending reminder","pending auto"]', 1, 1);

-- ============================================
-- Email Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- SendmailModule
('SendmailModule', 'Email sending backend module', 'Core::Email',
 0, 0, 1, 1, 200, 0, 0,
 '{"type":"select","default":"SMTP","options":[{"value":"SMTP","label":"SMTP"},{"value":"Sendmail","label":"Sendmail binary"},{"value":"Test","label":"Test (don''t send)"}]}',
 '{"type":"select","default":"SMTP"}',
 'Config.yaml', 'SMTP', 1, 1),

-- SendmailModule::Host
('SendmailModule::Host', 'SMTP server hostname', 'Core::Email::SMTP',
 0, 0, 0, 1, 200, 1, 0,
 '{"type":"string","default":"localhost","depends_on":"SendmailModule","depends_value":"SMTP"}',
 '{"type":"string","default":"localhost"}',
 'Config.yaml', 'localhost', 1, 1),

-- SendmailModule::Port
('SendmailModule::Port', 'SMTP server port', 'Core::Email::SMTP',
 0, 0, 0, 1, 200, 1, 0,
 '{"type":"integer","default":25,"min":1,"max":65535,"depends_on":"SendmailModule","depends_value":"SMTP"}',
 '{"type":"integer","default":25}',
 'Config.yaml', '25', 1, 1),

-- NotificationSenderName
('NotificationSenderName', 'Name used for system notification emails', 'Core::Email',
 0, 0, 1, 1, 100, 1, 0,
 '{"type":"string","default":"GoTRS Notification System"}',
 '{"type":"string","default":"GoTRS Notification System"}',
 'Config.yaml', 'GoTRS Notification System', 1, 1),

-- NotificationSenderEmail  
('NotificationSenderEmail', 'Email address used for system notifications', 'Core::Email',
 0, 0, 1, 1, 100, 1, 0,
 '{"type":"email","default":"noreply@gotrs.local"}',
 '{"type":"email","default":"noreply@gotrs.local"}',
 'Config.yaml', 'noreply@gotrs.local', 1, 1);

-- ============================================
-- Session Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- SessionMaxTime
('SessionMaxTime', 'Max session time in seconds (default 4 hours)', 'Core::Session',
 0, 0, 1, 1, 100, 1, 0,
 '{"type":"integer","default":14400,"min":600,"max":172800}',
 '{"type":"integer","default":14400}',
 'Config.yaml', '14400', 1, 1),

-- SessionMaxIdleTime
('SessionMaxIdleTime', 'Max idle time in seconds (default 2 hours)', 'Core::Session',
 0, 0, 1, 1, 100, 1, 0,
 '{"type":"integer","default":7200,"min":300,"max":86400}',
 '{"type":"integer","default":7200}',
 'Config.yaml', '7200', 1, 1),

-- SessionUseCookie
('SessionUseCookie', 'Use cookies for session management', 'Core::Session',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"boolean","default":true}',
 '{"type":"boolean","default":true}',
 'Config.yaml', 'true', 1, 1);

-- ============================================
-- Frontend Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- Frontend::WebPath
('Frontend::WebPath', 'Web path for static assets', 'Frontend::Base',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"string","default":"/"}',
 '{"type":"string","default":"/"}',
 'Config.yaml', '/', 1, 1),

-- Frontend::DefaultTheme
('Frontend::DefaultTheme', 'Default UI theme', 'Frontend::Base',
 0, 0, 1, 1, 200, 1, 0,
 '{"type":"select","default":"Standard","options":[{"value":"Standard","label":"Standard Theme"},{"value":"Dark","label":"Dark Theme"},{"value":"High Contrast","label":"High Contrast"}]}',
 '{"type":"select","default":"Standard"}',
 'Config.yaml', 'Standard', 1, 1),

-- Frontend::RichText
('Frontend::RichText', 'Enable rich text editor', 'Frontend::Base',
 0, 0, 1, 1, 200, 1, 0,
 '{"type":"boolean","default":true}',
 '{"type":"boolean","default":true}',
 'Config.yaml', 'true', 1, 1),

-- Frontend::AvatarEngine
('Frontend::AvatarEngine', 'Avatar generation service', 'Frontend::Base',
 0, 0, 1, 1, 300, 1, 0,
 '{"type":"select","default":"Gravatar","options":[{"value":"None","label":"No avatars"},{"value":"Gravatar","label":"Gravatar"},{"value":"Local","label":"Local storage"}]}',
 '{"type":"select","default":"Gravatar"}',
 'Config.yaml', 'Gravatar', 1, 1);

-- ============================================
-- Agent Interface Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- Agent::DefaultTicketView
('Agent::DefaultTicketView', 'Default ticket view for agents', 'Frontend::Agent',
 0, 0, 1, 1, 300, 1, 0,
 '{"type":"select","default":"Small","options":[{"value":"Small","label":"Small (25 per page)"},{"value":"Medium","label":"Medium (50 per page)"},{"value":"Large","label":"Large (100 per page)"},{"value":"Preview","label":"Preview mode"}]}',
 '{"type":"select","default":"Small"}',
 'Config.yaml', 'Small', 1, 1),

-- Agent::TicketViewRefreshTime
('Agent::TicketViewRefreshTime', 'Auto-refresh interval for ticket views (seconds, 0 = disabled)', 'Frontend::Agent',
 0, 0, 1, 1, 300, 1, 0,
 '{"type":"integer","default":0,"min":0,"max":86400}',
 '{"type":"integer","default":0}',
 'Config.yaml', '0', 1, 1);

-- ============================================
-- Customer Interface Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- Customer::DefaultTicketView
('Customer::DefaultTicketView', 'Default ticket view for customers', 'Frontend::Customer',
 0, 0, 1, 1, 300, 1, 0,
 '{"type":"select","default":"Small","options":[{"value":"Small","label":"Small (10 per page)"},{"value":"Medium","label":"Medium (20 per page)"},{"value":"Large","label":"Large (50 per page)"}]}',
 '{"type":"select","default":"Small"}',
 'Config.yaml', 'Small', 1, 1),

-- Customer::TicketViewableStatus
('Customer::TicketViewableStatus', 'Ticket statuses visible to customers', 'Frontend::Customer',
 0, 0, 1, 1, 200, 0, 0,
 '{"type":"array","default":["new","open","closed successful","closed unsuccessful"]}',
 '{"type":"array","default":["new","open","closed successful","closed unsuccessful"]}',
 'Config.yaml', '["new","open","closed successful","closed unsuccessful"]', 1, 1);

-- ============================================
-- Performance Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- Cache::Module
('Cache::Module', 'Cache backend module', 'Core::Cache',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"select","default":"Redis","options":[{"value":"Memory","label":"In-memory (not for production)"},{"value":"Redis","label":"Redis/Valkey"},{"value":"File","label":"File-based"}]}',
 '{"type":"select","default":"Redis"}',
 'Config.yaml', 'Redis', 1, 1),

-- Cache::TTL
('Cache::TTL', 'Default cache TTL in seconds', 'Core::Cache',
 0, 0, 1, 1, 200, 0, 0,
 '{"type":"integer","default":3600,"min":60,"max":86400}',
 '{"type":"integer","default":3600}',
 'Config.yaml', '3600', 1, 1);

-- ============================================
-- Database Settings (readonly)
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- DatabaseHost
('DatabaseHost', 'Database server hostname', 'Core::Database',
 0, 1, 1, 1, 50, 0, 0,
 '{"type":"string","default":"localhost"}',
 '{"type":"string","default":"localhost"}',
 'Config.yaml', 'localhost', 1, 1),

-- DatabaseName
('DatabaseName', 'Database name', 'Core::Database',
 0, 1, 1, 1, 50, 0, 0,
 '{"type":"string","default":"gotrs"}',
 '{"type":"string","default":"gotrs"}',
 'Config.yaml', 'gotrs', 1, 1);

-- ============================================
-- Log Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- LogModule
('LogModule', 'Logging backend module', 'Core::Log',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"select","default":"File","options":[{"value":"File","label":"Log to file"},{"value":"Syslog","label":"System logger"},{"value":"Console","label":"Console output"}]}',
 '{"type":"select","default":"File"}',
 'Config.yaml', 'File', 1, 1),

-- LogLevel
('LogLevel', 'Minimum log level', 'Core::Log',
 0, 0, 1, 1, 100, 1, 0,
 '{"type":"select","default":"info","options":[{"value":"debug","label":"Debug"},{"value":"info","label":"Info"},{"value":"warn","label":"Warning"},{"value":"error","label":"Error"}]}',
 '{"type":"select","default":"info"}',
 'Config.yaml', 'info', 1, 1);

-- ============================================
-- Security Settings
-- ============================================

INSERT INTO sysconfig_default (
    name, description, navigation, is_invisible, is_readonly, is_required,
    is_valid, has_configlevel, user_modification_possible, user_modification_active,
    xml_content_raw, xml_content_parsed, xml_filename, effective_value,
    create_by, change_by
) VALUES
-- SecureMode
('SecureMode', 'Enable secure mode (HTTPS only)', 'Core::Security',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"boolean","default":false}',
 '{"type":"boolean","default":false}',
 'Config.yaml', 'false', 1, 1),

-- PasswordMinLength
('PasswordMinLength', 'Minimum password length', 'Core::Security',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"integer","default":8,"min":6,"max":128}',
 '{"type":"integer","default":8}',
 'Config.yaml', '8', 1, 1),

-- PasswordRequireUppercase
('PasswordRequireUppercase', 'Require uppercase letters in passwords', 'Core::Security',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"boolean","default":true}',
 '{"type":"boolean","default":true}',
 'Config.yaml', 'true', 1, 1),

-- PasswordRequireLowercase
('PasswordRequireLowercase', 'Require lowercase letters in passwords', 'Core::Security',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"boolean","default":true}',
 '{"type":"boolean","default":true}',
 'Config.yaml', 'true', 1, 1),

-- PasswordRequireDigit
('PasswordRequireDigit', 'Require digits in passwords', 'Core::Security',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"boolean","default":true}',
 '{"type":"boolean","default":true}',
 'Config.yaml', 'true', 1, 1),

-- PasswordRequireSpecial
('PasswordRequireSpecial', 'Require special characters in passwords', 'Core::Security',
 0, 0, 1, 1, 100, 0, 0,
 '{"type":"boolean","default":false}',
 '{"type":"boolean","default":false}',
 'Config.yaml', 'false', 1, 1);

-- Reset sequence
SELECT setval('sysconfig_default_id_seq', (SELECT COALESCE(MAX(id), 1) FROM sysconfig_default), true);