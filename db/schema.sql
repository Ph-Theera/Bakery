SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: app_schema; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA app_schema;


--
-- Name: touch_updated_at(); Type: FUNCTION; Schema: app_schema; Owner: -
--

CREATE FUNCTION app_schema.touch_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_logs; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    actor_id uuid,
    action character varying(80) NOT NULL,
    entity_type character varying(80) NOT NULL,
    entity_id uuid NOT NULL,
    before_json jsonb,
    after_json jsonb,
    request_id character varying(100),
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: categories; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    slug character varying(160),
    name_th character varying(160),
    sort_order integer,
    is_active boolean,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    CONSTRAINT categories_soft_delete_check CHECK (((deleted_at IS NULL) = (deleted_by IS NULL)))
);


--
-- Name: customers; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.customers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    display_name character varying(160) NOT NULL,
    phone character varying(40) NOT NULL,
    line_display_name character varying(160),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    CONSTRAINT customers_check CHECK (((deleted_at IS NULL) = (deleted_by IS NULL)))
);


--
-- Name: inventory_movements; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.inventory_movements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    quantity_delta integer NOT NULL,
    reason character varying(30) NOT NULL,
    reference_type character varying(50),
    reference_id uuid,
    idempotency_key character varying(100) NOT NULL,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT inventory_movements_check CHECK (((reference_type IS NULL) = (reference_id IS NULL))),
    CONSTRAINT inventory_movements_quantity_delta_check CHECK ((quantity_delta <> 0)),
    CONSTRAINT inventory_movements_reason_check CHECK (((reason)::text = ANY ((ARRAY['restock'::character varying, 'sale'::character varying, 'cancellation'::character varying, 'refund'::character varying, 'adjustment'::character varying])::text[])))
);


--
-- Name: inventory_reservations; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.inventory_reservations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    product_id uuid NOT NULL,
    quantity integer NOT NULL,
    status character varying(20) DEFAULT 'reserved'::character varying NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    released_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    CONSTRAINT inventory_reservations_check CHECK (((deleted_at IS NULL) = (deleted_by IS NULL))),
    CONSTRAINT inventory_reservations_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT inventory_reservations_status_check CHECK (((status)::text = ANY ((ARRAY['reserved'::character varying, 'committed'::character varying, 'released'::character varying, 'expired'::character varying])::text[])))
);


--
-- Name: notification_outbox; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.notification_outbox (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    payload text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    next_attempt_at timestamp with time zone DEFAULT now() NOT NULL,
    sent_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: order_items; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    product_id uuid,
    product_name_snapshot character varying(200) NOT NULL,
    unit_price_satang bigint NOT NULL,
    quantity integer NOT NULL,
    line_total_satang bigint NOT NULL,
    special_instructions text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    CONSTRAINT order_items_check CHECK ((line_total_satang = (unit_price_satang * quantity))),
    CONSTRAINT order_items_check1 CHECK (((deleted_at IS NULL) = (deleted_by IS NULL))),
    CONSTRAINT order_items_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT order_items_unit_price_satang_check CHECK ((unit_price_satang >= 0))
);


--
-- Name: order_status_history; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.order_status_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    from_status character varying(30),
    to_status character varying(30) NOT NULL,
    changed_by uuid,
    note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT order_status_history_from_status_check CHECK (((from_status)::text = ANY ((ARRAY['pending_payment'::character varying, 'pending_verification'::character varying, 'confirmed'::character varying, 'preparing'::character varying, 'ready_for_pickup'::character varying, 'shipped'::character varying, 'completed'::character varying, 'cancelled'::character varying, 'expired'::character varying])::text[]))),
    CONSTRAINT order_status_history_to_status_check CHECK (((to_status)::text = ANY ((ARRAY['pending_payment'::character varying, 'pending_verification'::character varying, 'confirmed'::character varying, 'preparing'::character varying, 'ready_for_pickup'::character varying, 'shipped'::character varying, 'completed'::character varying, 'cancelled'::character varying, 'expired'::character varying])::text[])))
);


