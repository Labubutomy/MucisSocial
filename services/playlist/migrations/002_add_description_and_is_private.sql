-- Добавляем поля description и is_private в таблицу playlists
ALTER TABLE playlists 
ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS is_private BOOLEAN DEFAULT false;

