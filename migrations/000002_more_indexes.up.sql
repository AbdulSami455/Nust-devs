-- Faster lookup by GitHub username (very frequent query)
CREATE INDEX IF NOT EXISTS idx_developers_github_username ON developers(github_username);

-- Leaderboard sorts
CREATE INDEX IF NOT EXISTS idx_developers_followers    ON developers(followers DESC);
CREATE INDEX IF NOT EXISTS idx_developers_public_repos ON developers(public_repos DESC);

-- Top projects
CREATE INDEX IF NOT EXISTS idx_repos_stars ON repos(stars DESC);

-- Language stats aggregation
CREATE INDEX IF NOT EXISTS idx_repo_languages_language ON repo_languages(language);
