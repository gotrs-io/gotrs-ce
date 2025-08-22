-- Rollback OTRS Test Data
-- Remove all test data added in the up migration

-- Remove test permissions (user_groups)
DELETE FROM user_groups WHERE user_id IN (2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20);
DELETE FROM user_groups WHERE group_id IN (2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18);

-- Remove test group roles
DELETE FROM group_role WHERE group_id IN (2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18);

-- Remove test role permissions
DELETE FROM role_permission WHERE role_id IN (2, 3, 4, 5, 6, 7, 8, 9, 10);

-- Remove test roles
DELETE FROM role WHERE id IN (2, 3, 4, 5, 6, 7, 8, 9, 10);

-- Remove test groups
DELETE FROM groups WHERE id IN (2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18);

-- Remove test customer companies
DELETE FROM customer_company WHERE customer_id LIKE 'test_%' OR customer_id LIKE 'acme_%' OR customer_id LIKE 'widgets_%';

-- Remove test users
DELETE FROM users WHERE id IN (2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20);

-- Reset sequences to safe values
SELECT setval('users_id_seq', 1, true);
SELECT setval('groups_id_seq', 1, true);
SELECT setval('role_id_seq', 1, true);