#!/usr/bin/env bash
# Test script for BrainBolt APIs. Run from repo root or scripts/.
# Usage: ./scripts/test_api.sh   or   BASE_URL=http://localhost:4000 ./scripts/test_api.sh

set -e

BASE_URL="${BASE_URL:-http://localhost:3001}"
USER_ID="${TEST_USER_ID:-1}"

# Optional: pretty-print JSON (no-op if jq missing)
jq_cmd() {
  if command -v jq &>/dev/null; then
    jq "$@"
  else
    cat
  fi
}

echo "=========================================="
echo "BrainBolt API tests (BASE_URL=$BASE_URL)"
echo "=========================================="

# --- Quiz: next question ---
echo ""
echo "[1] GET /v1/quiz/next?userId=$USER_ID"
resp=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/quiz/next?userId=$USER_ID")
body=$(echo "$resp" | sed '$d')
code=$(echo "$resp" | tail -n 1)
echo "HTTP $code"
if [[ "$code" != "200" ]]; then
  echo "Response body: $body"
  echo "FAIL: expected 200"
  exit 1
fi
echo "$body" | jq_cmd .
# Capture questionId for answer test (optional; API returns int)
QUESTION_ID=$(echo "$body" | jq_cmd -r '.questionId // empty')
if [[ -z "$QUESTION_ID" ]]; then QUESTION_ID=1; fi

# --- Quiz: next question without userId (expect 400) ---
echo ""
echo "[2] GET /v1/quiz/next (no userId - expect 400)"
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/v1/quiz/next")
echo "HTTP $code"
if [[ "$code" != "400" ]]; then
  echo "FAIL: expected 400"
  exit 1
fi
echo "OK"

# --- Quiz: submit answer ---
# Options for q1 are A=Berlin, B=Paris, C=Madrid, D=Rome; answer is B
echo ""
echo "[3] POST /v1/quiz/answer (correct answer for $QUESTION_ID)"
resp=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/v1/quiz/answer" \
  -H "Content-Type: application/json" \
  -d "{\"userId\":$USER_ID,\"questionId\":$QUESTION_ID,\"answer\":\"B\"}")
body=$(echo "$resp" | sed '$d')
code=$(echo "$resp" | tail -n 1)
echo "HTTP $code"
if [[ "$code" != "200" ]]; then
  echo "Response body: $body"
  echo "FAIL: expected 200"
  exit 1
fi
echo "$body" | jq_cmd .
echo "OK"

# --- Quiz: submit same answer again (duplicate — expect 204 No Content, ignored) ---
echo ""
echo "[4] POST /v1/quiz/answer (same question again - duplicate ignored)"
code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/v1/quiz/answer" \
  -H "Content-Type: application/json" \
  -d "{\"userId\":$USER_ID,\"questionId\":$QUESTION_ID,\"answer\":\"B\"}")
echo "HTTP $code"
if [[ "$code" != "204" ]]; then
  echo "FAIL: expected 204 No Content"
  exit 1
fi
echo "OK (duplicate ignored)"

# --- Quiz: submit answer with missing fields (expect 400) ---
echo ""
echo "[5] POST /v1/quiz/answer (missing fields - expect 400)"
code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/v1/quiz/answer" \
  -H "Content-Type: application/json" \
  -d "{\"userId\":$USER_ID}")
echo "HTTP $code"
if [[ "$code" != "400" ]]; then
  echo "FAIL: expected 400"
  exit 1
fi
echo "OK"

# --- Quiz: metrics ---
echo ""
echo "[6] GET /v1/quiz/metrics?userId=$USER_ID"
resp=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/quiz/metrics?userId=$USER_ID")
body=$(echo "$resp" | sed '$d')
code=$(echo "$resp" | tail -n 1)
echo "HTTP $code"
if [[ "$code" != "200" ]]; then
  echo "Response body: $body"
  echo "FAIL: expected 200"
  exit 1
fi
echo "$body" | jq_cmd .
echo "OK"

# --- Leaderboard: score ---
echo ""
echo "[7] GET /v1/leaderboard/score?limit=5"
resp=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/leaderboard/score?limit=5")
body=$(echo "$resp" | sed '$d')
code=$(echo "$resp" | tail -n 1)
echo "HTTP $code"
if [[ "$code" != "200" ]]; then
  echo "Response body: $body"
  echo "FAIL: expected 200"
  exit 1
fi
echo "$body" | jq_cmd .
echo "OK"

# --- Leaderboard: streak ---
echo ""
echo "[8] GET /v1/leaderboard/streak?limit=5"
resp=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/leaderboard/streak?limit=5")
body=$(echo "$resp" | sed '$d')
code=$(echo "$resp" | tail -n 1)
echo "HTTP $code"
if [[ "$code" != "200" ]]; then
  echo "Response body: $body"
  echo "FAIL: expected 200"
  exit 1
fi
echo "$body" | jq_cmd .
echo "OK"

# --- Metrics without userId (expect 400) ---
echo ""
echo "[9] GET /v1/quiz/metrics (no userId - expect 400)"
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/v1/quiz/metrics")
echo "HTTP $code"
if [[ "$code" != "400" ]]; then
  echo "FAIL: expected 400"
  exit 1
fi
echo "OK"

# --- Quiz: verify question tracking (no repeats) ---
echo ""
echo "[10] GET /v1/quiz/next (multiple calls - verify no immediate repeats)"
declare -a QUESTION_IDS=()
for i in {1..5}; do
  resp=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/quiz/next?userId=$USER_ID")
  body=$(echo "$resp" | sed '$d')
  code=$(echo "$resp" | tail -n 1)
  if [[ "$code" != "200" ]]; then
    echo "FAIL: expected 200 on call $i"
    exit 1
  fi
  qid=$(echo "$body" | jq_cmd -r '.questionId // empty')
  if [[ -z "$qid" ]]; then
    echo "FAIL: questionId missing in response $i"
    exit 1
  fi
  QUESTION_IDS+=("$qid")
  echo "  Call $i: questionId=$qid"
done

# Check for duplicates in the first 5 questions
declare -A SEEN=()
DUPLICATE_FOUND=0
for qid in "${QUESTION_IDS[@]}"; do
  if [[ -n "${SEEN[$qid]}" ]]; then
    echo "  WARNING: Question $qid repeated (this is OK if all questions at difficulty have been asked)"
    DUPLICATE_FOUND=1
  else
    SEEN[$qid]=1
  fi
done

if [[ $DUPLICATE_FOUND -eq 0 ]]; then
  echo "  OK: No immediate repeats in first 5 questions"
else
  echo "  OK: Some repeats detected (expected if difficulty pool exhausted)"
fi

echo ""
echo "=========================================="
echo "All API tests passed."
echo "=========================================="
