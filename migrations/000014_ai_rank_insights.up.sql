CREATE TABLE IF NOT EXISTS ai_rank_insights (
    developer_id  UUID        PRIMARY KEY REFERENCES developers(id) ON DELETE CASCADE,
    headline      TEXT        NOT NULL DEFAULT '',
    summary       TEXT        NOT NULL DEFAULT '',
    highlights    TEXT[]      NOT NULL DEFAULT '{}',
    model_version TEXT        NOT NULL DEFAULT '',
    generated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
