-- migrate:up
-- Existing timestamp values are interpreted as UTC.
-- Existing integer THB prices are converted to satang.
CREATE TABLE app_schema.users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar(255),
    role varchar(20) NOT NULL DEFAULT 'customer' CHECK (role IN ('customer','staff','admin')),
    auth_provider varchar(30) CHECK (auth_provider IN ('line','email','google')),
    provider_subject varchar(255),
    is_active boolean NOT NULL DEFAULT true,
    UNIQUE (auth_provider, provider_subject),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES app_schema.users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES app_schema.users(id),
    deleted_at timestamptz,
    deleted_by uuid REFERENCES app_schema.users(id),
    CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))
);
CREATE UNIQUE INDEX ux_users_email ON app_schema.users(lower(email)) WHERE email IS NOT NULL;
CREATE TABLE app_schema.customers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid UNIQUE REFERENCES app_schema.users(id),
    display_name varchar(160) NOT NULL,
    phone varchar(40) NOT NULL,
    line_display_name varchar(160),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES app_schema.users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES app_schema.users(id),
    deleted_at timestamptz,
    deleted_by uuid REFERENCES app_schema.users(id),
    CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))
);

ALTER TABLE app_schema.categories
    ALTER COLUMN id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN slug TYPE varchar(160),
    ALTER COLUMN name_th TYPE varchar(160),
    ADD COLUMN created_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN created_by uuid REFERENCES app_schema.users(id),
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN updated_by uuid REFERENCES app_schema.users(id),
    ADD COLUMN deleted_at timestamptz,
    ADD COLUMN deleted_by uuid REFERENCES app_schema.users(id),
    ADD CONSTRAINT categories_soft_delete_check CHECK ((deleted_at IS NULL) = (deleted_by IS NULL));

ALTER TABLE app_schema.products RENAME COLUMN pric_thb TO price_satang;
ALTER TABLE app_schema.products RENAME COLUMN compare_at_price_thb TO compare_at_price_satang;
ALTER TABLE app_schema.products RENAME COLUMN is_feature TO is_featured;
ALTER TABLE app_schema.products
    ALTER COLUMN id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN name TYPE varchar(200),
    ALTER COLUMN slug TYPE varchar(220),
    ALTER COLUMN price_satang TYPE bigint USING price_satang::bigint * 100,
    ALTER COLUMN compare_at_price_satang TYPE bigint USING compare_at_price_satang::bigint * 100,
    ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE timestamptz USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN deleted_at TYPE timestamptz USING deleted_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN updated_at SET DEFAULT now(),
    ADD CONSTRAINT products_category_fk FOREIGN KEY (category_id) REFERENCES app_schema.categories(id),
    ADD CONSTRAINT products_created_by_fk FOREIGN KEY (created_by) REFERENCES app_schema.users(id),
    ADD CONSTRAINT products_updated_by_fk FOREIGN KEY (updated_by) REFERENCES app_schema.users(id),
    ADD CONSTRAINT products_deleted_by_fk FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id),
    ADD CONSTRAINT products_price_check CHECK (price_satang >= 0),
    ADD CONSTRAINT products_compare_price_check CHECK (compare_at_price_satang IS NULL OR compare_at_price_satang >= price_satang),
    ADD CONSTRAINT products_stock_check CHECK (stock_quantity >= 0),
    ADD CONSTRAINT products_soft_delete_check CHECK ((deleted_at IS NULL) = (deleted_by IS NULL));
