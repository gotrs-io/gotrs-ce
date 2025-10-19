-- Remove minimal upstream seed data from MySQL deployments.
START TRANSACTION;

DELETE FROM queue WHERE id IN (1,2,3,4);
DELETE FROM signature WHERE id = 1;
DELETE FROM salutation WHERE id = 1;
DELETE FROM system_address WHERE id = 1;
DELETE FROM groups WHERE id IN (1,2,3);
DELETE FROM follow_up_possible WHERE id IN (1,2,3);
DELETE FROM communication_channel WHERE id IN (1,2,3,4);
DELETE FROM article_sender_type WHERE id IN (1,2,3);
DELETE FROM ticket_lock_type WHERE id IN (1,2);
DELETE FROM ticket_type WHERE id IN (1,2,3,4,5);
DELETE FROM ticket_priority WHERE id IN (1,2,3,4,5);
DELETE FROM ticket_state WHERE id IN (1,2,3,4,5);
DELETE FROM ticket_state_type WHERE id IN (1,2,3,4,5);
DELETE FROM valid WHERE id IN (1,2,3);

COMMIT;