--
-- Name: orders; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_number character varying(40) NOT NULL,
    idempotency_key character varying(100) NOT NULL,
    customer_id uuid,
    customer_name_snapshot character varying(160) NOT NULL,
    customer_phone_snapshot character varying(40) NOT NULL,
    fulfillment_type character varying(20) NOT NULL,
    delivery_address_snapshot jsonb,
    delivery_note text,
    pickup_at timestamp with time zone,
    subtotal_satang bigint NOT NULL,
    delivery_fee_satang bigint DEFAULT 0 NOT NULL,
    discount_satang bigint DEFAULT 0 NOT NULL,
    total_satang bigint NOT NULL,
    currency character(3) DEFAULT 'THB'::bpchar NOT NULL,
    status character varying(30) DEFAULT 'pending_payment'::character varying NOT NULL,
    internal_note text,
    placed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    request_hash character varying(64),
    CONSTRAINT orders_check CHECK ((total_satang = ((subtotal_satang + delivery_fee_satang) - discount_satang))),
    CONSTRAINT orders_check1 CHECK (((((fulfillment_type)::text = 'delivery'::text) AND (delivery_address_snapshot IS NOT NULL) AND (pickup_at IS NULL)) OR (((fulfillment_type)::text = 'pickup'::text) AND (pickup_at IS NOT NULL) AND (delivery_address_snapshot IS NULL)))),
    CONSTRAINT orders_check2 CHECK (((deleted_at IS NULL) = (deleted_by IS NULL))),
    CONSTRAINT orders_currency_check CHECK ((currency = 'THB'::bpchar)),
    CONSTRAINT orders_delivery_fee_satang_check CHECK ((delivery_fee_satang >= 0)),
    CONSTRAINT orders_discount_satang_check CHECK ((discount_satang >= 0)),
    CONSTRAINT orders_fulfillment_type_check CHECK (((fulfillment_type)::text = ANY ((ARRAY['delivery'::character varying, 'pickup'::character varying])::text[]))),
    CONSTRAINT orders_status_check CHECK (((status)::text = ANY ((ARRAY['pending_payment'::character varying, 'pending_verification'::character varying, 'confirmed'::character varying, 'preparing'::character varying, 'ready_for_pickup'::character varying, 'shipped'::character varying, 'completed'::character varying, 'cancelled'::character varying, 'expired'::character varying])::text[]))),
    CONSTRAINT orders_subtotal_satang_check CHECK ((subtotal_satang >= 0)),
    CONSTRAINT orders_total_satang_check CHECK ((total_satang >= 0))
);


--
-- Name: payment_evidence; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.payment_evidence (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    payment_id uuid NOT NULL,
    object_key character varying(500) NOT NULL,
    mime_type character varying(100) NOT NULL,
    file_size_bytes integer NOT NULL,
    checksum character varying(128) NOT NULL,
    submitted_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    content bytea,
    CONSTRAINT payment_evidence_check CHECK (((deleted_at IS NULL) = (deleted_by IS NULL))),
    CONSTRAINT payment_evidence_file_size_bytes_check CHECK ((file_size_bytes > 0)),
    CONSTRAINT payment_evidence_mime_type_check CHECK (((mime_type)::text = ANY ((ARRAY['image/jpeg'::character varying, 'image/png'::character varying, 'image/webp'::character varying])::text[])))
);


--
-- Name: payments; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    method character varying(30) NOT NULL,
    expected_amount_satang bigint NOT NULL,
    paid_amount_satang bigint,
    status character varying(30) DEFAULT 'pending'::character varying NOT NULL,
    verified_by uuid,
    verified_at timestamp with time zone,
    rejection_reason text,
    idempotency_key character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    CONSTRAINT payments_check CHECK ((((status)::text <> 'approved'::text) OR ((verified_by IS NOT NULL) AND (verified_at IS NOT NULL) AND (paid_amount_satang IS NOT NULL) AND (paid_amount_satang = expected_amount_satang)))),
    CONSTRAINT payments_check1 CHECK ((((status)::text <> 'rejected'::text) OR (NULLIF(btrim(rejection_reason), ''::text) IS NOT NULL))),
    CONSTRAINT payments_check2 CHECK (((deleted_at IS NULL) = (deleted_by IS NULL))),
    CONSTRAINT payments_expected_amount_satang_check CHECK ((expected_amount_satang >= 0)),
    CONSTRAINT payments_method_check CHECK (((method)::text = ANY ((ARRAY['bank_transfer'::character varying, 'promptpay'::character varying])::text[]))),
    CONSTRAINT payments_paid_amount_satang_check CHECK ((paid_amount_satang >= 0)),
    CONSTRAINT payments_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'submitted'::character varying, 'approved'::character varying, 'rejected'::character varying, 'needs_resubmission'::character varying, 'expired'::character varying])::text[])))
);


