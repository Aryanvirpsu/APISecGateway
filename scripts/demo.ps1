# One-shot demo: drives the full attack sequence through the gateway and then
# prints what the gateway recorded in Postgres.
#
# Prerequisites: `docker compose up -d` and both services running
# (scripts\run-api.ps1 + scripts\run-gateway.ps1).
#
# Usage:
#   .\scripts\demo.ps1              # clean run (resets the tables first)
#   .\scripts\demo.ps1 -KeepData    # keep existing rows

param(
    [string]$Gateway = "http://localhost:8080",
    [string]$Token = "alice-token",
    [string]$Container = "api-sec-postgres",
    [switch]$KeepData
)

$ErrorActionPreference = "Stop"

function Write-Head($text) {
    Write-Host ""
    Write-Host "== $text ==" -ForegroundColor Cyan
}

function Get-Status($url, $token, $clientIP) {
    $curlArgs = @("-s", "-o", "NUL", "-w", "%{http_code}")
    if ($token) { $curlArgs += @("-H", "Authorization: Bearer $token") }
    if ($clientIP) { $curlArgs += @("-H", "X-Forwarded-For: $clientIP") }
    $curlArgs += $url
    return (& curl.exe @curlArgs)
}

function Show-Query($title, $sql) {
    Write-Head $title
    docker exec $Container psql -U secuser -d apisec -c $sql
}

# --- Preflight ------------------------------------------------------------
Write-Head "Preflight"
try {
    $health = & curl.exe -s -m 5 "$Gateway/health"
    if ($health -ne "ok") { throw "unexpected health response: '$health'" }
    Write-Host "gateway is up ($Gateway)" -ForegroundColor Green
} catch {
    Write-Host "Gateway not reachable at $Gateway." -ForegroundColor Red
    Write-Host "Start it first:  docker compose up -d ; .\scripts\run-api.ps1 ; .\scripts\run-gateway.ps1"
    exit 1
}

if (-not $KeepData) {
    Write-Head "Resetting request_logs and alerts for a clean run"
    docker exec $Container psql -U secuser -d apisec -c "TRUNCATE request_logs, alerts;" | Out-Null
    docker exec $Container psql -U secuser -d apisec -c "DELETE FROM blocked_ips;" | Out-Null
    Write-Host "done" -ForegroundColor Green
}

# --- Attack sequence ------------------------------------------------------
Write-Head "1. Happy path - alice reads her OWN profile (expect 200)"
& curl.exe -s -H "Authorization: Bearer $Token" "$Gateway/api/users/1"
Write-Host ""

Write-Head "2. AuthN - no token (expect 401)"
Write-Host ("status = {0}" -f (Get-Status "$Gateway/api/users/1" $null))

Write-Head "3. AuthN - unknown token (expect 401)"
Write-Host ("status = {0}" -f (Get-Status "$Gateway/api/users/1" "not-a-real-token"))

Write-Head "4. IDOR - alice tries to read BOB's profile (expect 403)"
Write-Host ("status = {0}" -f (Get-Status "$Gateway/api/users/2" $Token))

Write-Head "5. IDOR - alice tries to read BOB's orders (expect 403)"
Write-Host ("status = {0}" -f (Get-Status "$Gateway/api/users/2/orders" $Token))

# Use a fresh spoofed client IP (via X-Forwarded-For) so the abuse burst never
# blocks the legit caller and the run stays repeatable.
$abuser = "198.51.100." + (Get-Random -Minimum 2 -Maximum 250)
Write-Head "6. Rate limit - fire 60 rapid requests from abuser $abuser (expect a run of 429s)"
$limited = 0
foreach ($i in 1..60) {
    if ((Get-Status "$Gateway/api/users/1" $Token $abuser) -eq "429") { $limited++ }
}
Write-Host ("$limited of 60 requests were rate-limited (429)")

Write-Head "7. Auto-block - abuser $abuser after sustained abuse (expect 403 from block guard)"
Write-Host ("status = {0}" -f (Get-Status "$Gateway/api/users/1" $Token $abuser))

# --- What the gateway recorded -------------------------------------------
Show-Query "Traffic by status code" `
    "SELECT status, count(*) FROM request_logs GROUP BY status ORDER BY status;"

Show-Query "Recent requests (newest first)" `
    "SELECT ts, method, path, status, auth_subject, latency_ms FROM request_logs ORDER BY ts DESC LIMIT 10;"

Show-Query "Security alerts (with attacker context)" `
    "SELECT ts, alert_type, severity, source_ip, auth_subject, reason FROM alerts ORDER BY ts DESC LIMIT 15;"

Show-Query "One IDOR alert in detail" `
    "SELECT source_ip, auth_subject, reason, metadata FROM alerts WHERE alert_type='idor' LIMIT 1;"

Show-Query "Auto-blocked IPs" `
    "SELECT ip, reason, blocked_at FROM blocked_ips ORDER BY blocked_at DESC;"

Write-Host ""
Write-Host "Demo complete." -ForegroundColor Green
