ALTER TABLE developer_requests
    DROP COLUMN IF EXISTS course,
    DROP COLUMN IF EXISTS batch;
