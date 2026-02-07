#!/usr/bin/env bash
# Diagnostic script to check if everything is set up correctly
# Usage: ./scripts/check_setup.sh

set -e

DB_USER="${MYSQL_USER:-root}"
DB_NAME="${MYSQL_DB:-brainbolt}"

echo "=========================================="
echo "BrainBolt Setup Diagnostic"
echo "=========================================="
echo ""

echo "[1] Checking database connection..."
if mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" -e "USE $DB_NAME;" 2>/dev/null; then
  echo "✓ Database connection OK"
else
  echo "✗ Database connection FAILED"
  echo "  Make sure MySQL is running and credentials are correct"
  exit 1
fi

echo ""
echo "[2] Checking users table structure..."
if mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME" -e "DESCRIBE users;" 2>/dev/null | grep -q "id"; then
  echo "✓ users table has 'id' column"
else
  echo "✗ users table missing 'id' column"
  echo "  Run: mysql -u root -p brainbolt < scripts/recreate_schema.sql"
  exit 1
fi

echo ""
echo "[3] Checking user_questions table..."
if mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME" -e "DESCRIBE user_questions;" 2>/dev/null > /dev/null; then
  echo "✓ user_questions table exists"
else
  echo "✗ user_questions table missing"
  echo "  Run: mysql -u root -p brainbolt < scripts/schema.sql"
  exit 1
fi

echo ""
echo "[4] Checking for users..."
USER_COUNT=$(mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "0")
if [ "$USER_COUNT" -gt 0 ]; then
  echo "✓ Found $USER_COUNT user(s)"
  echo ""
  echo "Existing users:"
  mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME" -e "SELECT id, username FROM users;" 2>/dev/null
else
  echo "✗ No users found"
  echo "  Create a user with:"
  echo "  mysql -u root -p brainbolt -e \"INSERT INTO users (username, score, streak, max_streak, total_correct, total_answered, current_difficulty) VALUES ('testuser', 0, 0, 0, 0, 0, 1);\""
  exit 1
fi

echo ""
echo "=========================================="
echo "Setup looks good! You can use userId=1 (or any existing user ID)"
echo "=========================================="
