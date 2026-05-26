-- +goose Up
CREATE TABLE IF NOT EXISTS customers (
  id              bigserial    PRIMARY KEY,
  organization_id bigint       NOT NULL REFERENCES organizations(id),
  user_id         bigint       REFERENCES users(id),
  name            text         NOT NULL,
  phone           text         NOT NULL,
  notes           text,
  created_at      timestamptz  NOT NULL DEFAULT now(),
  updated_at      timestamptz  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_org_phone
  ON customers (organization_id, phone);

-- +goose Down
DROP TABLE IF EXISTS customers;
