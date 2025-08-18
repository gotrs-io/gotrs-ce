-- GOTRS Lookup Translations Table
-- Provides i18n support for database-driven fields while maintaining OTRS compatibility
-- 
-- This table allows translations for dropdown values without modifying OTRS data structures

CREATE TABLE IF NOT EXISTS lookup_translations (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,      -- 'ticket_states', 'ticket_priorities', 'queues', etc.
    field_value VARCHAR(200) NOT NULL,     -- The actual value from OTRS database
    language_code VARCHAR(10) NOT NULL,    -- 'en', 'de', 'ar', 'tlh', etc.
    translation VARCHAR(200) NOT NULL,     -- Translated display text
    is_system BOOLEAN DEFAULT FALSE,       -- TRUE for GOTRS/OTRS built-in values
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by INTEGER NOT NULL DEFAULT 1,
    change_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    change_by INTEGER NOT NULL DEFAULT 1,
    UNIQUE(table_name, field_value, language_code)
);

-- Create indexes for performance
CREATE INDEX idx_lookup_translations_lookup ON lookup_translations(table_name, field_value, language_code);
CREATE INDEX idx_lookup_translations_system ON lookup_translations(is_system);

-- Seed translations for common OTRS ticket states
-- English (base values - same as database)
INSERT INTO lookup_translations (table_name, field_value, language_code, translation, is_system) VALUES
-- Ticket States
('ticket_states', 'new', 'en', 'New', true),
('ticket_states', 'open', 'en', 'Open', true),
('ticket_states', 'pending reminder', 'en', 'Pending Reminder', true),
('ticket_states', 'pending auto close+', 'en', 'Pending Auto Close+', true),
('ticket_states', 'pending auto close-', 'en', 'Pending Auto Close-', true),
('ticket_states', 'closed successful', 'en', 'Closed Successful', true),
('ticket_states', 'closed unsuccessful', 'en', 'Closed Unsuccessful', true),
('ticket_states', 'merged', 'en', 'Merged', true),
('ticket_states', 'removed', 'en', 'Removed', true),
-- Ticket Priorities (OTRS default)
('ticket_priorities', '1 very low', 'en', '1 Very Low', true),
('ticket_priorities', '2 low', 'en', '2 Low', true),
('ticket_priorities', '3 normal', 'en', '3 Normal', true),
('ticket_priorities', '4 high', 'en', '4 High', true),
('ticket_priorities', '5 very high', 'en', '5 Very High', true);

-- German translations
INSERT INTO lookup_translations (table_name, field_value, language_code, translation, is_system) VALUES
-- Ticket States
('ticket_states', 'new', 'de', 'Neu', true),
('ticket_states', 'open', 'de', 'Offen', true),
('ticket_states', 'pending reminder', 'de', 'Warten zur Erinnerung', true),
('ticket_states', 'pending auto close+', 'de', 'Warten auf automatisches Schließen+', true),
('ticket_states', 'pending auto close-', 'de', 'Warten auf automatisches Schließen-', true),
('ticket_states', 'closed successful', 'de', 'Geschlossen erfolgreich', true),
('ticket_states', 'closed unsuccessful', 'de', 'Geschlossen erfolglos', true),
('ticket_states', 'merged', 'de', 'Zusammengefügt', true),
('ticket_states', 'removed', 'de', 'Entfernt', true),
-- Ticket Priorities
('ticket_priorities', '1 very low', 'de', '1 Sehr niedrig', true),
('ticket_priorities', '2 low', 'de', '2 Niedrig', true),
('ticket_priorities', '3 normal', 'de', '3 Normal', true),
('ticket_priorities', '4 high', 'de', '4 Hoch', true),
('ticket_priorities', '5 very high', 'de', '5 Sehr hoch', true);

-- Spanish translations
INSERT INTO lookup_translations (table_name, field_value, language_code, translation, is_system) VALUES
-- Ticket States
('ticket_states', 'new', 'es', 'Nuevo', true),
('ticket_states', 'open', 'es', 'Abierto', true),
('ticket_states', 'pending reminder', 'es', 'Pendiente de recordatorio', true),
('ticket_states', 'pending auto close+', 'es', 'Pendiente de cierre automático+', true),
('ticket_states', 'pending auto close-', 'es', 'Pendiente de cierre automático-', true),
('ticket_states', 'closed successful', 'es', 'Cerrado con éxito', true),
('ticket_states', 'closed unsuccessful', 'es', 'Cerrado sin éxito', true),
('ticket_states', 'merged', 'es', 'Fusionado', true),
('ticket_states', 'removed', 'es', 'Eliminado', true),
-- Ticket Priorities
('ticket_priorities', '1 very low', 'es', '1 Muy baja', true),
('ticket_priorities', '2 low', 'es', '2 Baja', true),
('ticket_priorities', '3 normal', 'es', '3 Normal', true),
('ticket_priorities', '4 high', 'es', '4 Alta', true),
('ticket_priorities', '5 very high', 'es', '5 Muy alta', true);

