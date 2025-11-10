#!/bin/bash
# Скрипт для применения миграции 003_remove_artists_foreign_key.sql

# Получаем DATABASE_URL из переменной окружения или используем значение по умолчанию
DATABASE_URL=${DATABASE_URL:-"postgres://postgres:postgres@localhost:5432/tracks_db?sslmode=disable"}

echo "Applying migration 003_remove_artists_foreign_key.sql..."
psql "$DATABASE_URL" -f migrations/003_remove_artists_foreign_key.sql

if [ $? -eq 0 ]; then
    echo "Migration applied successfully!"
else
    echo "Failed to apply migration!"
    exit 1
fi

