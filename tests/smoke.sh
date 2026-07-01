#!/usr/bin/env bash
# Basic end-to-end check: health, an authenticated request to the caller's own
# resources, and that an unauthenticated request is rejected.
set -euo pipefail

GATEWAY="${GATEWAY:-http://localhost:8080}"
TOKEN="${ALICE_TOKEN:-alice-token}"

echo "== gateway health =="
curl -fsS "$GATEWAY/health"; echo

echo "== alice reads her own profile =="
curl -fsS -H "Authorization: Bearer $TOKEN" "$GATEWAY/api/users/1"; echo

echo "== alice reads her own orders =="
curl -fsS -H "Authorization: Bearer $TOKEN" "$GATEWAY/api/users/1/orders"; echo

echo "== request without a token is rejected (expect 401) =="
code=$(curl -s -o /dev/null -w '%{http_code}' "$GATEWAY/api/users/1")
echo "status=$code"
[ "$code" = "401" ] || { echo "UNEXPECTED: expected 401"; exit 1; }

echo "smoke OK"
