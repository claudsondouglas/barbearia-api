package actions

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestDeleteCustomer_NoActiveSchedules verifica que customer sem schedules ativos é removido com sucesso.
func TestDeleteCustomer_NoActiveSchedules(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(2, 1, "Cliente Manual", "11999990099"))

	// Sem schedules ativos (pending/confirmed)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "schedules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "status"}))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "customers"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := Delete(db, "barbearia-test", 2, 10, "owner")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

// TestDeleteCustomer_WithPendingSchedules verifica que customer com pending schedule retorna ErrCustomerHasActiveSchedules.
func TestDeleteCustomer_WithPendingSchedules(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(1, 1, "Carlos Souza", "11999990004"))

	// Schedule ativo com status=pending
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "schedules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "status"}).
			AddRow(1, 1, "pending"))

	err := Delete(db, "barbearia-test", 1, 10, "owner")
	if !errors.Is(err, ErrCustomerHasActiveSchedules) {
		t.Errorf("expected ErrCustomerHasActiveSchedules, got: %v", err)
	}
}

// TestDeleteCustomer_WithCompletedSchedules verifica que schedules completed não bloqueiam o delete.
func TestDeleteCustomer_WithCompletedSchedules(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(1, 1, "Carlos Souza", "11999990004"))

	// Sem schedules ativos (apenas completed/cancelled existem mas não bloqueiam)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "schedules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "status"}))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "customers"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := Delete(db, "barbearia-test", 1, 10, "owner")
	if err != nil {
		t.Fatalf("expected nil error for completed schedules, got: %v", err)
	}
}

// TestDeleteCustomer_NotFound verifica que ID inexistente retorna ErrCustomerNotFound.
func TestDeleteCustomer_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	err := Delete(db, "barbearia-test", 999, 10, "owner")
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Errorf("expected ErrCustomerNotFound, got: %v", err)
	}
}
