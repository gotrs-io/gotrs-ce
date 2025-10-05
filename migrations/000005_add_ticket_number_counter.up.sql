-- Create ticket_number_counter table to support AutoIncrement and date-based ticket number generators
-- Allows multiple zero-value rows per counter_uid for concurrency batching
-- Do NOT add a UNIQUE constraint on counter_uid; algorithm depends on duplicate rows during allocation window

CREATE TABLE IF NOT EXISTS ticket_number_counter (
    id BIGSERIAL PRIMARY KEY,
    counter BIGINT NOT NULL DEFAULT 0,
    counter_uid VARCHAR(255) NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index to accelerate lookups for a given uid and zero/non-zero filtering
CREATE INDEX IF NOT EXISTS idx_ticket_number_counter_uid ON ticket_number_counter(counter_uid);
CREATE INDEX IF NOT EXISTS idx_ticket_number_counter_uid_counter ON ticket_number_counter(counter_uid, counter);
