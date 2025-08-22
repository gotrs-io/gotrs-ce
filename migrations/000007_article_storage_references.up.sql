-- Article storage references table for tracking files in different backends
-- This is used when ArticleStorageFS is enabled to track file locations

CREATE TABLE IF NOT EXISTS article_storage_references (
    id BIGSERIAL PRIMARY KEY,
    article_id BIGINT NOT NULL REFERENCES article(id) ON DELETE CASCADE,
    backend VARCHAR(10) NOT NULL, -- 'DB' or 'FS'
    location TEXT NOT NULL, -- File path for FS, table reference for DB
    content_type VARCHAR(250),
    file_name VARCHAR(250) NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    checksum VARCHAR(64), -- SHA256 hash
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accessed_time TIMESTAMP,
    UNIQUE(article_id, file_name, backend)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_article_storage_article_id ON article_storage_references(article_id);
CREATE INDEX IF NOT EXISTS idx_article_storage_backend ON article_storage_references(backend);
CREATE INDEX IF NOT EXISTS idx_article_storage_created ON article_storage_references(created_time);

-- Migration status tracking
CREATE TABLE IF NOT EXISTS article_storage_migration (
    id SERIAL PRIMARY KEY,
    source_backend VARCHAR(10) NOT NULL,
    target_backend VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'pending', 'in_progress', 'completed', 'failed'
    total_articles INTEGER DEFAULT 0,
    processed_articles INTEGER DEFAULT 0,
    failed_articles INTEGER DEFAULT 0,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    last_article_id BIGINT DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Function to update migration status
CREATE OR REPLACE FUNCTION update_migration_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for auto-updating updated_at
CREATE TRIGGER update_article_storage_migration_timestamp
    BEFORE UPDATE ON article_storage_migration
    FOR EACH ROW
    EXECUTE FUNCTION update_migration_timestamp();