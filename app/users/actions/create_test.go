package actions

import (
	"regexp"
	"testing"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
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

// TestCreateUser_Success verifica que Create retorna o user com ID populado.
func TestCreateUser_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	input := models.CreateUserInput{
		Name:     "Test User",
		Email:    "test@test.com",
		Password: "pass123",
	}

	user, err := Create(db, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.ID == 0 {
		t.Error("expected user.ID to be set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestCreateUser_DuplicateEmail verifica que Create retorna erro quando e-mail já existe.
func TestCreateUser_DuplicateEmail(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	input := models.CreateUserInput{
		Name:     "Test User",
		Email:    "duplicate@test.com",
		Password: "pass123",
	}

	_, err := Create(db, input)
	if err == nil {
		t.Error("expected error on duplicate email, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestCreateUser_PasswordIsHashed verifica que a senha armazenada é um hash bcrypt.
func TestCreateUser_PasswordIsHashed(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	input := models.CreateUserInput{
		Name:     "Test User",
		Email:    "hashed@test.com",
		Password: "plaintext",
	}

	user, err := Create(db, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("plaintext")); err != nil {
		t.Errorf("expected password to be bcrypt hash of 'plaintext', got: %s", user.Password)
	}
}
