DROP TABLE IF EXISTS developer_external_contributions;

ALTER TABLE developers
    DROP COLUMN IF EXISTS contribution_period_end,
    DROP COLUMN IF EXISTS contribution_period_start,
    DROP COLUMN IF EXISTS review_contributions,
    DROP COLUMN IF EXISTS issue_contributions,
    DROP COLUMN IF EXISTS pr_contributions;
