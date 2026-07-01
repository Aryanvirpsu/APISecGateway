# End-to-End API Security Gateway

A secure reverse-proxy gateway that sits in front of a backend API and enforces a
chain of security controls — request logging, authentication, rate limiting,
IDOR protection and automatic IP blocking — backed by Postgres.

## Focus

- **Bearer-token authentication** — every `/api/*` request needs a valid
  `Authorization: Bearer <token>`; unknown or revoked tokens are rejected
- **API access control** — tokens are scoped per user; IDOR (cross-user access)
  attempts are blocked and alerted
- **Rate limiting** — per-IP token bucket, with automatic IP blocking on
  sustained abuse
- **Request validation & logging** — every request is logged with identity,
  status and latency; security events are recorded as alerts
- **Security middleware chain** — composable controls in front of the upstream

> Scope note: authentication uses static bearer tokens (issued via config), not
> JWT or OAuth 2.0. Those are possible extensions, not part of the current build.

## Components

| Component  | Description                                                        |
| ---------- | ------------------------------------------------------------------ |
| `gateway`  | Reverse proxy on `:8080` that applies the security middleware chain |
| `demo-api` | Backend service on `:8081` exposing `users`/`orders` resources      |
| Postgres   | Stores request logs, alerts, blocked IPs and revoked tokens         |

Requests flow `client → gateway:8080 → demo-api:8081`. Everything under `/api/*`
is proxied (the `/api` prefix is stripped before forwarding); `/health` is served
by the gateway directly.

## Security controls

The middleware chain runs in this order (outermost first):

1. **Recover** — converts a downstream panic into a clean `500`.
2. **Request logging** — assigns a request id (returned as `X-Request-Id`) and
   writes every request to `request_logs` with status, latency and identity.
3. **Block guard** — rejects IPs present in the in-memory block list (`403`).
4. **Rate limiting** — per-IP token bucket. Each rejection returns `429` and
   raises a `rate_limit` alert; sustained abuse adds the IP to `blocked_ips`.
5. **Authentication** — requires a valid `Authorization: Bearer <token>`.
   Unknown or revoked tokens get `401`.
6. **IDOR protection** — a token is scoped to a single user id; reaching another
   user's `/api/users/{id}` resource raises an `idor` alert and returns `403`.

The blocked-IP and revoked-token caches are seeded from the database at start-up
and kept in memory to avoid a database round-trip per request.

## Demo credentials

Two demo tokens ship by default (override with the `API_TOKENS` env var):

| Token         | Subject | Scoped to user |
| ------------- | ------- | -------------- |
| `alice-token` | alice   | `1`            |
| `bob-token`   | bob     | `2`            |

## Endpoints (via the gateway)

| Method & path                | Notes                                |
| ---------------------------- | ------------------------------------ |
| `GET /health`                | Gateway liveness, no auth            |
| `GET /api/users/{id}`        | The caller's own profile             |
| `GET /api/users/{id}/orders` | The caller's own orders              |

## Run (Windows PowerShell)

```powershell
# 1. Start Postgres (schema in ./migrations is applied automatically)
docker compose up -d

# 2. Resolve module dependencies (first run only)
cd gateway;  go mod tidy; cd ..
cd demo-api; go mod tidy; cd ..

# 3. Start the services in separate terminals
.\scripts\run-api.ps1
.\scripts\run-gateway.ps1
```

The equivalent targets are available via `make up`, `make tidy`,
`make run-api` and `make run-gateway`.

## Demo (one shot)

With the stack up (`docker compose up -d` + both services running), run the
demo script. It drives the full attack sequence through the gateway and then
prints what the gateway recorded in Postgres:

```powershell
.\scripts\demo.ps1            # clean run (resets request_logs/alerts first)
.\scripts\demo.ps1 -KeepData  # keep existing rows
```

It walks through: a healthy request, a missing token (`401`), an unknown token
(`401`), two IDOR attempts (`403`), and a burst that trips the rate limiter
(`429`) — then shows the resulting `request_logs`, `alerts` and `blocked_ips`.

To try the pieces by hand:

```bash
# Healthy request to your own resource
curl -H "Authorization: Bearer alice-token" http://localhost:8080/api/users/1

# IDOR attempt — alice reaching bob's data, returns 403
curl -H "Authorization: Bearer alice-token" http://localhost:8080/api/users/2
```

## Tests

The scripts under `tests/` exercise the controls end to end (run them with the
stack up):

```bash
bash tests/smoke.sh        # health + auth happy path
bash tests/idor.sh         # cross-user access is blocked
bash tests/rate_abuse.sh   # rapid traffic trips the rate limiter
```

Inspect what the gateway recorded:

```sql
SELECT ts, method, path, status, source_ip FROM request_logs ORDER BY ts DESC LIMIT 20;
SELECT ts, alert_type, severity, source_ip, reason FROM alerts ORDER BY ts DESC;
SELECT * FROM blocked_ips;
```

## Configuration

All settings have defaults that match `docker-compose.yml`:

| Variable          | Default                  | Purpose                              |
| ----------------- | ------------------------ | ------------------------------------ |
| `GATEWAY_ADDR`    | `:8080`                  | Gateway listen address               |
| `UPSTREAM_URL`    | `http://localhost:8081`  | Backend the gateway proxies to       |
| `DB_HOST`         | `localhost`              | Postgres host                        |
| `DB_PORT`         | `5432`                   | Postgres port                        |
| `DB_USER`         | `secuser`                | Postgres user                        |
| `DB_PASSWORD`     | `secpass`                | Postgres password                    |
| `DB_NAME`         | `apisec`                 | Postgres database                    |
| `RATE_PER_SECOND` | `5`                      | Token refill rate per IP             |
| `RATE_BURST`      | `10`                     | Bucket size per IP                   |
| `BLOCK_AFTER`     | `20`                     | Rejections before an IP is blocked   |
| `API_TOKENS`      | demo set                 | `token:subject:userid,...`           |
