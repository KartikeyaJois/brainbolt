-- Add last_answered_at to users (for existing databases)
-- Usage: mysql -u root -p brainbolt < scripts/add_last_answered_at.sql

ALTER TABLE users ADD COLUMN last_answered_at DATETIME(3) NULL;
