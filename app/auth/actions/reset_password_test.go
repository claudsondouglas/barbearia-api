package authactions

import (
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func otpRows(id uint, email, code string, expiresAt time.Time, used bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "email", "code", "expires_at", "used", "created_at"}).
		AddRow(id, email, code, expiresAt, used, time.Now())
}

func TestResetPassword_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "password_reset_otps" WHERE code = $1 AND used = false AND expires_at > $2 ORDER BY "password_reset_otps"."id" LIMIT $3`)).
		WithArgs("123456", sqlmock.AnyArg(), 1).
		WillReturnRows(otpRows(1, "user@test.com", "123456", time.Now().Add(10*time.Minute), false))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("user@test.com", 1).
		WillReturnRows(userRows(1, "user@test.com", "oldhash"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "password"=$1,"updated_at"=$2 WHERE "id" = $3`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "password_reset_otps" SET "used"=$1 WHERE "id" = $2`)).
		WithArgs(true, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := ResetPassword(db, ResetPasswordInput{
		Code:        "123456",
		NewPassword: "newpass123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

func TestResetPassword_InvalidOTP(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "password_reset_otps" WHERE code = $1 AND used = false AND expires_at > $2 ORDER BY "password_reset_otps"."id" LIMIT $3`)).
		WithArgs("000000", sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows([]string{}))

	err := ResetPassword(db, ResetPasswordInput{
		Code:        "000000",
		NewPassword: "newpass123",
	})
	if err != ErrInvalidOrExpiredOTP {
		t.Errorf("expected ErrInvalidOrExpiredOTP, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

func TestResetPassword_ExpiredOTP(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "password_reset_otps" WHERE code = $1 AND used = false AND expires_at > $2 ORDER BY "password_reset_otps"."id" LIMIT $3`)).
		WithArgs("123456", sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows([]string{}))

	err := ResetPassword(db, ResetPasswordInput{
		Code:        "123456",
		NewPassword: "newpass123",
	})
	if err != ErrInvalidOrExpiredOTP {
		t.Errorf("expected ErrInvalidOrExpiredOTP, got %v", err)
	}
}
