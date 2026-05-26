package actions

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestAutoCreate_NewCustomer verifica que AutoCreate cria customer quando não existe.
func TestAutoCreate_NewCustomer(t *testing.T) {
	db, mock := newTestDB(t)

	// Verifica se já existe customer com user_id na org
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	customer, err := AutoCreate(db, 1, 22, "Pedro", "11999990005")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if customer == nil {
		t.Fatal("expected customer, got nil")
	}
	if customer.Name != "Pedro" {
		t.Errorf("expected name=Pedro, got: %s", customer.Name)
	}
	if customer.Phone != "11999990005" {
		t.Errorf("expected phone=11999990005, got: %s", customer.Phone)
	}
}

// TestAutoCreate_ExistingCustomer verifica que AutoCreate retorna existente sem criar novo.
func TestAutoCreate_ExistingCustomer(t *testing.T) {
	db, mock := newTestDB(t)

	// Customer já existe com user_id=20
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(1, 1, "Carlos Souza", "11999990004"))

	customer, err := AutoCreate(db, 1, 20, "Carlos Souza", "11999990004")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if customer == nil {
		t.Fatal("expected existing customer, got nil")
	}
	if customer.ID != 1 {
		t.Errorf("expected existing customer id=1, got: %d", customer.ID)
	}
}

// TestAutoCreate_UserWithoutPhone verifica que phone vazio resulta em nil, nil.
func TestAutoCreate_UserWithoutPhone(t *testing.T) {
	db, _ := newTestDB(t)

	customer, err := AutoCreate(db, 1, 21, "Ana Lima", "")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if customer != nil {
		t.Errorf("expected nil customer for user without phone, got: %+v", customer)
	}
}