CREATE TABLE app_schema.product_images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES app_schema.products(id),
    object_key varchar(500) NOT NULL,
    mime_type varchar(100) NOT NULL,
    alt_text varchar(255),
    sort_order integer NOT NULL DEFAULT 0,
    is_primary boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES app_schema.users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES app_schema.users(id),
    deleted_at timestamptz,
    deleted_by uuid REFERENCES app_schema.users(id),
    CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))
);
CREATE TABLE app_schema.orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number varchar(40) NOT NULL UNIQUE,
    idempotency_key varchar(100) NOT NULL UNIQUE,
    customer_id uuid REFERENCES app_schema.customers(id),
    customer_name_snapshot varchar(160) NOT NULL,
    customer_phone_snapshot varchar(40) NOT NULL,
    fulfillment_type varchar(20) NOT NULL CHECK (fulfillment_type IN ('delivery','pickup')),
    delivery_address_snapshot jsonb,
    delivery_note text,
    pickup_at timestamptz,
    subtotal_satang bigint NOT NULL CHECK (subtotal_satang >= 0),
    delivery_fee_satang bigint NOT NULL DEFAULT 0 CHECK (delivery_fee_satang >= 0),
    discount_satang bigint NOT NULL DEFAULT 0 CHECK (discount_satang >= 0),
    total_satang bigint NOT NULL CHECK (total_satang >= 0),
    currency char(3) NOT NULL DEFAULT 'THB' CHECK (currency = 'THB'),
    status varchar(30) NOT NULL DEFAULT 'pending_payment' CHECK (status IN ('pending_payment','pending_verification','confirmed','preparing','ready_for_pickup','shipped','completed','cancelled','expired')),
    internal_note text,
    placed_at timestamptz,
    CHECK (total_satang = subtotal_satang + delivery_fee_satang - discount_satang),
    CHECK ((fulfillment_type = 'delivery' AND delivery_address_snapshot IS NOT NULL AND pickup_at IS NULL)
        OR (fulfillment_type = 'pickup' AND pickup_at IS NOT NULL AND delivery_address_snapshot IS NULL)),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES app_schema.users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES app_schema.users(id),
    deleted_at timestamptz,
    deleted_by uuid REFERENCES app_schema.users(id),
    CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))
);
CREATE TABLE app_schema.order_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL REFERENCES app_schema.orders(id),
    product_id uuid REFERENCES app_schema.products(id),
    product_name_snapshot varchar(200) NOT NULL,
    unit_price_satang bigint NOT NULL CHECK (unit_price_satang >= 0),
    quantity integer NOT NULL CHECK (quantity > 0),
    line_total_satang bigint NOT NULL,
    special_instructions text,
    CHECK (line_total_satang = unit_price_satang * quantity),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES app_schema.users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES app_schema.users(id),
    deleted_at timestamptz,
    deleted_by uuid REFERENCES app_schema.users(id),
    CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))
);
CREATE TABLE app_schema.payments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL REFERENCES app_schema.orders(id),
    method varchar(30) NOT NULL CHECK (method IN ('bank_transfer','promptpay')),
    expected_amount_satang bigint NOT NULL CHECK (expected_amount_satang >= 0),
    paid_amount_satang bigint CHECK (paid_amount_satang >= 0),
    status varchar(30) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','submitted','approved','rejected','needs_resubmission','expired')),
    verified_by uuid REFERENCES app_schema.users(id),
    verified_at timestamptz,
    rejection_reason text,
    idempotency_key varchar(100) NOT NULL UNIQUE,
    CHECK (status <> 'approved' OR (verified_by IS NOT NULL AND verified_at IS NOT NULL
        AND paid_amount_satang IS NOT NULL AND paid_amount_satang = expected_amount_satang)),
    CHECK (status <> 'rejected' OR NULLIF(btrim(rejection_reason), '') IS NOT NULL),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES app_schema.users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES app_schema.users(id),
    deleted_at timestamptz,
    deleted_by uuid REFERENCES app_schema.users(id),
    CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))
);
CREATE TABLE app_schema.payment_evidence (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id uuid NOT NULL REFERENCES app_schema.payments(id),
    object_key varchar(500) NOT NULL,
    mime_type varchar(100) NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','image/webp')),
    file_size_bytes integer NOT NULL CHECK (file_size_bytes > 0),
    checksum varchar(128) NOT NULL,
    submitted_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES app_schema.users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES app_schema.users(id),
    deleted_at timestamptz,
    deleted_by uuid REFERENCES app_schema.users(id),
    CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))
);
CREATE TABLE app_schema.inventory_movements (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES app_schema.products(id),
    quantity_delta integer NOT NULL CHECK (quantity_delta <> 0),
    reason varchar(30) NOT NULL CHECK (reason IN ('restock','sale','cancellation','refund','adjustment')),
    reference_type varchar(50),
    reference_id uuid,
    idempotency_key varchar(100) NOT NULL UNIQUE,
    created_by uuid REFERENCES app_schema.users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((reference_type IS NULL) = (reference_id IS NULL))
);
CREATE TABLE app_schema.inventory_reservations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL REFERENCES app_schema.orders(id),
    product_id uuid NOT NULL REFERENCES app_schema.products(id),
    quantity integer NOT NULL CHECK (quantity > 0),
    status varchar(20) NOT NULL DEFAULT 'reserved' CHECK (status IN ('reserved','committed','released','expired')),
    expires_at timestamptz NOT NULL,
    released_at timestamptz,
    UNIQUE (order_id, product_id),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES app_schema.users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES app_schema.users(id),
    deleted_at timestamptz,
    deleted_by uuid REFERENCES app_schema.users(id),
    CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))
);
CREATE TABLE app_schema.order_status_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL REFERENCES app_schema.orders(id),
    from_status varchar(30) CHECK (from_status IN ('pending_payment','pending_verification','confirmed','preparing','ready_for_pickup','shipped','completed','cancelled','expired')),
    to_status varchar(30) NOT NULL CHECK (to_status IN ('pending_payment','pending_verification','confirmed','preparing','ready_for_pickup','shipped','completed','cancelled','expired')),
    changed_by uuid REFERENCES app_schema.users(id),
    note text,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE app_schema.audit_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id uuid REFERENCES app_schema.users(id),
    action varchar(80) NOT NULL,
    entity_type varchar(80) NOT NULL,
    entity_id uuid NOT NULL,
    before_json jsonb,
    after_json jsonb,
    request_id varchar(100),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_product_one_primary_image ON app_schema.product_images(product_id) WHERE is_primary AND deleted_at IS NULL;
