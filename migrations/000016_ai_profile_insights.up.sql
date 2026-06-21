CREATE TABLE IF NOT EXISTS ai_profile_insights (
    developer_id         UUID        PRIMARY KEY REFERENCES developers(id) ON DELETE CASCADE,
    headline             TEXT        NOT NULL DEFAULT '',
    recent_activity_recap TEXT       NOT NULL DEFAULT '',
    top_achievements     TEXT[]      NOT NULL DEFAULT '{}',
    completion_tips      TEXT[]      NOT NULL DEFAULT '{}',
    model_version        TEXT        NOT NULL DEFAULT '',
    generated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
