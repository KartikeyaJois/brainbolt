#!/usr/bin/env bash
# BrainBolt load test — Optimized for App Latency measurement.
# Usage:
#   ./scripts/loadtest.sh
#   ./scripts/loadtest.sh 20 120   # 20 users, 120 seconds

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/loadtest_config.env"
if [[ -f "$CONFIG_FILE" ]]; then
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
echo "BrainBolt Load Test (Optimized)"
echo "=========================================="
echo "BASE_URL=$BASE_URL"
echo "CONCURRENT_USERS=$CONCURRENT_USERS"
echo "DURATION_SECONDS=$DURATION_SECONDS"
echo "User ID range: $USER_ID_MIN .. $USER_ID_MAX"
echo "=========================================="

# Write one line per request: "latency_ms http_code endpoint"
log_request() {
  echo "$1 $2 $3" >> "$LATENCY_FILE"
}

random_user_id() {
  echo $((USER_ID_MIN + RANDOM % (USER_ID_MAX - USER_ID_MIN + 1)))
}

random_pct() {
  echo $((RANDOM % 100))
}

run_one_request() {
  local p=$(random_pct)
  local uid=$(random_user_id)
  local out endpoint

  if (( p < PCT_NEXT )); then
    endpoint="next"
    # %{time_starttransfer} is the time from start until first byte is received.
    # This represents the "App Latency" + Network RTT.
    out=$(curl -s -o /dev/null -w "%{http_code} %{time_starttransfer}" "${BASE_URL}/v1/quiz/next?userId=${uid}")
  elif (( p < PCT_NEXT + PCT_ANSWER )); then
    endpoint="answer"
    # For /answer, we first need a questionId. We'll just use a random one 1..50 to save an extra call
    local qid=$((1 + RANDOM % 50))
    out=$(curl -s -o /dev/null -w "%{http_code} %{time_starttransfer}" -X POST "${BASE_URL}/v1/quiz/answer" \
      -H "Content-Type: application/json" \
      -d "{\"userId\":${uid},\"questionId\":${qid},\"answer\":\"A\"}")
  elif (( p < PCT_NEXT + PCT_ANSWER + PCT_METRICS )); then
    endpoint="metrics"
    out=$(curl -s -o /dev/null -w "%{http_code} %{time_starttransfer}" "${BASE_URL}/v1/quiz/metrics?userId=${uid}")
  elif (( p < PCT_NEXT + PCT_ANSWER + PCT_METRICS + PCT_LEADERBOARD_SCORE )); then
    endpoint="lb_score"
    out=$(curl -s -o /dev/null -w "%{http_code} %{time_starttransfer}" "${BASE_URL}/v1/leaderboard/score?limit=10")
  else
    endpoint="lb_streak"
    out=$(curl -s -o /dev/null -w "%{http_code} %{time_starttransfer}" "${BASE_URL}/v1/leaderboard/streak?limit=10")
  fi

  local code=$(echo $out | cut -d' ' -f1)
  local lat_sec=$(echo $out | cut -d' ' -f2)
  # Convert seconds to milliseconds
  local lat_ms=$(awk "BEGIN { printf \"%.0f\", $lat_sec * 1000 }")
  log_request "$lat_ms" "$code" "$endpoint"
}

export -f log_request random_user_id random_pct run_one_request
export BASE_URL USER_ID_MIN USER_ID_MAX PCT_NEXT PCT_ANSWER PCT_METRICS PCT_LEADERBOARD_SCORE PCT_LEADERBOARD_STREAK LATENCY_FILE

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

total=$(wc -l < "$LATENCY_FILE")
if [[ "$total" -eq 0 ]]; then
  echo "No requests recorded."
  exit 0
fi

errors=$(awk '$2 !~ /^2[0-9][0-9]$/ { count++ } END { print 0+count }' "$LATENCY_FILE")
ok=$(( total - errors ))
error_pct=$(awk "BEGIN { printf \"%.2f\", $errors * 100 / $total }")
throughput=$(awk "BEGIN { printf \"%.1f\", $total / $DURATION_SECONDS }")

sort -n -k1,1 "$LATENCY_FILE" > "${LATENCY_FILE}.sorted"
p50=$(awk -v n="$total" '{ a[NR]=$1 } END { i=int(n*0.5); if(i<1)i=1; print a[i] }' "${LATENCY_FILE}.sorted")
p95=$(awk -v n="$total" '{ a[NR]=$1 } END { i=int(n*0.95); if(i<1)i=1; print a[i] }' "${LATENCY_FILE}.sorted")
p99=$(awk -v n="$total" '{ a[NR]=$1 } END { i=int(n*0.99); if(i<1)i=1; print a[i] }' "${LATENCY_FILE}.sorted")

{
  echo "=========================================="
  echo "BrainBolt Load Test Results (App Latency)"
  echo "=========================================="
  echo "Note: Latency now uses curl's %{time_starttransfer},"
  echo "which mimics the time taken by the server to respond."
  echo "------------------------------------------"
  echo "Duration: ${DURATION_SECONDS}s | Users: $CONCURRENT_USERS"
  echo "Total requests: $total"
  echo "Successful: $ok | Errors: $errors (${error_pct}%)"
  echo "Throughput: ${throughput} req/s"
  echo "Latency (ms): p50=$p50 p95=$p95 p99=$p99"
  echo "------------------------------------------"
} | tee "$RESULTS_FILE"

rm -f "$LATENCY_FILE" "${LATENCY_FILE}.sorted"
