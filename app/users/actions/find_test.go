package actions

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

// TestFind_Success verifica que Find retorna o usuário quando encontrado.
func TestFind_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Test User", "test@test.com", "hashed", "user", time.Now(), time.Now()))

	user, err := Find(db, "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.ID != 1 {
		t.Errorf("expected user.ID=1, got %d", user.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestFind_NotFound verifica que Find retorna gorm.ErrRecordNotFound quando o usuário não existe.
func TestFind_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("999", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}))

	_, err := Find(db, "999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
