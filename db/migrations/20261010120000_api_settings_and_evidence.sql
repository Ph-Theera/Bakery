-- migrate:up
CREATE TABLE app_schema.store_settings (
  key varchar(80) PRIMARY KEY,
  value jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO app_schema.store_settings(key,value) VALUES ('store','{"store_name":"TongBai","description":"ขนมโฮมเมดอบสดใหม่ทุกวัน","delivery_fee_thb":50,"free_delivery_threshold_thb":800,"accepting_orders":true,"line_oa_url":""}');
-- Private slip contents; never exposed as a public static URL.
ALTER TABLE app_schema.payment_evidence ADD COLUMN content bytea;
ALTER TABLE app_schema.orders ADD COLUMN request_hash varchar(64);
CREATE TABLE app_schema.notification_outbox (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id uuid NOT NULL UNIQUE REFERENCES app_schema.orders(id),
  payload text NOT NULL,
  attempts integer NOT NULL DEFAULT 0,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  sent_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- migrate:down
DROP TABLE app_schema.notification_outbox;
ALTER TABLE app_schema.orders DROP COLUMN request_hash;
ALTER TABLE app_schema.payment_evidence DROP COLUMN content;
DROP TABLE app_schema.store_settings;
