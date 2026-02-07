-- Add 2 test users for leaderboard tests (run after schema)
-- Usage: mysql -u root -p brainbolt < scripts/seed_two_users.sql
-- Or from MySQL: USE brainbolt; then paste the INSERT below.

INSERT INTO users (username, score, streak, max_streak, total_correct, total_answered, current_difficulty)
VALUES
  ('testuser2', 200, 2, 4, 10, 12, 1),
  ('testuser3', 150, 1, 6, 8, 10, 1);
