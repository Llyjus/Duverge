# Runtime Validation Report

**Generated**: 2026-10-17T20:11:52+01:00
**Target**: Duverge Docker Compose stack

## Summary

| Step | Status | Exit Code | Details |
|------|--------|-----------|---------|
| Startup | PASS | n/a | Gateway, Room, UserData, MySQL, and Redis were running; `GET /health` returned 200 and `ok`. |
| Pure black-box login/heartbeat | PASS | 0 for heartbeat assertions | Two users logged in over HTTP, authenticated over TCP, and received successful heartbeat responses at 0s, 10s, and 20s. |
| Pure black-box room flow | FAIL | 1 | Room creation succeeded, but joining failed because TCP authentication stores `user_status:<user>` as the scalar string `in_lobby`, while Room expects a JSON `UserStatus`. |
| Full chat flow with explicit fixture | PASS | 0 | After seeding the expected JSON status shape as a test fixture, room join, in-room heartbeat, room chat, and private chat all passed. |

**Overall**: NEEDS SIGNOFF

Heartbeat itself works. The unmodified end-to-end flow is blocked by the status-format mismatch described above. The chat result is valid for the rest of the stack, but used an explicit Redis fixture to bypass that pre-existing blocker.

## Test Evidence

- Runnable test: [`scripts/heartbeat_blackbox.go`](scripts/heartbeat_blackbox.go)
- Final command: `SEED_STATUS=1 go run scripts/heartbeat_blackbox.go`
- Final result: exit code `0`
- Heartbeat cadence: two users, every 10 seconds for 20 seconds before joining; one additional heartbeat per user while in the room
- Redis evidence immediately after the final run: both `user_status:*` keys had TTL `26` seconds, confirming renewal from the 30-second expiry window
- Redis was reset with `FLUSHALL` before the final run; it contained 7 test keys afterward

## Issues Found

| # | Severity | Description | Evidence |
|---|----------|-------------|----------|
| 1 | HIGH | TCP authentication writes `user_status:<user>` as `in_lobby`, but Room unmarshals that key as JSON. The pure flow therefore cannot join a room after login. | `user_status` contains `in_lobby`; Room join expects `{"user_id":"...","status":"in_lobby"}` |
| 2 | MEDIUM | Successful TCP authentication has no response packet. The client must infer authentication completion from subsequent behavior or Redis state. | The test had to treat auth as write-only because the handler continues reading without sending a success response. |

