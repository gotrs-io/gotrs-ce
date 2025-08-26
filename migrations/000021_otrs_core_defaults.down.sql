-- ============================================
-- Rollback OTRS Core Default Data
-- ============================================

-- Remove default queue auto responses
DELETE FROM queue_auto_response WHERE queue_id = 1 AND auto_response_type_id IN (1, 2);

-- Remove default auto responses
DELETE FROM auto_response WHERE id <= 2;

-- Remove default auto response types
DELETE FROM auto_response_type WHERE id <= 5;

-- Remove default communication channels
DELETE FROM communication_channel WHERE id <= 5;

-- Remove default services (but keep any user-created ones)
DELETE FROM service WHERE name IN ('IT Support', 'IT Support::Hardware', 'IT Support::Software', 'IT Support::Network')
AND id <= 10;  -- Only delete if they're in the default ID range

-- Remove default SLAs (but keep any user-created ones)
DELETE FROM sla WHERE name IN ('Standard', 'Premium', 'Gold', 'Critical')
AND id <= 10;  -- Only delete if they're in the default ID range

-- Remove default roles (optional - these might be critical)
-- DELETE FROM roles WHERE name IN ('Agent', 'Admin', 'Customer');

-- Note: We don't remove article_type data as it might be referenced