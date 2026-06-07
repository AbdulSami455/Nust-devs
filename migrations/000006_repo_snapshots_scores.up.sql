CREATE TABLE repo_snapshots (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id       UUID NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    snapshot_date DATE NOT NULL,
    stars         INT NOT NULL DEFAULT 0,
    forks         INT NOT NULL DEFAULT 0,
    pushed_at     TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (repo_id, snapshot_date)
);

CREATE INDEX idx_repo_snapshots_repo_date ON repo_snapshots(repo_id, snapshot_date DESC);
CREATE INDEX idx_repo_snapshots_date ON repo_snapshots(snapshot_date DESC);

ALTER TABLE developers
    ADD COLUMN builder_score      NUMERIC(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN contributor_score  NUMERIC(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN reviewer_score     NUMERIC(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN community_score    NUMERIC(10, 2) NOT NULL DEFAULT 0;

ALTER TABLE developer_snapshots
    ADD COLUMN builder_score      NUMERIC(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN contributor_score  NUMERIC(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN reviewer_score     NUMERIC(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN community_score    NUMERIC(10, 2) NOT NULL DEFAULT 0;

CREATE INDEX idx_developers_builder_score ON developers(builder_score DESC);
CREATE INDEX idx_developers_contributor_score ON developers(contributor_score DESC);
CREATE INDEX idx_developers_reviewer_score ON developers(reviewer_score DESC);
CREATE INDEX idx_developers_community_score ON developers(community_score DESC);
