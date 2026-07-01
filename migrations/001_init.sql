create table if not exists request_logs (
  id uuid primary key,
  ts timestamptz not null default now(),
  request_id uuid not null,
  source_ip text not null,
  method text not null,
  path text not null,
  status int not null,
  latency_ms int not null,
  user_agent text,
  auth_subject text,
  token_id text
);

create table if not exists alerts (
  id uuid primary key,
  ts timestamptz not null default now(),
  request_id uuid,
  source_ip text not null,
  auth_subject text,
  alert_type text not null,
  severity int not null,
  reason text not null,
  metadata jsonb
);

create table if not exists blocked_ips (
  ip text primary key,
  blocked_at timestamptz default now(),
  reason text
);

create table if not exists revoked_tokens (
  token_id text primary key,
  revoked_at timestamptz default now(),
  reason text
);

-- Helpful indexes for the dashboards/queries you'll likely run.
create index if not exists request_logs_ts_idx on request_logs (ts desc);
create index if not exists request_logs_source_ip_idx on request_logs (source_ip);
create index if not exists alerts_ts_idx on alerts (ts desc);
create index if not exists alerts_type_idx on alerts (alert_type);
