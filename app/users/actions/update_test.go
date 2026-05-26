package actions

import (
	"regexp"
	"testing"
	"time"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func strPtr(s string) *string { return &s }

// TestUpdate_Name verifica que Update atualiza o name do usuário.
func TestUpdate_Name(t *testing.T) {
	db, mock := newTestDB(t)

	// Find chamado dentro de Update
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Old Name", "test@test.com", "hashed", "user", time.Now(), time.Now()))

	// db.Model(user).Updates(updates)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	newName := "New Name"
	user, err := Update(db, "1", models.UpdateUserInput{Name: &newName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestUpdate_NotFound verifica que Update retorna gorm.ErrRecordNotFound para ID inexistente.
func TestUpdate_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("999", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}))

	newName := "Any"
	_, err := Update(db, "999", models.UpdateUserInput{Name: &newName})
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

// TestUpdate_NoFields verifica que Update retorna ErrNoFieldsToUpdate quando nenhum campo é informado.
func TestUpdate_NoFields(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Test", "test@test.com", "hashed", "user", time.Now(), time.Now()))

	_, err := Update(db, "1", models.UpdateUserInput{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrNoFieldsToUpdate {
		t.Errorf("expected ErrNoFieldsToUpdate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
