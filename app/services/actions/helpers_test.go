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

func orgRows(id uint, slug string, ownerID uint) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
		AddRow(id, slug, ownerID, time.Now(), nil)
}

func serviceRows(id uint, orgID uint, name string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "organization_id", "name", "description", "price", "duration_min",
		"active", "created_at", "updated_at", "deleted_at",
	}).AddRow(id, orgID, name, nil, 30.00, 30, true, time.Now(), time.Now(), nil)
}

func userRows(id uint, name, email, role string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
		AddRow(id, name, email, "hashed", role, time.Now(), time.Now())
}
