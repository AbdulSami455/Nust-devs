CREATE TABLE IF NOT EXISTS ai_normalized_tags (
    entity_type   TEXT        NOT NULL,
    entity_id     UUID        NOT NULL,
    headline      TEXT        NOT NULL DEFAULT '',
    summary       TEXT        NOT NULL DEFAULT '',
    languages     TEXT[]      NOT NULL DEFAULT '{}',
    skills        TEXT[]      NOT NULL DEFAULT '{}',
    tags          TEXT[]      NOT NULL DEFAULT '{}',
    model_version TEXT        NOT NULL DEFAULT '',
    generated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (entity_type, entity_id)
);
