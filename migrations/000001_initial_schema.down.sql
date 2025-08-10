-- Drop all triggers first
DROP TRIGGER IF EXISTS trigger_system_config_change_time ON system_config;
DROP TRIGGER IF EXISTS trigger_articles_change_time ON articles;
DROP TRIGGER IF EXISTS trigger_tickets_change_time ON tickets;
DROP TRIGGER IF EXISTS trigger_queues_change_time ON queues;
DROP TRIGGER IF EXISTS trigger_roles_change_time ON roles;
DROP TRIGGER IF EXISTS trigger_groups_change_time ON groups;
DROP TRIGGER IF EXISTS trigger_users_change_time ON users;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_change_time();

-- Drop indexes
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP INDEX IF EXISTS idx_user_groups_group_id;
DROP INDEX IF EXISTS idx_user_groups_user_id;
DROP INDEX IF EXISTS idx_sessions_last_activity;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_users_valid_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_login;
DROP INDEX IF EXISTS idx_article_attachments_article_id;
DROP INDEX IF EXISTS idx_articles_type_id;
DROP INDEX IF EXISTS idx_articles_create_time;
DROP INDEX IF EXISTS idx_articles_ticket_id;
DROP INDEX IF EXISTS idx_tickets_archive_flag;
DROP INDEX IF EXISTS idx_tickets_create_time;
DROP INDEX IF EXISTS idx_tickets_customer_id;
DROP INDEX IF EXISTS idx_tickets_user_id;
DROP INDEX IF EXISTS idx_tickets_priority_id;
DROP INDEX IF EXISTS idx_tickets_state_id;
DROP INDEX IF EXISTS idx_tickets_queue_id;
DROP INDEX IF EXISTS idx_tickets_tn;

-- Drop tables in reverse order (considering foreign key constraints)
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS system_config;
DROP TABLE IF EXISTS role_groups;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS user_groups;
DROP TABLE IF EXISTS article_attachments;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS ticket_priorities;
DROP TABLE IF EXISTS ticket_states;
DROP TABLE IF EXISTS queues;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;