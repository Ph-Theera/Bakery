-- migrate:up
CREATE TABLE app_schema.categories (
    id uuid PRIMARY KEY,
    slug varchar UNIQUE,
    name_th varchar,
    sort_order integer,
    is_active boolean
);

-- migrate:down

DROP TABLE IF EXISTS app_schema.categories;