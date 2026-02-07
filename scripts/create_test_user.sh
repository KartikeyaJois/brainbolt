#!/usr/bin/env bash
# Helper script to create a test user or get existing user ID
# Usage: ./scripts/create_test_user.sh [username]

set -e

USERNAME="${1:-testuser}"
DB_USER="${MYSQL_USER:-root}"
DB_NAME="${MYSQL_DB:-brainbolt}"

echo "Checking for existing user: $USERNAME"
EXISTING_ID=$(mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME" -sN -e "SELECT id FROM users WHERE username = '$USERNAME';" 2>/dev/null || echo "")

if [ -n "$EXISTING_ID" ]; then
  echo "User '$USERNAME' already exists with ID: $EXISTING_ID"
  echo "You can use userId=$EXISTING_ID in your API calls"
  exit 0
fi

echo "Creating new user: $USERNAME"
mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME" -e "INSERT INTO users (username, score, streak, max_streak, total_correct, total_answered, current_difficulty) VALUES ('$USERNAME', 0, 0, 0, 0, 0, 1);"

NEW_ID=$(mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME" -sN -e "SELECT id FROM users WHERE username = '$USERNAME';" 2>/dev/null)
echo "User '$USERNAME' created with ID: $NEW_ID"
echo "You can use userId=$NEW_ID in your API calls"
