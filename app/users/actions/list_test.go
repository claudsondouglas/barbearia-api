package actions

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestList_ReturnsAll verifica que List retorna todos os usuários cadastrados.
func TestList_ReturnsAll(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Alice", "alice@test.com", "hashed", "user", time.Now(), time.Now()).
			AddRow(2, "Bob", "bob@test.com", "hashed", "admin", time.Now(), time.Now()))

	users, err := List(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestList_Empty verifica que List retorna slice vazio sem erro quando não há usuários.
func TestList_Empty(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}))

	users, err := List(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty slice, got %d users", len(users))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
