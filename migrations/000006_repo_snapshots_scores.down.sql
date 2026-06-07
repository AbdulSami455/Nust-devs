DROP INDEX IF EXISTS idx_developers_community_score;
DROP INDEX IF EXISTS idx_developers_reviewer_score;
DROP INDEX IF EXISTS idx_developers_contributor_score;
DROP INDEX IF EXISTS idx_developers_builder_score;

ALTER TABLE developer_snapshots
    DROP COLUMN IF EXISTS community_score,
    DROP COLUMN IF EXISTS reviewer_score,
    DROP COLUMN IF EXISTS contributor_score,
    DROP COLUMN IF EXISTS builder_score;

ALTER TABLE developers
    DROP COLUMN IF EXISTS community_score,
    DROP COLUMN IF EXISTS reviewer_score,
    DROP COLUMN IF EXISTS contributor_score,
    DROP COLUMN IF EXISTS builder_score;

DROP TABLE IF EXISTS repo_snapshots;
