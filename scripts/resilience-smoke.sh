#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir"

base_url="${GATEWAY_BASE_URL:-http://127.0.0.1:26526}"
timeout_seconds="${SMOKE_REQUEST_TIMEOUT_SECONDS:-12}"
temp_file="$(mktemp)"
userdata_was_running=false
redis_was_running=false
docker compose ps --status running --services | rg -qx 'userdata' && userdata_was_running=true
docker compose ps --status running --services | rg -qx 'redis' && redis_was_running=true

cleanup() {
  if [[ "$userdata_was_running" == true ]]; then docker compose start userdata >/dev/null 2>&1 || true; fi
  if [[ "$redis_was_running" == true ]]; then docker compose start redis >/dev/null 2>&1 || true; fi
  rm -f "$temp_file"
}
trap cleanup EXIT INT TERM

request() {
  local path="$1"
  local body="$2"
  curl -sS --max-time "$timeout_seconds" \
    -o "$temp_file" -w '%{http_code} %{time_total}' \
    -H 'Content-Type: application/json' \
    -X POST "$base_url$path" --data "$body"
}

wait_for_userdata() {
  for _ in $(seq 1 40); do
    if curl -fsS --max-time 1 http://127.0.0.1:9091/actuator/prometheus >/dev/null; then
      return 0
    fi
    sleep 0.5
  done
  echo 'UserData did not recover' >&2
  return 1
}

wait_for_redis() {
  for _ in $(seq 1 40); do
    if docker compose exec -T redis redis-cli ping 2>/dev/null | rg -q '^PONG$'; then
      return 0
    fi
    sleep 0.5
  done
  echo 'Redis did not recover' >&2
  return 1
}

echo 'T2: stop UserData; a registration call must fail within the HTTP request budget.'
docker compose stop userdata >/dev/null
result="$(request /login '{"type":"register","accountId":"resilience-unavailable-probe","password":"not-a-real-account"}')"
status="${result%% *}"
elapsed="${result#* }"
echo "  HTTP $status in ${elapsed}s; body: $(cat "$temp_file")"
if [[ "$status" == '000' || "$status" == '200' ]] || awk -v n="$elapsed" 'BEGIN { exit !(n >= 12) }'; then
  echo 'T2 failed: call did not return a bounded error' >&2
  exit 1
fi
docker compose start userdata >/dev/null
wait_for_userdata
echo '  UserData recovered.'

echo 'T3: stop Redis; a session lookup must fail within the request budget.'
docker compose stop redis >/dev/null
result="$(request /login_session '{"sessionId":"resilience-invalid-probe"}')"
status="${result%% *}"
elapsed="${result#* }"
outage_body="$(cat "$temp_file")"
echo "  HTTP $status in ${elapsed}s; body: $outage_body"
if [[ "$status" == '000' ]] || awk -v n="$elapsed" 'BEGIN { exit !(n >= 12) }'; then
  echo 'T3 failed: call did not return a bounded response' >&2
  exit 1
fi
docker compose start redis >/dev/null
wait_for_redis
result="$(request /login_session '{"sessionId":"resilience-invalid-probe"}')"
status="${result%% *}"
recovery_body="$(cat "$temp_file")"
echo "  after Redis recovery: HTTP $status; body: $recovery_body"
if [[ "$status" != '401' || "$recovery_body" != *'Invalid session ID or session expired'* ]]; then
  echo 'T3 failed: Gateway did not observe Redis recovery as an ordinary missing session' >&2
  exit 1
fi
echo 'T3 passed: request returned during outage and Redis recovery was observed.'

health_status="$(curl -sS --max-time 2 -o /dev/null -w '%{http_code}' "$base_url/health")"
if [[ "$health_status" != '200' ]]; then
  echo "Gateway health check failed after smoke tests: HTTP $health_status" >&2
  exit 1
fi
