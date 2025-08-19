-- Rollback generated test data
-- This file is auto-generated and gitignored

-- Remove test articles
DELETE FROM article_data_mime WHERE article_id IN (
    SELECT a.id FROM article a 
    JOIN ticket t ON a.ticket_id = t.id 
    WHERE t.tn LIKE '202501%'
);

DELETE FROM article WHERE ticket_id IN (
    SELECT id FROM ticket WHERE tn LIKE '202501%'
);

-- Remove test tickets
DELETE FROM ticket WHERE tn LIKE '202501%';

-- Remove test user groups
DELETE FROM user_groups WHERE user_id IN (
    SELECT id FROM users WHERE login IN ('admin', 'agent.smith', 'agent.jones')
);

-- Remove test users
DELETE FROM users WHERE login IN ('admin', 'agent.smith', 'agent.jones');

-- Remove test customers
DELETE FROM customer_user WHERE login IN ('john.customer', 'jane.customer', 'bob.customer');

-- Remove test companies
DELETE FROM customer_company WHERE customer_id IN ('COMP1', 'COMP2', 'COMP3');