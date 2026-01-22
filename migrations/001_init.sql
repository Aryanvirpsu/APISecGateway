create table request_logs (
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

create table alerts (
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

create table blocked_ips (
  ip text primary key,
  blocked_at timestamptz default now(),
  reason text
);

create table revoked_tokens (
  token_id text primary key,
  revoked_at timestamptz default now(),
  reason text
);
