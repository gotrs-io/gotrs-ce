-- Rollback Performance Indexes
-- Remove indexes added in the up migration

DROP INDEX IF EXISTS idx_ticket_queue_state;
DROP INDEX IF EXISTS idx_ticket_customer_user;
DROP INDEX IF EXISTS idx_ticket_priority;
DROP INDEX IF EXISTS idx_ticket_create_time;
DROP INDEX IF EXISTS idx_article_ticket_id;
DROP INDEX IF EXISTS idx_article_create_time;
DROP INDEX IF EXISTS idx_users_login;
DROP INDEX IF EXISTS idx_users_valid_id;
DROP INDEX IF EXISTS idx_groups_name;
DROP INDEX IF EXISTS idx_groups_valid_id;
DROP INDEX IF EXISTS idx_queue_name;
DROP INDEX IF EXISTS idx_queue_valid_id;
DROP INDEX IF EXISTS idx_priority_name;
DROP INDEX IF EXISTS idx_ticket_priority_name;
DROP INDEX IF EXISTS idx_ticket_state_name;
DROP INDEX IF EXISTS idx_ticket_type_name;
DROP INDEX IF EXISTS idx_customer_user_login;
DROP INDEX IF EXISTS idx_customer_user_email;
DROP INDEX IF EXISTS idx_customer_company_id;