--
-- Name: product_images; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.product_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    object_key character varying(500) NOT NULL,
    mime_type character varying(100) NOT NULL,
    alt_text character varying(255),
    sort_order integer DEFAULT 0 NOT NULL,
    is_primary boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    content bytea,
    CONSTRAINT product_images_check CHECK (((deleted_at IS NULL) = (deleted_by IS NULL)))
);


--
-- Name: products; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    category_id uuid NOT NULL,
    name character varying(200) NOT NULL,
    slug character varying(220),
    description_th text NOT NULL,
    price_satang bigint NOT NULL,
    compare_at_price_satang bigint,
    stock_quantity integer NOT NULL,
    is_active boolean NOT NULL,
    is_featured boolean NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    CONSTRAINT products_compare_price_check CHECK (((compare_at_price_satang IS NULL) OR (compare_at_price_satang >= price_satang))),
    CONSTRAINT products_price_check CHECK ((price_satang >= 0)),
    CONSTRAINT products_soft_delete_check CHECK (((deleted_at IS NULL) = (deleted_by IS NULL))),
    CONSTRAINT products_stock_check CHECK ((stock_quantity >= 0))
);


--
-- Name: store_settings; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.store_settings (
    key character varying(80) NOT NULL,
    value jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: app_schema; Owner: -
--

CREATE TABLE app_schema.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255),
    role character varying(20) DEFAULT 'customer'::character varying NOT NULL,
    auth_provider character varying(30),
    provider_subject character varying(255),
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    CONSTRAINT users_auth_provider_check CHECK (((auth_provider)::text = ANY ((ARRAY['line'::character varying, 'email'::character varying, 'google'::character varying])::text[]))),
    CONSTRAINT users_check CHECK (((deleted_at IS NULL) = (deleted_by IS NULL))),
    CONSTRAINT users_role_check CHECK (((role)::text = ANY ((ARRAY['customer'::character varying, 'staff'::character varying, 'admin'::character varying])::text[])))
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(128) NOT NULL
);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: categories categories_slug_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.categories
    ADD CONSTRAINT categories_slug_key UNIQUE (slug);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: customers customers_user_id_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.customers
    ADD CONSTRAINT customers_user_id_key UNIQUE (user_id);


--
-- Name: inventory_movements inventory_movements_idempotency_key_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_movements
    ADD CONSTRAINT inventory_movements_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: inventory_movements inventory_movements_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_movements
    ADD CONSTRAINT inventory_movements_pkey PRIMARY KEY (id);


--
-- Name: inventory_reservations inventory_reservations_order_id_product_id_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_reservations
    ADD CONSTRAINT inventory_reservations_order_id_product_id_key UNIQUE (order_id, product_id);


--
-- Name: inventory_reservations inventory_reservations_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_reservations
    ADD CONSTRAINT inventory_reservations_pkey PRIMARY KEY (id);


--
-- Name: notification_outbox notification_outbox_order_id_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.notification_outbox
    ADD CONSTRAINT notification_outbox_order_id_key UNIQUE (order_id);


--
-- Name: notification_outbox notification_outbox_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.notification_outbox
    ADD CONSTRAINT notification_outbox_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: order_status_history order_status_history_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_status_history
    ADD CONSTRAINT order_status_history_pkey PRIMARY KEY (id);


--
-- Name: orders orders_idempotency_key_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.orders
    ADD CONSTRAINT orders_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: orders orders_order_number_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.orders
    ADD CONSTRAINT orders_order_number_key UNIQUE (order_number);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: payment_evidence payment_evidence_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payment_evidence
    ADD CONSTRAINT payment_evidence_pkey PRIMARY KEY (id);


--
-- Name: payments payments_idempotency_key_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payments
    ADD CONSTRAINT payments_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: product_images product_images_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: products products_slug_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.products
    ADD CONSTRAINT products_slug_key UNIQUE (slug);


--
-- Name: store_settings store_settings_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.store_settings
    ADD CONSTRAINT store_settings_pkey PRIMARY KEY (key);


--
-- Name: users users_auth_provider_provider_subject_key; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.users
    ADD CONSTRAINT users_auth_provider_provider_subject_key UNIQUE (auth_provider, provider_subject);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: ix_audit_entity_created; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_audit_entity_created ON app_schema.audit_logs USING btree (entity_type, entity_id, created_at DESC);


--
-- Name: ix_history_order_created; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_history_order_created ON app_schema.order_status_history USING btree (order_id, created_at DESC);


--
-- Name: ix_inventory_product_created; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_inventory_product_created ON app_schema.inventory_movements USING btree (product_id, created_at DESC);


