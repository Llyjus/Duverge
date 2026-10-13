#!/usr/bin/env bash
set -u

BASE_URL="${BASE_URL:-http://127.0.0.1:26526}"
CURL_TIMEOUT="${CURL_TIMEOUT:-10}"
REDIS_CONTAINER="${REDIS_CONTAINER:-redis}"
USER_ID="curl_test_$(date +%s)_$$"
SECOND_USER="${USER_ID}_second"
PASSWORD="curl-test-password"
SESSION_ID=""
SECOND_SESSION_ID=""
ROOM_ID=""
PASSED=0
FAILED=0

cleanup() {
  rm -f "${TMP_DIR}/response"
  rmdir "${TMP_DIR}" 2>/dev/null || true
}

TMP_DIR="$(mktemp -d)"
trap cleanup EXIT

print_result() {
  local name="$1" expected="$2" actual="$3" body="$4"
  if [[ "$actual" == "$expected" ]]; then
    printf 'PASS %-30s expected=%s actual=%s\n' "$name" "$expected" "$actual"
    PASSED=$((PASSED + 1))
  else
    printf 'FAIL %-30s expected=%s actual=%s body=%s\n' \
      "$name" "$expected" "$actual" "$(tr '\n' ' ' < "$body")"
    FAILED=$((FAILED + 1))
  fi
}

request() {
  local name="$1" method="$2" path="$3" expected="$4" payload="${5:-}"
  local body="${TMP_DIR}/response"
  local status

  if [[ -n "$payload" ]]; then
    status="$(curl --silent --show-error --max-time "$CURL_TIMEOUT" \
      --request "$method" \
      --header 'Content-Type: application/json' \
      --data "$payload" \
      --output "$body" \
      --write-out '%{http_code}' \
      "${BASE_URL}${path}" 2>"${TMP_DIR}/curl-error")"
  else
    status="$(curl --silent --show-error --max-time "$CURL_TIMEOUT" \
      --request "$method" \
      --output "$body" \
      --write-out '%{http_code}' \
      "${BASE_URL}${path}" 2>"${TMP_DIR}/curl-error")"
  fi

  if [[ "$status" == "000" ]]; then
    printf 'FAIL %-30s expected=%s actual=000 curl_error=%s\n' \
      "$name" "$expected" "$(tr '\n' ' ' < "${TMP_DIR}/curl-error")"
    FAILED=$((FAILED + 1))
  else
    print_result "$name" "$expected" "$status" "$body"
  fi

  cat "$body" 2>/dev/null | tr '\n' ' ' | cut -c1-240
  printf '\n'
}

extract_json_value() {
  local key="$1"
  sed -nE "s/.*\"${key}\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" \
    "${TMP_DIR}/response" | head -n1
}

seed_lobby_status() {
  if ! command -v docker >/dev/null 2>&1; then
    printf 'FAIL %-30s docker command is required to seed Room test state\n' \
      "seed second user status"
    FAILED=$((FAILED + 1))
    return
  fi

  if docker compose exec -T "$REDIS_CONTAINER" redis-cli \
      SET "user_status:${SECOND_USER}" \
      "{\"user_id\":\"${SECOND_USER}\",\"status\":\"in_lobby\"}" \
      >/dev/null; then
    printf 'PASS %-30s second user status seeded\n' "seed second user status"
    PASSED=$((PASSED + 1))
  else
    printf 'FAIL %-30s unable to seed Redis lobby status\n' \
      "seed second user status"
    FAILED=$((FAILED + 1))
  fi
}

echo "Testing ${BASE_URL}"

# Public health endpoint.
request "GET /health" GET "/health" 200

# Invalid input checks for every JSON route.
request "POST /login invalid JSON" POST "/login" 400 '{'
request "POST /login invalid type" POST "/login" 400 \
  '{"type":"unknown","accountId":"invalid","password":"invalid"}'
request "POST /login_session invalid JSON" POST "/login_session" 400 '{'
request "POST /room/user invalid JSON" POST "/room/user" 400 '{'
request "POST /room/userPreparation invalid JSON" POST "/room/userPreparation" 400 '{'
request "POST /room/room invalid JSON" POST "/room/room" 400 '{'
request "POST /room/endGame invalid JSON" POST "/room/endGame" 400 '{'

# Registration and login flow.
request "POST /login register" POST "/login" 200 \
  "{\"type\":\"register\",\"accountId\":\"${USER_ID}\",\"password\":\"${PASSWORD}\"}"
