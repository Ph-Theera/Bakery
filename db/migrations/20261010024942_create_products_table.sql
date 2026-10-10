-- migrate:up
CREATE TABLE app_schema.products (
    "id" uuid PRIMARY KEY,
    "category_id" uuid NOT NULL,
    "name" varchar NOT NULL,
    "slug" varchar UNIQUE,
    "description_th" text NOT NULL,
    "pric_thb" integer NOT NULL,
    "compare_at_price_thb" integer,
    "stock_quantity" integer NOT NULL,
    "is_active" boolean NOT NULL,
    "is_feature" boolean NOT NULL,
    "created_at" timestamp NOT NULL,
    "created_by" uuid,
    "updated_at" timestamp NOT NULL,
    "updated_by" uuid,
    "deleted_at" timestamp,
    "deleted_by" uuid
);

-- migrate:down

DROP TABLE IF EXISTS app_schema.products;