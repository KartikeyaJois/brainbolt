# BrainBolt

BrainBolt is a high-performance quiz game backend built with Go, MySQL, and Redis.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Or Go 1.21+, MySQL, and Redis (for local development)

### Running with Docker

The easiest way to get started is using Docker Compose:

```bash
docker-compose up --build
```

This will start:
- **API Server** at `http://localhost:3001`
- **MySQL** at `localhost:3307`
- **Redis** at `localhost:6380`

The database is automatically initialized with the schema and test data.

## Scripts & Testing

Documentation for utility scripts can be found in the [scripts directory](./scripts/README.md).

### API Tests

Run the following command to test the API endpoints:

```bash
./scripts/test_api.sh
```

### Load Testing

To perform a load test against the running server:

```bash
# Default: 10 concurrent users for 60 seconds
./scripts/loadtest.sh

# Custom: 20 concurrent users for 120 seconds
./scripts/loadtest.sh 20 120
```

Load test results, including latency percentiles (p50, p95, p99) and throughput, are saved in the `loadtest_results/` directory.

Configuration for the load test can be modified in `scripts/loadtest_config.env`.
