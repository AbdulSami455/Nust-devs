UPDATE developers SET email = '' WHERE email IS NULL;
ALTER TABLE developers ALTER COLUMN email SET NOT NULL;
