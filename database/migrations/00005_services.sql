-- +goose Up
CREATE TABLE IF NOT EXISTS services (
  id              bigserial        PRIMARY KEY,
  organization_id bigint           NOT NULL REFERENCES organizations(id),
  name            text             NOT NULL,
  description     text,
  price           numeric(10,2)    NOT NULL,
  duration_min    int              NOT NULL,
  active          boolean          NOT NULL DEFAULT true,
  created_at      timestamptz      NOT NULL DEFAULT now(),
  updated_at      timestamptz      NOT NULL DEFAULT now(),
  deleted_at      timestamptz
);

-- +goose Down
DROP TABLE IF EXISTS services;
