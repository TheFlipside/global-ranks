-- Resolve duplicate usernames before adding uniqueness constraint.
-- For each set of duplicates, keep the earliest created row unchanged
-- and append a numeric suffix to the rest.
WITH duplicates AS (
    SELECT uuid, username,
           ROW_NUMBER() OVER (PARTITION BY username ORDER BY created_at) AS rn
    FROM users
)
UPDATE users
SET username = LEFT(duplicates.username, 28) || '_' || duplicates.rn,
    updated_at = NOW()
FROM duplicates
WHERE users.uuid = duplicates.uuid
  AND duplicates.rn > 1;

DROP INDEX IF EXISTS idx_users_username;
CREATE UNIQUE INDEX idx_users_username ON users (username);