request "POST /login login" POST "/login" 200 \
  "{\"type\":\"login\",\"accountId\":\"${USER_ID}\",\"password\":\"${PASSWORD}\"}"
SESSION_ID="$(extract_json_value sessionId)"

if [[ -z "$SESSION_ID" ]]; then
  printf 'FAIL %-30s login response did not contain sessionId\n' "extract sessionId"
  FAILED=$((FAILED + 1))
else
  printf 'PASS %-30s sessionId received\n' "extract sessionId"
  PASSED=$((PASSED + 1))
fi

if [[ -n "$SESSION_ID" ]]; then
  request "POST /login_session valid before failure" POST "/login_session" 200 \
    "{\"sessionId\":\"${SESSION_ID}\"}"
fi

request "POST /login wrong password" POST "/login" 401 \
  "{\"type\":\"login\",\"accountId\":\"${USER_ID}\",\"password\":\"wrong-password\"}"
request "POST /login_session invalid session" POST "/login_session" 401 \
  '{"sessionId":"invalid-session"}'

if [[ -n "$SESSION_ID" ]]; then
  # This check intentionally detects whether a failed login invalidates a valid session.
  request "POST /login_session valid after failure" POST "/login_session" 200 \
    "{\"sessionId\":\"${SESSION_ID}\"}"

  # Re-authenticate so the remaining endpoint checks can run even when the
  # failed-login session regression is present.
  request "POST /login re-login for room flow" POST "/login" 200 \
    "{\"type\":\"login\",\"accountId\":\"${USER_ID}\",\"password\":\"${PASSWORD}\"}"
  SESSION_ID="$(extract_json_value sessionId)"

  request "POST /room/user createRoom" POST "/room/user" 200 \
    "{\"type\":\"createRoom\",\"sessionId\":\"${SESSION_ID}\"}"
  ROOM_ID="$(extract_json_value roomId)"

  if [[ -z "$ROOM_ID" ]]; then
    printf 'FAIL %-30s createRoom response did not contain roomId\n' "extract roomId"
    FAILED=$((FAILED + 1))
    ROOM_ID="missing-room"
  else
    printf 'PASS %-30s roomId received\n' "extract roomId"
    PASSED=$((PASSED + 1))
  fi

  request "POST /login register second user" POST "/login" 200 \
    "{\"type\":\"register\",\"accountId\":\"${SECOND_USER}\",\"password\":\"${PASSWORD}\"}"
  request "POST /login login second user" POST "/login" 200 \
    "{\"type\":\"login\",\"accountId\":\"${SECOND_USER}\",\"password\":\"${PASSWORD}\"}"
  SECOND_SESSION_ID="$(extract_json_value sessionId)"
  seed_lobby_status

  request "POST /room/room getRoomInfo" POST "/room/room" 200 \
    "{\"type\":\"getRoomInfo\",\"sessionId\":\"${SESSION_ID}\",\"roomId\":\"${ROOM_ID}\"}"
  request "POST /room/room joinRoom" POST "/room/room" 200 \
    "{\"type\":\"joinRoom\",\"sessionId\":\"${SECOND_SESSION_ID}\",\"roomId\":\"${ROOM_ID}\"}"
  request "POST /room/userPreparation setPreparation" POST "/room/userPreparation" 200 \
    "{\"type\":\"setPreparation\",\"sessionId\":\"${SESSION_ID}\",\"roomId\":\"${ROOM_ID}\",\"targetStatus\":true}"
  request "POST /room/userPreparation second user ready" POST "/room/userPreparation" 200 \
    "{\"type\":\"setPreparation\",\"sessionId\":\"${SECOND_SESSION_ID}\",\"roomId\":\"${ROOM_ID}\",\"targetStatus\":true}"
  request "POST /room/room startGame" POST "/room/room" 200 \
    "{\"type\":\"startGame\",\"sessionId\":\"${SESSION_ID}\",\"roomId\":\"${ROOM_ID}\"}"
  request "POST /room/endGame endGame" POST "/room/endGame" 200 \
    "{\"type\":\"endGame\",\"roomId\":\"${ROOM_ID}\"}"
  request "POST /room/room leaveRoom" POST "/room/room" 200 \
    "{\"type\":\"leaveRoom\",\"sessionId\":\"${SECOND_SESSION_ID}\",\"roomId\":\"${ROOM_ID}\"}"
else
  printf 'SKIP %-30s no valid session was created\n' "room flow"
  FAILED=$((FAILED + 1))
fi

printf '\nSummary: %d passed, %d failed\n' "$PASSED" "$FAILED"
if (( FAILED > 0 )); then
  exit 1
fi
