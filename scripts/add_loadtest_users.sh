#!/usr/bin/env bash
# Add N loadtest users to the DB (ids 1..N, INSERT IGNORE). Safe to run multiple times.
# Usage: ./scripts/add_loadtest_users.sh [N]
#   N defaults to 50. Uses MYSQL_USER, MYSQL_PASSWORD, MYSQL_DB (default: root, -, brainbolt).

set -e

N="${1:-50}"
DB_USER="${MYSQL_USER:-root}"
DB_NAME="${MYSQL_DB:-brainbolt}"

if ! [[ "$N" =~ ^[0-9]+$ ]] || [[ "$N" -lt 1 ]]; then
  echo "Usage: $0 [N]" >&2
  echo "  N = number of users to ensure (ids 1..N). Default 50." >&2
  exit 1
fi

echo "Adding loadtest users 1..$N to $DB_NAME (INSERT IGNORE)..."

{
  echo "INSERT IGNORE INTO users (id, username, score, streak, max_streak, total_correct, total_answered, current_difficulty) VALUES"
  for i in $(seq 1 "$N"); do
    [[ $i -gt 1 ]] && echo -n ","
    echo "  ($i, 'loadtest_$i', 0, 0, 0, 0, 0, 1)"
  done
  echo ";"
} | mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME"

count=$(mysql -u "$DB_USER" -p"${MYSQL_PASSWORD:-}" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM users WHERE id BETWEEN 1 AND $N;" 2>/dev/null || echo "?")
echo "Done. Users with id 1..$N: $count."
echo "Set USER_ID_MAX=$N in scripts/loadtest_config.env to use them in the load test."
