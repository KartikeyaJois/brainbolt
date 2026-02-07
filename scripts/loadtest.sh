#!/usr/bin/env bash
# BrainBolt load test — bash only (curl + shell).
# Usage:
#   ./scripts/loadtest.sh
#   ./scripts/loadtest.sh 20 120   # 20 users, 120 seconds
#   BASE_URL=http://localhost:3001 DURATION_SECONDS=30 ./scripts/loadtest.sh
#
# Requires: curl, awk. Optional: jq (for parsing next questionId for /answer).

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/loadtest_config.env"
if [[ -f "$CONFIG_FILE" ]]; then
  # shellcheck source=loadtest_config.env
  source "$CONFIG_FILE"
fi

BASE_URL="${BASE_URL:-http://localhost:3001}"
CONCURRENT_USERS="${1:-${CONCURRENT_USERS:-10}}"
DURATION_SECONDS="${2:-${DURATION_SECONDS:-60}}"
USER_ID_MIN="${USER_ID_MIN:-1}"
USER_ID_MAX="${USER_ID_MAX:-50}"
RESULTS_DIR="${RESULTS_DIR:-./loadtest_results}"
PCT_NEXT="${PCT_NEXT:-35}"
PCT_ANSWER="${PCT_ANSWER:-35}"
PCT_METRICS="${PCT_METRICS:-15}"
PCT_LEADERBOARD_SCORE="${PCT_LEADERBOARD_SCORE:-8}"
PCT_LEADERBOARD_STREAK="${PCT_LEADERBOARD_STREAK:-7}"

mkdir -p "$RESULTS_DIR"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
RESULTS_FILE="${RESULTS_DIR}/loadtest_${TIMESTAMP}.txt"
LATENCY_FILE="${RESULTS_DIR}/latency_${TIMESTAMP}.tmp"

echo "=========================================="
echo "BrainBolt Load Test (bash)"
echo "=========================================="
echo "BASE_URL=$BASE_URL"
echo "CONCURRENT_USERS=$CONCURRENT_USERS"
echo "DURATION_SECONDS=$DURATION_SECONDS"
echo "Request mix: next=${PCT_NEXT}% answer=${PCT_ANSWER}% metrics=${PCT_METRICS}% lb_score=${PCT_LEADERBOARD_SCORE}% lb_streak=${PCT_LEADERBOARD_STREAK}%"
echo "User ID range: $USER_ID_MIN .. $USER_ID_MAX"
echo "Results: $RESULTS_FILE"
echo "=========================================="

# Millisecond timestamp (portable: macOS date doesn't support %N)
now_ms() {
  if date +%s%3N 2>/dev/null | grep -q '^[0-9]\{10,15\}$'; then
    date +%s%3N
  else
    python3 -c 'import time; print(int(time.time()*1000))' 2>/dev/null || echo "$(date +%s)000"
  fi
}

# Write one line per request: "duration_ms http_code endpoint"
log_request() {
  echo "$1 $2 $3" >> "$LATENCY_FILE"
}

random_user_id() {
  echo $((USER_ID_MIN + RANDOM % (USER_ID_MAX - USER_ID_MIN + 1)))
}

# Returns 0-99
random_pct() {
  echo $((RANDOM % 100))
}

run_one_request() {
  local p
  p=$(random_pct)
  local uid
  uid=$(random_user_id)

  if (( p < PCT_NEXT )); then
    local start end code
    start=$(now_ms)
    code=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/v1/quiz/next?userId=${uid}")
    end=$(now_ms)
    log_request $(( end - start )) "$code" "next"
    return 0
  fi

  if (( p < PCT_NEXT + PCT_ANSWER )); then
    local start end code body
    start=$(now_ms)
    body=$(curl -s "${BASE_URL}/v1/quiz/next?userId=${uid}")
    questionId=1
    if command -v jq &>/dev/null; then
      questionId=$(echo "$body" | jq -r '.questionId // 1')
    fi
    # Submit one of A/B/C/D
    code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/v1/quiz/answer" \
      -H "Content-Type: application/json" \
      -d "{\"userId\":${uid},\"questionId\":${questionId},\"answer\":\"A\"}")
    end=$(now_ms)
    log_request $(( end - start )) "$code" "answer"
    return 0
  fi

  if (( p < PCT_NEXT + PCT_ANSWER + PCT_METRICS )); then
    local start end code
    start=$(now_ms)
    code=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/v1/quiz/metrics?userId=${uid}")
    end=$(now_ms)
    log_request $(( end - start )) "$code" "metrics"
    return 0
  fi

  if (( p < PCT_NEXT + PCT_ANSWER + PCT_METRICS + PCT_LEADERBOARD_SCORE )); then
    local start end code
    start=$(now_ms)
    code=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/v1/leaderboard/score?limit=10")
    end=$(now_ms)
    log_request $(( end - start )) "$code" "leaderboard_score"
    return 0
  fi

  local start end code
  start=$(now_ms)
  code=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/v1/leaderboard/streak?limit=10")
  end=$(now_ms)
  log_request $(( end - start )) "$code" "leaderboard_streak"
}

