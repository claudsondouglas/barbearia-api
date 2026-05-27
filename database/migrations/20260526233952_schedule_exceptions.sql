-- +goose Up
CREATE TABLE IF NOT EXISTS schedule_exceptions (
  id              bigserial    PRIMARY KEY,
  organization_id bigint       NOT NULL REFERENCES organizations(id),
  user_id         bigint       REFERENCES users(id),
  date            date         NOT NULL,
  closed          boolean      NOT NULL DEFAULT true,
  open_time       text,
  close_time      text,
  reason          text,
  created_at      timestamptz  NOT NULL DEFAULT now(),
  updated_at      timestamptz  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_schedule_exceptions_org_date
  ON schedule_exceptions (organization_id, date)
  WHERE user_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_schedule_exceptions_member_date
  ON schedule_exceptions (organization_id, user_id, date)
  WHERE user_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS schedule_exceptions;
