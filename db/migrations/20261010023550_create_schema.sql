-- migrate:up
CREATE SCHEMA IF NOT EXISTS app_schema;


-- migrate:down
DROP SCHEMA IF EXISTS app_schema;
