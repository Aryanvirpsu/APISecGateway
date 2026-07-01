#!/usr/bin/env bash
# Rate-limit check: fire a burst of requests from a single client and confirm
# the gateway starts returning 429 once the token bucket is drained.
set -euo pipefail

GATEWAY="${GATEWAY:-http://localhost:8080}"
TOKEN="${ALICE_TOKEN:-alice-token}"
N="${N:-60}"

echo "== firing $N rapid requests to trip the rate limiter =="
limited=0
for _ in $(seq 1 "$N"); do
  code=$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$GATEWAY/api/users/1")
  [ "$code" = "429" ] && limited=$((limited + 1))
done

echo "received $limited rate-limited (429) responses out of $N"
[ "$limited" -gt 0 ] || { echo "UNEXPECTED: no 429s seen, rate limiter did not trip"; exit 1; }

echo "rate limiting works (sustained abuse also lands the IP in blocked_ips)"
