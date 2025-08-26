-- ============================================
-- Rollback Lookup Table Defaults
-- ============================================

-- Remove default salutations (keep any user-created ones with id > 4)
DELETE FROM salutation WHERE id <= 4;

-- Remove default signatures (keep any user-created ones with id > 4)
DELETE FROM signature WHERE id <= 4;

-- Remove default system addresses (keep any user-created ones with id > 6)
DELETE FROM system_address WHERE id <= 6;

-- Note: We don't remove valid statuses as they are core to the system functioning