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

// scheduleRows retorna um *sqlmock.Rows com colunas básicas de um agendamento.
func scheduleRows(id uint, orgID, serviceID, professionalID uint, status string) *sqlmock.Rows {
	now := time.Now()
	scheduledAt := now.Add(24 * time.Hour)
	endsAt := scheduledAt.Add(30 * time.Minute)
	return sqlmock.NewRows([]string{
		"id", "organization_id", "service_id", "professional_id", "client_id", "customer_id",
		"status", "scheduled_at", "ends_at", "price_snapshot", "duration_min_snapshot",
		"notes", "original_scheduled_at", "rescheduled_at", "rescheduled_by",
		"confirmed_at", "confirmed_by", "cancelled_at", "cancelled_by",
		"completed_at", "completed_by", "no_show_at", "no_show_by",
		"created_at", "updated_at",
	}).AddRow(
		id, orgID, serviceID, professionalID, nil, nil,
		status, scheduledAt, endsAt, 30.00, 30,
		nil, nil, nil, nil,
		nil, nil, nil, nil,
		nil, nil, nil, nil,
		now, now,
	)
}

// orgRows retorna um *sqlmock.Rows com colunas básicas de uma organização.
func orgRows(id uint, slug string, ownerID uint) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "timezone", "created_at", "deleted_at"}).
		AddRow(id, slug, ownerID, "Barbearia Test", "11999990001", "org@test.com", "America/Sao_Paulo", time.Now(), nil)
}
