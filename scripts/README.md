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

## Load tests

The load test simulates concurrent users making various API requests (next question, answer, metrics, leaderboard) to measure application latency and throughput.

```bash
# Run with default settings (10 users, 60 seconds)
./scripts/loadtest.sh

# Run with custom settings (e.g., 20 users for 120 seconds)
./scripts/loadtest.sh 20 120
```

### Configuration

You can customize the test behavior by editing `scripts/loadtest_config.env` or by setting environment variables:

- `BASE_URL`: The URL of the running server (default: `http://localhost:3001`).
- `CONCURRENT_USERS`: Number of parallel workers.
- `DURATION_SECONDS`: How long to run the test.
- `USER_ID_MIN` / `USER_ID_MAX`: Range of user IDs to use for requests (ensure these users exist in your database).

### Results

Results are saved in the `loadtest_results/` directory with a timestamp. The script outputs:
- Total requests and successful/error counts.
- Throughput (req/s).
- Latency percentiles (p50, p95, p99) using `time_starttransfer` (server response time).
