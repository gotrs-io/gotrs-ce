-- Drop lookup translations table and related objects

DROP TRIGGER IF EXISTS trigger_lookup_translations_change_time ON lookup_translations;
DROP FUNCTION IF EXISTS get_lookup_translation(VARCHAR, VARCHAR, VARCHAR);
DROP TABLE IF EXISTS lookup_translations;