-- French translations
INSERT INTO lookup_translations (table_name, field_value, language_code, translation, is_system) VALUES
-- Ticket States
('ticket_states', 'new', 'fr', 'Nouveau', true),
('ticket_states', 'open', 'fr', 'Ouvert', true),
('ticket_states', 'pending reminder', 'fr', 'En attente de rappel', true),
('ticket_states', 'pending auto close+', 'fr', 'En attente de fermeture auto+', true),
('ticket_states', 'pending auto close-', 'fr', 'En attente de fermeture auto-', true),
('ticket_states', 'closed successful', 'fr', 'Fermé avec succès', true),
('ticket_states', 'closed unsuccessful', 'fr', 'Fermé sans succès', true),
('ticket_states', 'merged', 'fr', 'Fusionné', true),
('ticket_states', 'removed', 'fr', 'Supprimé', true),
-- Ticket Priorities
('ticket_priorities', '1 very low', 'fr', '1 Très faible', true),
('ticket_priorities', '2 low', 'fr', '2 Faible', true),
('ticket_priorities', '3 normal', 'fr', '3 Normal', true),
('ticket_priorities', '4 high', 'fr', '4 Élevée', true),
('ticket_priorities', '5 very high', 'fr', '5 Très élevée', true);

-- Arabic translations
INSERT INTO lookup_translations (table_name, field_value, language_code, translation, is_system) VALUES
-- Ticket States
('ticket_states', 'new', 'ar', 'جديد', true),
('ticket_states', 'open', 'ar', 'مفتوح', true),
('ticket_states', 'pending reminder', 'ar', 'في انتظار التذكير', true),
('ticket_states', 'pending auto close+', 'ar', 'في انتظار الإغلاق التلقائي+', true),
('ticket_states', 'pending auto close-', 'ar', 'في انتظار الإغلاق التلقائي-', true),
('ticket_states', 'closed successful', 'ar', 'مغلق بنجاح', true),
('ticket_states', 'closed unsuccessful', 'ar', 'مغلق دون نجاح', true),
('ticket_states', 'merged', 'ar', 'مدمج', true),
('ticket_states', 'removed', 'ar', 'محذوف', true),
-- Ticket Priorities
('ticket_priorities', '1 very low', 'ar', '١ منخفضة جداً', true),
('ticket_priorities', '2 low', 'ar', '٢ منخفضة', true),
('ticket_priorities', '3 normal', 'ar', '٣ عادية', true),
('ticket_priorities', '4 high', 'ar', '٤ عالية', true),
('ticket_priorities', '5 very high', 'ar', '٥ عالية جداً', true);

-- Klingon translations (for fun!)
INSERT INTO lookup_translations (table_name, field_value, language_code, translation, is_system) VALUES
-- Ticket States
('ticket_states', 'new', 'tlh', 'chu''', true),
('ticket_states', 'open', 'tlh', 'poQ', true),
('ticket_states', 'pending reminder', 'tlh', 'loS qej', true),
('ticket_states', 'pending auto close+', 'tlh', 'loS SoQ''egh+', true),
('ticket_states', 'pending auto close-', 'tlh', 'loS SoQ''egh-', true),
('ticket_states', 'closed successful', 'tlh', 'SoQ potlh', true),
('ticket_states', 'closed unsuccessful', 'tlh', 'SoQ potlhbe''', true),
('ticket_states', 'merged', 'tlh', 'rarpu''', true),
('ticket_states', 'removed', 'tlh', 'teq', true),
-- Ticket Priorities
('ticket_priorities', '1 very low', 'tlh', '1 puS qu''', true),
('ticket_priorities', '2 low', 'tlh', '2 puS', true),
('ticket_priorities', '3 normal', 'tlh', '3 motlh', true),
('ticket_priorities', '4 high', 'tlh', '4 naDev', true),
('ticket_priorities', '5 very high', 'tlh', '5 naDev qu''', true);

-- Add more languages as needed
-- Japanese, Chinese, Russian, Portuguese, Italian, Dutch, Polish, Turkish, Hebrew

-- Create function to get translated value
CREATE OR REPLACE FUNCTION get_lookup_translation(
    p_table_name VARCHAR,
    p_field_value VARCHAR,
    p_language_code VARCHAR
) RETURNS VARCHAR AS $$
DECLARE
    v_translation VARCHAR;
BEGIN
    -- Try to get translation for requested language
    SELECT translation INTO v_translation
    FROM lookup_translations
    WHERE table_name = p_table_name
      AND field_value = p_field_value
      AND language_code = p_language_code
    LIMIT 1;
    
    -- If no translation found, return original value
    IF v_translation IS NULL THEN
        RETURN p_field_value;
    END IF;
    
    RETURN v_translation;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update change_time
CREATE TRIGGER trigger_lookup_translations_change_time
    BEFORE UPDATE ON lookup_translations
    FOR EACH ROW
    EXECUTE FUNCTION update_change_time();