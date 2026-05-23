-- +goose Up
-- Migra coluna phone -> email em users (se ainda existir)
-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'users' AND column_name = 'phone'
  ) THEN
    IF NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_name = 'users' AND column_name = 'email'
    ) THEN
      ALTER TABLE users ADD COLUMN email text;
      UPDATE users SET email = phone;
    END IF;
    ALTER TABLE users ALTER COLUMN email SET NOT NULL;
    ALTER TABLE users DROP COLUMN phone;
  END IF;
END $$;
-- +goose StatementEnd

-- Migra coluna phone -> email em password_reset_otps (se ainda existir)
-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'password_reset_otps' AND column_name = 'phone'
  ) THEN
    TRUNCATE TABLE password_reset_otps;
    IF NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_name = 'password_reset_otps' AND column_name = 'email'
    ) THEN
      ALTER TABLE password_reset_otps ADD COLUMN email text NOT NULL DEFAULT '';
      ALTER TABLE password_reset_otps ALTER COLUMN email DROP DEFAULT;
    END IF;
    ALTER TABLE password_reset_otps DROP COLUMN phone;
  END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- Irreversível: não é possível recuperar dados de phone a partir de email
