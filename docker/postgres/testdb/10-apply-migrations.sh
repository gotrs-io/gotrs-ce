#!/bin/bash
set -euo pipefail

psql=(psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB")

MIGRATIONS_DIR="/docker-entrypoint-initdb.d/migrations"

apply_migration() {
    local file="$1"
    echo "Applying migration: $file"
    "${psql[@]}" -f "$MIGRATIONS_DIR/$file"
}

MIGRATION_FILES=(
    000001_schema_alignment.up.sql
)

for file in "${MIGRATION_FILES[@]}"; do
    if [ -f "$MIGRATIONS_DIR/$file" ]; then
        apply_migration "$file"
    else
        echo "Skipping missing migration: $file"
    fi
done

echo "All migrations applied successfully."
