CREATE TABLE developer_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_username TEXT NOT NULL,
    email           TEXT,
    display_name    TEXT,
    message         TEXT,
    status          TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected')),
    admin_notes     TEXT,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_developer_requests_username_pending
    ON developer_requests (lower(github_username))
    WHERE status = 'pending';

CREATE INDEX idx_developer_requests_status_created
    ON developer_requests (status, created_at DESC);
