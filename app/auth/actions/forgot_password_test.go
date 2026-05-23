package authactions

import (
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

type mockSender struct {
	called bool
	err    error
}

func (m *mockSender) SendEmail(to, subject, body string) error {
	m.called = true
	return m.err
}

func TestForgotPassword_UserNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("notfound@test.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{}))

	sender := &mockSender{}
	if err := ForgotPassword(db, "notfound@test.com", sender); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if sender.called {
		t.Error("sender should not be called when user does not exist")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

func TestForgotPassword_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("user@test.com", 1).
		WillReturnRows(userRows(1, "user@test.com", "anyhash"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "password_reset_otps" WHERE email = $1`)).
		WithArgs("user@test.com").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "password_reset_otps"`).
		WithArgs("user@test.com", sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "id"}).AddRow(time.Now(), 1))
	mock.ExpectCommit()

	sender := &mockSender{}
	if err := ForgotPassword(db, "user@test.com", sender); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !sender.called {
		t.Error("expected sender to be called")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

func TestForgotPassword_SenderError(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("user@test.com", 1).
		WillReturnRows(userRows(1, "user@test.com", "anyhash"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "password_reset_otps" WHERE email = $1`)).
		WithArgs("user@test.com").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "password_reset_otps"`).
		WithArgs("user@test.com", sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "id"}).AddRow(time.Now(), 1))
	mock.ExpectCommit()

	sender := &mockSender{err: errors.New("smtp error")}
	if err := ForgotPassword(db, "user@test.com", sender); err == nil {
		t.Error("expected error from sender, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