-- Do not allow soft deletion to erase the fact an order has already been paid.
CREATE UNIQUE INDEX ux_payments_one_approved ON app_schema.payments(order_id) WHERE status = 'approved';
CREATE INDEX ix_products_category_active ON app_schema.products(category_id, is_active, is_featured) WHERE deleted_at IS NULL;
CREATE INDEX ix_orders_customer_created ON app_schema.orders(customer_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX ix_orders_status_created ON app_schema.orders(status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX ix_payments_status_created ON app_schema.payments(status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX ix_inventory_product_created ON app_schema.inventory_movements(product_id, created_at DESC);
CREATE INDEX ix_audit_entity_created ON app_schema.audit_logs(entity_type, entity_id, created_at DESC);
CREATE INDEX ix_order_items_order ON app_schema.order_items(order_id);
CREATE INDEX ix_order_items_product ON app_schema.order_items(product_id);
CREATE INDEX ix_payments_order ON app_schema.payments(order_id);
CREATE INDEX ix_payment_evidence_payment ON app_schema.payment_evidence(payment_id);
CREATE INDEX ix_history_order_created ON app_schema.order_status_history(order_id, created_at DESC);
CREATE INDEX ix_reservations_product ON app_schema.inventory_reservations(product_id) WHERE status = 'reserved' AND deleted_at IS NULL;
CREATE INDEX ix_reservations_expiry ON app_schema.inventory_reservations(expires_at) WHERE status = 'reserved' AND deleted_at IS NULL;

CREATE FUNCTION app_schema.touch_updated_at() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$;
CREATE TRIGGER users_updated_at BEFORE UPDATE ON app_schema.users FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER customers_updated_at BEFORE UPDATE ON app_schema.customers FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER categories_updated_at BEFORE UPDATE ON app_schema.categories FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER products_updated_at BEFORE UPDATE ON app_schema.products FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER product_images_updated_at BEFORE UPDATE ON app_schema.product_images FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER orders_updated_at BEFORE UPDATE ON app_schema.orders FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER order_items_updated_at BEFORE UPDATE ON app_schema.order_items FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER payments_updated_at BEFORE UPDATE ON app_schema.payments FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER payment_evidence_updated_at BEFORE UPDATE ON app_schema.payment_evidence FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();
CREATE TRIGGER inventory_reservations_updated_at BEFORE UPDATE ON app_schema.inventory_reservations FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();

-- migrate:down
-- Refuse rollback if satang values cannot be represented by the old integer THB columns.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM app_schema.products WHERE price_satang % 100 <> 0 OR compare_at_price_satang % 100 <> 0) THEN
        RAISE EXCEPTION 'Cannot rollback: product prices include fractional THB';
    END IF;
END;
$$;
DROP TRIGGER users_updated_at ON app_schema.users;
DROP TRIGGER customers_updated_at ON app_schema.customers;
DROP TRIGGER categories_updated_at ON app_schema.categories;
DROP TRIGGER products_updated_at ON app_schema.products;
DROP TRIGGER product_images_updated_at ON app_schema.product_images;
DROP TRIGGER orders_updated_at ON app_schema.orders;
DROP TRIGGER order_items_updated_at ON app_schema.order_items;
DROP TRIGGER payments_updated_at ON app_schema.payments;
DROP TRIGGER payment_evidence_updated_at ON app_schema.payment_evidence;
DROP TRIGGER inventory_reservations_updated_at ON app_schema.inventory_reservations;
DROP FUNCTION app_schema.touch_updated_at();
DROP TABLE app_schema.audit_logs;
DROP TABLE app_schema.order_status_history;
DROP TABLE app_schema.inventory_reservations;
DROP TABLE app_schema.inventory_movements;
DROP TABLE app_schema.payment_evidence;
DROP TABLE app_schema.payments;
DROP TABLE app_schema.order_items;
DROP TABLE app_schema.orders;
DROP TABLE app_schema.product_images;

DROP INDEX app_schema.ix_products_category_active;
ALTER TABLE app_schema.products
    DROP CONSTRAINT products_category_fk,
    DROP CONSTRAINT products_created_by_fk,
    DROP CONSTRAINT products_updated_by_fk,
    DROP CONSTRAINT products_deleted_by_fk,
    DROP CONSTRAINT products_price_check,
    DROP CONSTRAINT products_compare_price_check,
    DROP CONSTRAINT products_stock_check,
    DROP CONSTRAINT products_soft_delete_check,
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN name TYPE varchar,
    ALTER COLUMN slug TYPE varchar,
    ALTER COLUMN created_at DROP DEFAULT,
    ALTER COLUMN updated_at DROP DEFAULT,
    ALTER COLUMN price_satang TYPE integer USING (price_satang / 100)::integer,
    ALTER COLUMN compare_at_price_satang TYPE integer USING (compare_at_price_satang / 100)::integer,
    ALTER COLUMN created_at TYPE timestamp USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE timestamp USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN deleted_at TYPE timestamp USING deleted_at AT TIME ZONE 'UTC';
ALTER TABLE app_schema.products RENAME COLUMN price_satang TO pric_thb;
ALTER TABLE app_schema.products RENAME COLUMN compare_at_price_satang TO compare_at_price_thb;
ALTER TABLE app_schema.products RENAME COLUMN is_featured TO is_feature;
ALTER TABLE app_schema.categories
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN slug TYPE varchar,
    ALTER COLUMN name_th TYPE varchar,
    DROP COLUMN created_at,
    DROP COLUMN created_by,
    DROP COLUMN updated_at,
    DROP COLUMN updated_by,
    DROP COLUMN deleted_at,
    DROP COLUMN deleted_by;
DROP TABLE app_schema.customers;
DROP TABLE app_schema.users;
