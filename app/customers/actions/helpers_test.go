package actions

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	return gormDB, mock
}

func customerRows(id uint, orgID uint, name, phone string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}).
		AddRow(id, orgID, nil, name, phone, nil, time.Now(), time.Now())
}

func orgRows(id uint, slug string, ownerID uint) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
		AddRow(id, slug, ownerID, "Org Name", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", time.Now(), time.Now(), nil)
}
