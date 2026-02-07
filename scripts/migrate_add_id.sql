-- Migrate existing users table to add id column
-- This preserves existing data
-- Usage: mysql -u root -p brainbolt < scripts/migrate_add_id.sql

-- Step 1: Check current structure (run this first to see what you have)
-- DESCRIBE users;

-- Step 2: Add id column (if it doesn't exist)
-- First, let's create a temporary table with the correct structure
CREATE TABLE IF NOT EXISTS users_temp (
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

-- Step 3: Copy data from old table to new table (preserving all data)
-- This assumes username exists in the old table
INSERT INTO users_temp (username, score, streak, max_streak, total_correct, total_answered, current_difficulty, last_answer_correct)
SELECT username, score, streak, max_streak, total_correct, total_answered, current_difficulty, last_answer_correct
FROM users
ON DUPLICATE KEY UPDATE users_temp.username = users.username;

-- Step 4: Drop old table and rename new one
DROP TABLE IF EXISTS users;
RENAME TABLE users_temp TO users;

-- Step 5: Recreate user_questions table with proper foreign key
DROP TABLE IF EXISTS user_questions;
CREATE TABLE user_questions (
  user_id     INT NOT NULL,
  question_id INT NOT NULL,
  asked_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, question_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_user_questions_user_id (user_id),
  INDEX idx_user_questions_question_id (question_id)
);
