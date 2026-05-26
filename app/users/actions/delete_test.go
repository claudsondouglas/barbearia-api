package actions

import (
	"regexp"
	"testing"
	"time"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestDelete_Success verifica que Delete remove o usuário sem erro.
func TestDelete_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &models.User{
		ID:        1,
		Name:      "Test",
		Email:     "test@test.com",
		Password:  "hashed",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := Delete(db, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestDelete_UserNotFound verifica que Delete retorna erro quando o usuário não existe no banco.
func TestDelete_UserNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(999).
		WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	user := &models.User{
		ID:        999,
		Name:      "Ghost",
		Email:     "ghost@test.com",
		Password:  "hashed",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := Delete(db, user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
