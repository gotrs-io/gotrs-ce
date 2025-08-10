-- Initial seed data for GOTRS

-- Insert default system user (for system operations)
INSERT INTO users (login, password_hash, first_name, last_name, email, create_by, change_by) 
VALUES ('system', '$2a$12$dummy.hash.for.system.user', 'System', 'User', 'system@localhost', 1, 1);

-- Insert default admin user (password: admin123)
INSERT INTO users (login, password_hash, first_name, last_name, email, create_by, change_by) 
-- semgrep:ignore generic.secrets.security.detected-bcrypt-hash.detected-bcrypt-hash: This is a legitimate test bcrypt hash for development
VALUES ('admin', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/lewBUz8n8p9UqRVe2', 'Admin', 'User', 'admin@localhost', 1, 1);

-- Insert default groups
INSERT INTO groups (name, comment, create_by, change_by) VALUES 
('admin', 'Administrator group', 1, 1),
('users', 'Regular users group', 1, 1),
('customer', 'Customer group', 1, 1),
('stats', 'Statistics group', 1, 1);

-- Insert default roles
INSERT INTO roles (name, comment, create_by, change_by) VALUES 
('admin', 'Administrator role with full access', 1, 1),
('agent', 'Agent role for ticket handling', 1, 1),
('customer', 'Customer role for ticket creation', 1, 1);

-- Insert default queues
INSERT INTO queues (name, group_id, comment, create_by, change_by) VALUES 
('Raw', 1, 'All new tickets are placed in this queue by default', 1, 1),
('Junk', 1, 'Spam and junk emails', 1, 1),
('Misc', 2, 'Miscellaneous tickets', 1, 1),
('Support', 2, 'General support requests', 1, 1);

-- Insert default ticket states
INSERT INTO ticket_states (name, type_id, create_by, change_by) VALUES 
('new', 1, 1, 1),
('open', 2, 1, 1),
('pending reminder', 5, 1, 1),
('pending auto', 5, 1, 1),
('pending auto close+', 5, 1, 1),
('pending auto close-', 5, 1, 1),
('closed successful', 3, 1, 1),
('closed unsuccessful', 3, 1, 1),
('merged', 4, 1, 1),
('removed', 4, 1, 1);

-- Insert default ticket priorities
INSERT INTO ticket_priorities (name, create_by, change_by) VALUES 
('1 very low', 1, 1),
('2 low', 1, 1),
('3 normal', 1, 1),
('4 high', 1, 1),
('5 very high', 1, 1);

-- Assign admin user to admin group with full permissions
INSERT INTO user_groups (user_id, group_id, permission_key, permission_value, create_by, change_by) VALUES 
(2, 1, 'ro', 1, 1, 1),
(2, 1, 'move_into', 1, 1, 1),
(2, 1, 'create', 1, 1, 1),
(2, 1, 'note', 1, 1, 1),
(2, 1, 'owner', 1, 1, 1),
(2, 1, 'priority', 1, 1, 1),
(2, 1, 'rw', 1, 1, 1);

-- Assign admin user to admin role
INSERT INTO user_roles (user_id, role_id, create_by, change_by) VALUES 
(2, 1, 1, 1);

-- Grant admin role access to all groups
INSERT INTO role_groups (role_id, group_id, permission_key, permission_value, create_by, change_by) VALUES 
(1, 1, 'ro', 1, 1, 1),
(1, 1, 'move_into', 1, 1, 1),
(1, 1, 'create', 1, 1, 1),
(1, 1, 'note', 1, 1, 1),
(1, 1, 'owner', 1, 1, 1),
(1, 1, 'priority', 1, 1, 1),
(1, 1, 'rw', 1, 1, 1),
(1, 2, 'ro', 1, 1, 1),
(1, 2, 'move_into', 1, 1, 1),
(1, 2, 'create', 1, 1, 1),
(1, 2, 'note', 1, 1, 1),
(1, 2, 'owner', 1, 1, 1),
(1, 2, 'priority', 1, 1, 1),
(1, 2, 'rw', 1, 1, 1);

-- Insert basic system configuration
INSERT INTO system_config (name, value, description, create_by, change_by) VALUES 
('SystemID', '10', 'Unique system ID for ticket numbers', 1, 1),
('FQDN', 'localhost', 'Fully qualified domain name', 1, 1),
('HttpType', 'http', 'HTTP protocol type', 1, 1),
('ScriptAlias', 'gotrs/', 'Script alias for web interface', 1, 1),
('ProductName', 'GOTRS', 'Product name displayed in interface', 1, 1),
('AdminEmail', 'admin@localhost', 'Administrator email address', 1, 1),
('Organization', 'GOTRS Organization', 'Organization name', 1, 1),
('TicketNumberGenerator', 'DateTimeRandom', 'Ticket number generation method', 1, 1),
('TicketHook', 'Ticket#', 'Ticket hook prefix for ticket numbers', 1, 1),
('TicketHookDivider', '', 'Ticket hook divider', 1, 1),
('DefaultLanguage', 'en', 'Default system language', 1, 1),
('DefaultCharset', 'utf-8', 'Default character set', 1, 1),
('TimeZone', 'UTC', 'System timezone', 1, 1),
('SendmailModule', 'SMTP', 'Email sending module', 1, 1),
('CheckEmailAddresses', '1', 'Check email address validity', 1, 1);