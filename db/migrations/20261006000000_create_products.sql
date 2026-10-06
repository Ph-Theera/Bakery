-- migrate:up
CREATE TABLE public.products (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 sku varchar(64) NOT NULL UNIQUE CHECK (length(btrim(sku)) > 0),
 name varchar(200) NOT NULL CHECK (length(btrim(name)) > 0),
 price_satang bigint NOT NULL CHECK (price_satang >= 0),
 stock bigint NOT NULL CHECK (stock >= 0)
);
-- Access is through the Go backend's database role. No public Data API policies.
ALTER TABLE public.products ENABLE ROW LEVEL SECURITY;
-- migrate:down
DROP TABLE public.products;