--
-- Name: ix_order_items_order; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_order_items_order ON app_schema.order_items USING btree (order_id);


--
-- Name: ix_order_items_product; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_order_items_product ON app_schema.order_items USING btree (product_id);


--
-- Name: ix_orders_customer_created; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_orders_customer_created ON app_schema.orders USING btree (customer_id, created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: ix_orders_status_created; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_orders_status_created ON app_schema.orders USING btree (status, created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: ix_payment_evidence_payment; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_payment_evidence_payment ON app_schema.payment_evidence USING btree (payment_id);


--
-- Name: ix_payments_order; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_payments_order ON app_schema.payments USING btree (order_id);


--
-- Name: ix_payments_status_created; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_payments_status_created ON app_schema.payments USING btree (status, created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: ix_products_category_active; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_products_category_active ON app_schema.products USING btree (category_id, is_active, is_featured) WHERE (deleted_at IS NULL);


--
-- Name: ix_reservations_expiry; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_reservations_expiry ON app_schema.inventory_reservations USING btree (expires_at) WHERE (((status)::text = 'reserved'::text) AND (deleted_at IS NULL));


--
-- Name: ix_reservations_product; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE INDEX ix_reservations_product ON app_schema.inventory_reservations USING btree (product_id) WHERE (((status)::text = 'reserved'::text) AND (deleted_at IS NULL));


--
-- Name: ux_payments_one_approved; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE UNIQUE INDEX ux_payments_one_approved ON app_schema.payments USING btree (order_id) WHERE ((status)::text = 'approved'::text);


--
-- Name: ux_product_one_primary_image; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE UNIQUE INDEX ux_product_one_primary_image ON app_schema.product_images USING btree (product_id) WHERE (is_primary AND (deleted_at IS NULL));


--
-- Name: ux_users_email; Type: INDEX; Schema: app_schema; Owner: -
--

CREATE UNIQUE INDEX ux_users_email ON app_schema.users USING btree (lower((email)::text)) WHERE (email IS NOT NULL);


--
-- Name: categories categories_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER categories_updated_at BEFORE UPDATE ON app_schema.categories FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: customers customers_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER customers_updated_at BEFORE UPDATE ON app_schema.customers FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: inventory_reservations inventory_reservations_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER inventory_reservations_updated_at BEFORE UPDATE ON app_schema.inventory_reservations FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: order_items order_items_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER order_items_updated_at BEFORE UPDATE ON app_schema.order_items FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: orders orders_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER orders_updated_at BEFORE UPDATE ON app_schema.orders FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: payment_evidence payment_evidence_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER payment_evidence_updated_at BEFORE UPDATE ON app_schema.payment_evidence FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: payments payments_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER payments_updated_at BEFORE UPDATE ON app_schema.payments FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: product_images product_images_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER product_images_updated_at BEFORE UPDATE ON app_schema.product_images FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: products products_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER products_updated_at BEFORE UPDATE ON app_schema.products FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: users users_updated_at; Type: TRIGGER; Schema: app_schema; Owner: -
--

CREATE TRIGGER users_updated_at BEFORE UPDATE ON app_schema.users FOR EACH ROW EXECUTE FUNCTION app_schema.touch_updated_at();


--
-- Name: audit_logs audit_logs_actor_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.audit_logs
    ADD CONSTRAINT audit_logs_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES app_schema.users(id);


--
-- Name: categories categories_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.categories
    ADD CONSTRAINT categories_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: categories categories_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.categories
    ADD CONSTRAINT categories_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: categories categories_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.categories
    ADD CONSTRAINT categories_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: customers customers_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.customers
    ADD CONSTRAINT customers_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: customers customers_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.customers
    ADD CONSTRAINT customers_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: customers customers_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.customers
    ADD CONSTRAINT customers_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: customers customers_user_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.customers
    ADD CONSTRAINT customers_user_id_fkey FOREIGN KEY (user_id) REFERENCES app_schema.users(id);


--
-- Name: inventory_movements inventory_movements_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_movements
    ADD CONSTRAINT inventory_movements_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: inventory_movements inventory_movements_product_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_movements
    ADD CONSTRAINT inventory_movements_product_id_fkey FOREIGN KEY (product_id) REFERENCES app_schema.products(id);


--
-- Name: inventory_reservations inventory_reservations_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_reservations
    ADD CONSTRAINT inventory_reservations_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: inventory_reservations inventory_reservations_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_reservations
    ADD CONSTRAINT inventory_reservations_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: inventory_reservations inventory_reservations_order_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_reservations
    ADD CONSTRAINT inventory_reservations_order_id_fkey FOREIGN KEY (order_id) REFERENCES app_schema.orders(id);


--
-- Name: inventory_reservations inventory_reservations_product_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_reservations
    ADD CONSTRAINT inventory_reservations_product_id_fkey FOREIGN KEY (product_id) REFERENCES app_schema.products(id);


--
-- Name: inventory_reservations inventory_reservations_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.inventory_reservations
    ADD CONSTRAINT inventory_reservations_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: notification_outbox notification_outbox_order_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.notification_outbox
    ADD CONSTRAINT notification_outbox_order_id_fkey FOREIGN KEY (order_id) REFERENCES app_schema.orders(id);


--
-- Name: order_items order_items_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_items
    ADD CONSTRAINT order_items_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: order_items order_items_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_items
    ADD CONSTRAINT order_items_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: order_items order_items_order_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES app_schema.orders(id);


--
-- Name: order_items order_items_product_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_items
    ADD CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES app_schema.products(id);


--
-- Name: order_items order_items_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_items
    ADD CONSTRAINT order_items_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: order_status_history order_status_history_changed_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_status_history
    ADD CONSTRAINT order_status_history_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES app_schema.users(id);


--
-- Name: order_status_history order_status_history_order_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.order_status_history
    ADD CONSTRAINT order_status_history_order_id_fkey FOREIGN KEY (order_id) REFERENCES app_schema.orders(id);


--
-- Name: orders orders_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.orders
    ADD CONSTRAINT orders_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: orders orders_customer_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.orders
    ADD CONSTRAINT orders_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES app_schema.customers(id);


--
-- Name: orders orders_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.orders
    ADD CONSTRAINT orders_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: orders orders_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.orders
    ADD CONSTRAINT orders_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: payment_evidence payment_evidence_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payment_evidence
    ADD CONSTRAINT payment_evidence_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: payment_evidence payment_evidence_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payment_evidence
    ADD CONSTRAINT payment_evidence_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: payment_evidence payment_evidence_payment_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payment_evidence
    ADD CONSTRAINT payment_evidence_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES app_schema.payments(id);


--
-- Name: payment_evidence payment_evidence_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payment_evidence
    ADD CONSTRAINT payment_evidence_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: payments payments_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payments
    ADD CONSTRAINT payments_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: payments payments_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payments
    ADD CONSTRAINT payments_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: payments payments_order_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payments
    ADD CONSTRAINT payments_order_id_fkey FOREIGN KEY (order_id) REFERENCES app_schema.orders(id);


--
-- Name: payments payments_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payments
    ADD CONSTRAINT payments_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: payments payments_verified_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.payments
    ADD CONSTRAINT payments_verified_by_fkey FOREIGN KEY (verified_by) REFERENCES app_schema.users(id);


--
-- Name: product_images product_images_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.product_images
    ADD CONSTRAINT product_images_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: product_images product_images_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.product_images
    ADD CONSTRAINT product_images_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: product_images product_images_product_id_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.product_images
    ADD CONSTRAINT product_images_product_id_fkey FOREIGN KEY (product_id) REFERENCES app_schema.products(id);


--
-- Name: product_images product_images_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.product_images
    ADD CONSTRAINT product_images_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: products products_category_fk; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.products
    ADD CONSTRAINT products_category_fk FOREIGN KEY (category_id) REFERENCES app_schema.categories(id);


--
-- Name: products products_created_by_fk; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.products
    ADD CONSTRAINT products_created_by_fk FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: products products_deleted_by_fk; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.products
    ADD CONSTRAINT products_deleted_by_fk FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: products products_updated_by_fk; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.products
    ADD CONSTRAINT products_updated_by_fk FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- Name: users users_created_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.users
    ADD CONSTRAINT users_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_schema.users(id);


--
-- Name: users users_deleted_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.users
    ADD CONSTRAINT users_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES app_schema.users(id);


--
-- Name: users users_updated_by_fkey; Type: FK CONSTRAINT; Schema: app_schema; Owner: -
--

ALTER TABLE ONLY app_schema.users
    ADD CONSTRAINT users_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES app_schema.users(id);


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20261010023550'),
    ('20261010024134'),
    ('20261010024942'),
    ('20261010030000'),
    ('20261010120000'),
    ('20261013040000');