export BASE_URL USER_ID_MIN USER_ID_MAX PCT_NEXT PCT_ANSWER PCT_METRICS PCT_LEADERBOARD_SCORE PCT_LEADERBOARD_STREAK
export LATENCY_FILE
export -f now_ms log_request random_user_id random_pct run_one_request

# Clear latency file
: > "$LATENCY_FILE"

end_epoch=$(( $(date +%s) + DURATION_SECONDS ))
worker() {
  while (( $(date +%s) < end_epoch )); do
    run_one_request
  done
}

echo "Starting load at $(date +%H:%M:%S) for ${DURATION_SECONDS}s..."
for ((i=0; i<CONCURRENT_USERS; i++)); do
  worker &
done
wait
echo "Load finished at $(date +%H:%M:%S)."

# Aggregate results
total=$(wc -l < "$LATENCY_FILE")
if [[ "$total" -eq 0 ]]; then
  echo "No requests recorded."
  exit 0
fi

errors=$(awk '$2 !~ /^2[0-9][0-9]$/ { count++ } END { print 0+count }' "$LATENCY_FILE")
ok=$(( total - errors ))
error_pct="0"
if [[ $total -gt 0 ]]; then
  error_pct=$(awk "BEGIN { printf \"%.2f\", $errors * 100 / $total }")
fi
duration_actual="$DURATION_SECONDS"
throughput=$(awk "BEGIN { printf \"%.1f\", $total / $duration_actual }")
rpm=$(awk "BEGIN { printf \"%.0f\", $total * 60 / $duration_actual }")

# Latency percentiles (ms)
sort -n -k1,1 "$LATENCY_FILE" > "${LATENCY_FILE}.sorted"
p50=$(awk -v n="$total" '{ a[NR]=$1 } END { i=int(n*0.5); if(i<1)i=1; print a[i] }' "${LATENCY_FILE}.sorted")
p95=$(awk -v n="$total" '{ a[NR]=$1 } END { i=int(n*0.95); if(i<1)i=1; print a[i] }' "${LATENCY_FILE}.sorted")
p99=$(awk -v n="$total" '{ a[NR]=$1 } END { i=int(n*0.99); if(i<1)i=1; print a[i] }' "${LATENCY_FILE}.sorted")

{
  echo "=========================================="
  echo "BrainBolt Load Test Results"
  echo "=========================================="
  echo "Timestamp: $TIMESTAMP"
  echo "Duration: ${DURATION_SECONDS}s | Users: $CONCURRENT_USERS"
  echo "Total requests: $total"
  echo "Successful (2xx): $ok | Errors: $errors (${error_pct}%)"
  echo "Throughput: ${throughput} req/s"
  echo "RPM (requests/min): $rpm"
  echo "Latency (ms): p50=$p50 p95=$p95 p99=$p99"
  echo "------------------------------------------"
  echo "CPU/Memory: not measured by this script. Run server under top/ps or your monitoring and correlate with RPM above."
} | tee "$RESULTS_FILE"

# Per-endpoint summary (optional)
echo "" >> "$RESULTS_FILE"
echo "Per-endpoint counts:" >> "$RESULTS_FILE"
awk '{ print $3 }' "$LATENCY_FILE" | sort | uniq -c | sort -rn >> "$RESULTS_FILE"

rm -f "${LATENCY_FILE}.sorted"
echo ""
echo "Results saved to $RESULTS_FILE"
