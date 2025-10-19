-- Drop all tables in the current schema (used to roll back full MySQL schema alignment).

SET @drop_sql = NULL;
SELECT GROUP_CONCAT(CONCAT('DROP TABLE IF EXISTS `', table_name, '`') SEPARATOR '; ')
INTO @drop_sql
FROM information_schema.tables
WHERE table_schema = DATABASE();

SET FOREIGN_KEY_CHECKS = 0;
SET @drop_sql = IFNULL(@drop_sql, 'SELECT 1');
PREPARE stmt FROM @drop_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET FOREIGN_KEY_CHECKS = 1;
