-- Create user_questions table
-- Usage: mysql -u root -p brainbolt < scripts/create_user_questions_table.sql

CREATE TABLE IF NOT EXISTS user_questions (
  user_id     INT NOT NULL,
  question_id INT NOT NULL,
  asked_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, question_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_user_questions_user_id (user_id),
  INDEX idx_user_questions_question_id (question_id)
);
