-- migrate:up
CREATE TABLE IF NOT EXISTS schema_seeds (
    name VARCHAR(255) NOT NULL PRIMARY KEY,
    executed_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- migrate:down
DROP TABLE IF EXISTS schema_seeds;
