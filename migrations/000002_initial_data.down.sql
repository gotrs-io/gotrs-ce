-- Remove initial seed data

-- Remove system configuration
DELETE FROM system_config WHERE create_by IN (1, 2);

-- Remove role-group relationships
DELETE FROM role_groups WHERE create_by IN (1, 2);

-- Remove user-role relationships
DELETE FROM user_roles WHERE create_by IN (1, 2);

-- Remove user-group relationships
DELETE FROM user_groups WHERE create_by IN (1, 2);

-- Remove ticket priorities
DELETE FROM ticket_priorities WHERE create_by IN (1, 2);

-- Remove ticket states
DELETE FROM ticket_states WHERE create_by IN (1, 2);

-- Remove queues
DELETE FROM queues WHERE create_by IN (1, 2);

-- Remove roles
DELETE FROM roles WHERE create_by IN (1, 2);

-- Remove groups
DELETE FROM groups WHERE create_by IN (1, 2);

-- Remove users (admin and system)
DELETE FROM users WHERE create_by IN (1, 2) OR id IN (1, 2);