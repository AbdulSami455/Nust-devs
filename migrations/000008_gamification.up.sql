ALTER TABLE developers
    ADD COLUMN current_streak INT NOT NULL DEFAULT 0,
    ADD COLUMN longest_streak INT NOT NULL DEFAULT 0,
    ADD COLUMN streak_multiplier NUMERIC(4, 2) NOT NULL DEFAULT 1.0,
    ADD COLUMN xp INT NOT NULL DEFAULT 0,
    ADD COLUMN power_level INT NOT NULL DEFAULT 1;

CREATE TABLE dev_of_month_winners (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    developer_id    UUID NOT NULL REFERENCES developers(id) ON DELETE CASCADE,
    year            INT NOT NULL,
    month           INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    score           NUMERIC(10, 2) NOT NULL DEFAULT 0,
    activity_points INT NOT NULL DEFAULT 0,
    rank_gain       INT NOT NULL DEFAULT 0,
    stars_gained    INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (year, month)
);

CREATE INDEX idx_developers_current_streak ON developers(current_streak DESC);
CREATE INDEX idx_developers_xp ON developers(xp DESC);
CREATE INDEX idx_developers_power_level ON developers(power_level DESC);
CREATE INDEX idx_dev_of_month_year_month ON dev_of_month_winners(year DESC, month DESC);
