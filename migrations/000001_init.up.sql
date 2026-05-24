CREATE TABLE admin_users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE developers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_username  TEXT NOT NULL UNIQUE,
    email            TEXT NOT NULL,
    display_name     TEXT,
    notes            TEXT,
    avatar_url       TEXT,
    bio              TEXT,
    location         TEXT,
    company          TEXT,
    website          TEXT,
    followers        INT NOT NULL DEFAULT 0,
    following        INT NOT NULL DEFAULT 0,
    public_repos     INT NOT NULL DEFAULT 0,
    total_stars      INT NOT NULL DEFAULT 0,
    activity_score   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    verification_status TEXT NOT NULL DEFAULT 'registered'
        CHECK (verification_status IN ('registered', 'email_verified', 'manual_verified')),
    last_synced_at   TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE repos (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_id    BIGINT NOT NULL UNIQUE,
    owner        TEXT NOT NULL,
    name         TEXT NOT NULL,
    full_name    TEXT NOT NULL UNIQUE,
    description  TEXT,
    url          TEXT NOT NULL,
    language     TEXT,
    stars        INT NOT NULL DEFAULT 0,
    forks        INT NOT NULL DEFAULT 0,
    is_fork      BOOLEAN NOT NULL DEFAULT FALSE,
    pushed_at    TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE developer_repos (
    developer_id UUID NOT NULL REFERENCES developers(id) ON DELETE CASCADE,
    repo_id      UUID NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    PRIMARY KEY (developer_id, repo_id)
);

CREATE TABLE repo_languages (
    repo_id   UUID NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    language  TEXT NOT NULL,
    bytes     BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (repo_id, language)
);

CREATE TABLE contribution_days (
    developer_id UUID NOT NULL REFERENCES developers(id) ON DELETE CASCADE,
    date         DATE NOT NULL,
    count        INT NOT NULL DEFAULT 0,
    PRIMARY KEY (developer_id, date)
);

CREATE TABLE developer_snapshots (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    developer_id   UUID NOT NULL REFERENCES developers(id) ON DELETE CASCADE,
    snapshot_date  DATE NOT NULL,
    public_repos   INT NOT NULL DEFAULT 0,
    total_stars    INT NOT NULL DEFAULT 0,
    followers      INT NOT NULL DEFAULT 0,
    activity_score NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (developer_id, snapshot_date)
);

CREATE TABLE sync_jobs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    developer_id UUID REFERENCES developers(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'done', 'failed')),
    error        TEXT,
    started_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for hot queries
CREATE INDEX idx_developers_activity_score ON developers(activity_score DESC);
CREATE INDEX idx_developers_total_stars    ON developers(total_stars DESC);
CREATE INDEX idx_developer_snapshots_date  ON developer_snapshots(developer_id, snapshot_date DESC);
CREATE INDEX idx_contribution_days_dev     ON contribution_days(developer_id, date DESC);
CREATE INDEX idx_sync_jobs_status          ON sync_jobs(status, created_at DESC);
