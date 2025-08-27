-- ============================================
-- Rollback System Configuration Defaults
-- ============================================

-- Remove all configuration defaults
DELETE FROM sysconfig_default WHERE xml_filename = 'Config.yaml';

-- Also remove any modified settings based on these defaults
DELETE FROM sysconfig_modified WHERE sysconfig_default_id IN (
    SELECT id FROM sysconfig_default WHERE xml_filename = 'Config.yaml'
);