-- Fix schema: Add id column if missing, or recreate table
-- Usage: mysql -u root -p brainbolt < scripts/fix_schema.sql

-- First, check if we need to add the id column
-- If the table exists without id, we'll need to handle it

-- Option 1: Drop and recreate (WARNING: This will delete all data!)
-- DROP TABLE IF EXISTS user_questions;
-- DROP TABLE IF EXISTS users;

-- Option 2: Add id column if missing (safer, but may fail if primary key exists)
-- Check if id column exists first
SET @col_exists = (
  SELECT COUNT(*) 
  FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = 'brainbolt' 
    AND TABLE_NAME = 'users' 
    AND COLUMN_NAME = 'id'
);

-- If id doesn't exist, we need to add it
-- But this is complex if there's already a primary key, so let's just recreate

-- Safe approach: Create new table structure
CREATE TABLE IF NOT EXISTS users_new (
  id                 INT          AUTO_INCREMENT PRIMARY KEY,
  username           VARCHAR(255) NOT NULL UNIQUE,
  score              BIGINT       NOT NULL DEFAULT 0,
  streak             INT          NOT NULL DEFAULT 0,
  max_streak         INT          NOT NULL DEFAULT 0,
  total_correct      INT          NOT NULL DEFAULT 0,
  total_answered     INT          NOT NULL DEFAULT 0,
  current_difficulty INT          NULL DEFAULT 1,
  last_answer_correct TINYINT(1)  NULL
);

-- Migrate data if old table exists (assuming username was primary key before)
-- INSERT INTO users_new (username, score, streak, max_streak, total_correct, total_answered, current_difficulty, last_answer_correct)
-- SELECT username, score, streak, max_streak, total_correct, total_answered, current_difficulty, last_answer_correct
-- FROM users
-- ON DUPLICATE KEY UPDATE users_new.username = users.username;

-- Drop old table and rename new one
-- DROP TABLE IF EXISTS users;
-- RENAME TABLE users_new TO users;

-- Actually, simpler: Just recreate the table structure
-- But first backup your data if needed!
