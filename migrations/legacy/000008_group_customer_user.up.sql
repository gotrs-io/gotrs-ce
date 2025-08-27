-- Customer User to Group mapping table (OTRS compatible)
-- This feature allows customer users to be assigned to groups
-- Must be enabled in system configuration to use

CREATE TABLE IF NOT EXISTS group_customer_user (
    user_id VARCHAR(100) NOT NULL,      -- References customer_user.login
    group_id INTEGER NOT NULL REFERENCES groups(id),
    permission_key VARCHAR(20) NOT NULL,
    permission_value SMALLINT NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by INTEGER NOT NULL,
    change_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    change_by INTEGER NOT NULL,
    PRIMARY KEY(user_id, group_id, permission_key)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_group_customer_user_group_id ON group_customer_user(group_id);
CREATE INDEX IF NOT EXISTS idx_group_customer_user_user_id ON group_customer_user(user_id);

-- Add some test data (optional, can be commented out for production)
-- Assign customer user 'hans' to group 'users' with read-write permission
INSERT INTO group_customer_user (user_id, group_id, permission_key, permission_value, create_by, change_by)
SELECT 'hans', 2, 'rw', 1, 1, 1
WHERE EXISTS (SELECT 1 FROM customer_user WHERE login = 'hans')
  AND EXISTS (SELECT 1 FROM groups WHERE id = 2)
  AND NOT EXISTS (SELECT 1 FROM group_customer_user WHERE user_id = 'hans' AND group_id = 2 AND permission_key = 'rw');