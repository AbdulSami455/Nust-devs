CREATE TABLE IF NOT EXISTS audit_logs (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_type    TEXT        NOT NULL,
    actor_id      TEXT,
    action        TEXT        NOT NULL,
    resource_type TEXT        NOT NULL DEFAULT '',
    resource_id   TEXT        NOT NULL DEFAULT '',
    method        TEXT        NOT NULL DEFAULT '',
    path          TEXT        NOT NULL DEFAULT '',
    status_code   INT         NOT NULL DEFAULT 0,
    ip            TEXT        NOT NULL DEFAULT '',
    user_agent    TEXT        NOT NULL DEFAULT '',
    metadata      JSONB       NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at
    ON audit_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_created
    ON audit_logs (actor_type, created_at DESC);

CREATE TABLE IF NOT EXISTS agent_runs (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id     TEXT        NOT NULL UNIQUE,
    agent_name     TEXT        NOT NULL,
    user_message   TEXT        NOT NULL DEFAULT '',
    input_hash     TEXT        NOT NULL DEFAULT '',
    status         TEXT        NOT NULL DEFAULT 'running',
    ip             TEXT        NOT NULL DEFAULT '',
    user_agent     TEXT        NOT NULL DEFAULT '',
    tool_calls     INT         NOT NULL DEFAULT 0,
    latency_ms     INT         NOT NULL DEFAULT 0,
    error_message  TEXT        NOT NULL DEFAULT '',
    response_chars INT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_agent_runs_created_at
    ON agent_runs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_runs_status_created
    ON agent_runs (status, created_at DESC);

CREATE TABLE IF NOT EXISTS agent_run_events (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id        UUID        NOT NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
    event_type    TEXT        NOT NULL,
    tool_name     TEXT        NOT NULL DEFAULT '',
    payload       JSONB       NOT NULL DEFAULT '{}',
    latency_ms    INT         NOT NULL DEFAULT 0,
    success       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_run_events_run_created
    ON agent_run_events (run_id, created_at ASC);
