# Scripts

## Database setup (required before first run)

Create the database and `users` table:

```bash
# Create DB if needed (MySQL client)
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS brainbolt;"

# Apply schema
mysql -u root -p brainbolt < scripts/schema.sql
```

Use the same user/password as your app (`MYSQL_USER`, `MYSQL_PASSWORD` env vars or defaults in `database.go`).

## API tests

Run against a running BrainBolt server (default: http://localhost:3001).

```bash
# From repo root
./scripts/test_api.sh

# Custom base URL
BASE_URL=http://localhost:4000 ./scripts/test_api.sh
```

Requires: `curl`, `jq` (optional, for pretty-printed JSON).
