-- BrainBolt database schema (MySQL)
-- Run once before using the app. Ensure database exists: CREATE DATABASE IF NOT EXISTS brainbolt;
-- Usage: mysql -u root -p brainbolt < scripts/schema.sql

CREATE TABLE IF NOT EXISTS users (
  id                 INT          AUTO_INCREMENT PRIMARY KEY,
  username           VARCHAR(255) NOT NULL UNIQUE,
  score              BIGINT       NOT NULL DEFAULT 0,
  streak             INT          NOT NULL DEFAULT 0,
  max_streak         INT          NOT NULL DEFAULT 0,
  total_correct      INT          NOT NULL DEFAULT 0,
  total_answered     INT          NOT NULL DEFAULT 0,
  current_difficulty INT          NULL DEFAULT 1,
  last_answer_correct TINYINT(1)  NULL,
  last_answered_at    DATETIME(3) NULL
);

CREATE TABLE IF NOT EXISTS user_questions (
  user_id     INT NOT NULL,
  question_id INT NOT NULL,
  asked_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, question_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_user_questions_user_id (user_id),
  INDEX idx_user_questions_question_id (question_id)
);
