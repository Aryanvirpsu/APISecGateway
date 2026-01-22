# End-to-End API Security System

Two services:

- `demo-api` (backend API)
- `gateway` (reverse proxy + security controls)

Infra (Docker):

- Postgres
- Redis

## Run (Windows PowerShell)

```powershell
docker compose up -d
.\scripts\run-api.ps1
.\scripts\run-gateway.ps1
```
