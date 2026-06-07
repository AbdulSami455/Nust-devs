ALTER TABLE developers
    ADD COLUMN pr_contributions INT NOT NULL DEFAULT 0,
    ADD COLUMN issue_contributions INT NOT NULL DEFAULT 0,
    ADD COLUMN review_contributions INT NOT NULL DEFAULT 0,
    ADD COLUMN contribution_period_start DATE,
    ADD COLUMN contribution_period_end DATE;

CREATE TABLE developer_external_contributions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    developer_id    UUID NOT NULL REFERENCES developers(id) ON DELETE CASCADE,
    repo_full_name  TEXT NOT NULL,
    repo_url        TEXT NOT NULL,
    pr_count        INT NOT NULL DEFAULT 0,
    issue_count     INT NOT NULL DEFAULT 0,
    review_count    INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (developer_id, repo_full_name)
);

CREATE INDEX idx_dev_ext_contrib_dev ON developer_external_contributions(developer_id);
