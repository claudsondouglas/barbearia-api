package authactions

import (
	"regexp"
	"testing"
	"time"

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

func userRows(id uint, email, passwordHash string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
		AddRow(id, "Test User", email, passwordHash, "user", time.Now(), time.Now())
}

func TestLogin_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	db, mock := newTestDB(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("user@test.com", 1).
		WillReturnRows(userRows(1, "user@test.com", string(hash)))

	tokens, err := Login(db, models.LoginInput{Email: "user@test.com", Password: "pass123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("expected both tokens to be non-empty")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("notfound@test.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{}))

	_, err := Login(db, models.LoginInput{Email: "notfound@test.com", Password: "pass123"})
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	db, mock := newTestDB(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.MinCost)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("user@test.com", 1).
		WillReturnRows(userRows(1, "user@test.com", string(hash)))

	_, err := Login(db, models.LoginInput{Email: "user@test.com", Password: "wrong-pass"})
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
