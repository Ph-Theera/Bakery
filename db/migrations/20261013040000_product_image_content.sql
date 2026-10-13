-- migrate:up
ALTER TABLE app_schema.product_images ADD COLUMN content bytea;

-- migrate:down
ALTER TABLE app_schema.product_images DROP COLUMN content;
