START TRANSACTION;

SET @has_sysconfig_modified := (
  SELECT COUNT(*)
    FROM information_schema.tables
   WHERE table_schema = DATABASE()
     AND table_name = 'sysconfig_modified'
);
SET @has_sysconfig_modified := IFNULL(@has_sysconfig_modified, 0);

SET @has_sysconfig_default := (
  SELECT COUNT(*)
    FROM information_schema.tables
   WHERE table_schema = DATABASE()
     AND table_name = 'sysconfig_default'
);
SET @has_sysconfig_default := IFNULL(@has_sysconfig_default, 0);

SET @sql := IF(@has_sysconfig_modified = 1,
  'DELETE FROM sysconfig_modified WHERE name IN (''CustomerPortal::Enabled'',''CustomerPortal::LoginRequired'',''CustomerPortal::Title'',''CustomerPortal::FooterText'',''CustomerPortal::LandingPage'');',
  'SELECT 0'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(@has_sysconfig_default = 1,
  'DELETE FROM sysconfig_default WHERE name IN (''CustomerPortal::Enabled'',''CustomerPortal::LoginRequired'',''CustomerPortal::Title'',''CustomerPortal::FooterText'',''CustomerPortal::LandingPage'');',
  'SELECT 0'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

COMMIT;
