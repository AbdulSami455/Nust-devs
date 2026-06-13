CREATE TABLE IF NOT EXISTS ai_developer_summaries (
    developer_id  UUID        PRIMARY KEY REFERENCES developers(id) ON DELETE CASCADE,
    headline      TEXT        NOT NULL DEFAULT '',
    summary       TEXT        NOT NULL DEFAULT '',
    strengths     TEXT[]      NOT NULL DEFAULT '{}',
    model_version TEXT        NOT NULL DEFAULT '',
    generated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_eval_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_name  TEXT        NOT NULL,
    input_hash  TEXT        NOT NULL,
    output      JSONB       NOT NULL DEFAULT '{}',
    latency_ms  INT         NOT NULL DEFAULT 0,
    success     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_eval_logs_agent_created
    ON ai_eval_logs (agent_name, created_at DESC);
