CREATE TABLE IF NOT EXISTS ai_project_summaries (
    repo_id       UUID        PRIMARY KEY REFERENCES repos(id) ON DELETE CASCADE,
    headline      TEXT        NOT NULL DEFAULT '',
    summary       TEXT        NOT NULL DEFAULT '',
    model_version TEXT        NOT NULL DEFAULT '',
    generated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
