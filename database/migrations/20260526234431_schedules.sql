-- +goose Up
CREATE TABLE IF NOT EXISTS schedules (
  id                      bigserial        PRIMARY KEY,
  organization_id         bigint           NOT NULL REFERENCES organizations(id),
  service_id              bigint           NOT NULL REFERENCES services(id),
  professional_id         bigint           NOT NULL REFERENCES users(id),
  client_id               bigint           REFERENCES users(id),
  customer_id             bigint           REFERENCES customers(id),
  status                  text             NOT NULL DEFAULT 'pending',
  scheduled_at            timestamptz      NOT NULL,
  ends_at                 timestamptz      NOT NULL,
  price_snapshot          numeric(10,2)    NOT NULL,
  duration_min_snapshot   int              NOT NULL,
  notes                   text,
  original_scheduled_at   timestamptz,
  rescheduled_at          timestamptz,
  rescheduled_by          bigint           REFERENCES users(id),
  confirmed_at            timestamptz,
  confirmed_by            bigint           REFERENCES users(id),
  cancelled_at            timestamptz,
  cancelled_by            bigint           REFERENCES users(id),
  completed_at            timestamptz,
  completed_by            bigint           REFERENCES users(id),
  no_show_at              timestamptz,
  no_show_by              bigint           REFERENCES users(id),
  created_at              timestamptz      NOT NULL DEFAULT now(),
  updated_at              timestamptz      NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_schedules_professional_status
  ON schedules (professional_id, status);

CREATE INDEX IF NOT EXISTS idx_schedules_org
  ON schedules (organization_id);

-- +goose Down
DROP TABLE IF EXISTS schedules;
