package actions

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestFindCustomer_Success verifica que owner encontra customer existente.
func TestFindCustomer_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(1, 1, "Carlos Souza", "11999990004"))

	customer, err := Find(db, "barbearia-test", 1, 10, "user")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if customer == nil {
		t.Fatal("expected customer, got nil")
	}
}

// TestFindCustomer_NotFound verifica que ID inexistente retorna ErrCustomerNotFound.
func TestFindCustomer_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	_, err := Find(db, "barbearia-test", 999, 10, "user")
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Errorf("expected ErrCustomerNotFound, got: %v", err)
	}
}

// TestFindCustomer_WrongOrg verifica que customer de outra org retorna ErrCustomerNotFound.
func TestFindCustomer_WrongOrg(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	// Customer ID=3 pertence à org 2, não à org 1 — query com org_id=1 não retorna nada
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	_, err := Find(db, "barbearia-test", 3, 10, "user")
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Errorf("expected ErrCustomerNotFound, got: %v", err)
	}
}

// TestFindCustomer_Forbidden verifica que usuário que não é owner recebe ErrForbidden.
func TestFindCustomer_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)

	// org com owner_id=10; requestingUserID=20 (não é owner)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	_, err := Find(db, "barbearia-test", 1, 20, "user")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}
