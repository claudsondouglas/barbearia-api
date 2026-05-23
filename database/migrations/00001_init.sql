-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id         bigserial    PRIMARY KEY,
  name       text         NOT NULL,
  email      text         NOT NULL UNIQUE,
  password   text         NOT NULL,
  created_at timestamptz  NOT NULL DEFAULT now(),
  updated_at timestamptz  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS password_reset_otps (
  id         bigserial    PRIMARY KEY,
  email      text         NOT NULL,
  code       text         NOT NULL,
  expires_at timestamptz  NOT NULL,
  used       boolean      NOT NULL DEFAULT false,
  created_at timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_otps_email ON password_reset_otps (email);

-- +goose Down
DROP TABLE IF EXISTS password_reset_otps;
DROP TABLE IF EXISTS users;
