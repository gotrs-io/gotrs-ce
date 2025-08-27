-- Remove Configuration queue if we need to rollback
-- Note: This will fail if there are tickets in the Configuration queue
DELETE FROM queue WHERE name = 'Configuration' AND NOT EXISTS (
    SELECT 1 FROM ticket WHERE queue_id = (SELECT id FROM queue WHERE name = 'Configuration')
);

-- Also remove the additional queues we added
DELETE FROM queue WHERE name IN ('IT Support', 'HR', 'Sales', 'Finance') 
AND NOT EXISTS (
    SELECT 1 FROM ticket WHERE queue_id = queue.id
);