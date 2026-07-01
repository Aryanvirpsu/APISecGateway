#!/usr/bin/env bash
# IDOR check: alice (user 1) tries to read bob's (user 2) resources. The gateway
# should detect the cross-user access, raise an alert and answer 403.
set -euo pipefail

GATEWAY="${GATEWAY:-http://localhost:8080}"
TOKEN="${ALICE_TOKEN:-alice-token}"

echo "== alice tries to read bob's profile (expect 403) =="
code=$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$GATEWAY/api/users/2")
echo "status=$code"
[ "$code" = "403" ] || { echo "UNEXPECTED: IDOR was not blocked"; exit 1; }

echo "== alice tries to read bob's orders (expect 403) =="
code=$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$GATEWAY/api/users/2/orders")
echo "status=$code"
[ "$code" = "403" ] || { echo "UNEXPECTED: IDOR was not blocked"; exit 1; }

echo "IDOR blocked as expected (check the alerts table for the recorded events)"
