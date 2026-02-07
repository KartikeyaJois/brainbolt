-- BrainBolt database schema (MySQL)
-- Run once before using the app. Ensure database exists: CREATE DATABASE IF NOT EXISTS brainbolt;
-- Usage: mysql -u root -p brainbolt < scripts/schema.sql

CREATE TABLE IF NOT EXISTS users (
  username           VARCHAR(255) NOT NULL PRIMARY KEY,
  score              BIGINT       NOT NULL DEFAULT 0,
  streak             INT          NOT NULL DEFAULT 0,
  max_streak         INT          NOT NULL DEFAULT 0,
  total_correct      INT          NOT NULL DEFAULT 0,
  total_answered     INT          NOT NULL DEFAULT 0,
  current_difficulty INT          NULL DEFAULT 1,
  last_answer_correct TINYINT(1)  NULL
);
