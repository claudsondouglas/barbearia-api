-- +goose Up
CREATE TABLE IF NOT EXISTS organizations (
  id            bigserial        PRIMARY KEY,
  owner_id      bigint           NOT NULL REFERENCES users(id),
  name          text             NOT NULL,
  slug          text             NOT NULL UNIQUE,
  phone         text             NOT NULL,
  email         text             NOT NULL UNIQUE,
  description   text,
  logo_url      text,
  street        text             NOT NULL,
  number        text             NOT NULL,
  complement    text,
  neighborhood  text             NOT NULL,
  city          text             NOT NULL,
  state         text             NOT NULL,
  zip_code      text             NOT NULL,
  latitude      double precision,
  longitude     double precision,
  timezone      text             NOT NULL DEFAULT 'America/Sao_Paulo',
  deleted_at    timestamptz,
  created_at    timestamptz      NOT NULL DEFAULT now(),
  updated_at    timestamptz      NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS org_members (
  id              bigserial    PRIMARY KEY,
  organization_id bigint       NOT NULL REFERENCES organizations(id),
  user_id         bigint       NOT NULL REFERENCES users(id),
  created_at      timestamptz  NOT NULL DEFAULT now(),
  deleted_at      timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_org_members_active
  ON org_members (organization_id, user_id)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS member_business_hours (
  id            bigserial    PRIMARY KEY,
  org_member_id bigint       NOT NULL REFERENCES org_members(id),
  day_of_week   int          NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
  closed        boolean      NOT NULL DEFAULT true,
  open_time     text,
  close_time    text
);

-- +goose Down
DROP TABLE IF EXISTS member_business_hours;
DROP TABLE IF EXISTS org_members;
DROP TABLE IF EXISTS organizations;
