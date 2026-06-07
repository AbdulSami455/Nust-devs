DROP TABLE IF EXISTS dev_of_month_winners;

ALTER TABLE developers
    DROP COLUMN IF EXISTS power_level,
    DROP COLUMN IF EXISTS xp,
    DROP COLUMN IF EXISTS streak_multiplier,
    DROP COLUMN IF EXISTS longest_streak,
    DROP COLUMN IF EXISTS current_streak;
