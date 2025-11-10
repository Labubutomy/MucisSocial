-- Удаляем foreign key constraint на artists, так как tracks-service больше не хранит артистов
-- Артисты хранятся в artists-service, tracks-service только маппит связи

-- Получаем имя constraint (может отличаться в зависимости от версии PostgreSQL)
DO $$
DECLARE
    constraint_name text;
BEGIN
    -- Ищем constraint по имени колонки
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'track_artists'::regclass
      AND confrelid = 'artists'::regclass
      AND contype = 'f'
    LIMIT 1;
    
    -- Удаляем constraint, если он существует
    IF constraint_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE track_artists DROP CONSTRAINT %I', constraint_name);
    END IF;
END $$;

-- Удаляем таблицу artists, так как она больше не нужна в tracks-service
DROP TABLE IF EXISTS artists;

