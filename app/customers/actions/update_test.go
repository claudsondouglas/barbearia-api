package actions

import (
	"errors"
	"regexp"
	"testing"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestUpdateCustomer_Name verifica que atualização de nome funciona com sucesso.
func TestUpdateCustomer_Name(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(1, 1, "Carlos Souza", "11999990004"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "customers"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	newName := "Carlos Silva"
	input := models.UpdateCustomerInput{Name: &newName}
	customer, err := Update(db, "barbearia-test", 1, 10, "owner", input)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if customer == nil {
		t.Fatal("expected customer, got nil")
	}
}

// TestUpdateCustomer_PhoneDuplicate verifica que phone duplicado retorna ErrDuplicatePhone.
func TestUpdateCustomer_PhoneDuplicate(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	// Customer atual
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(1, 1, "Carlos Souza", "11999990004"))

	// Verifica duplicidade — outro customer com o mesmo phone
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(2, 1, "Cliente Manual", "11999990099"))

	newPhone := "11999990099"
	input := models.UpdateCustomerInput{Phone: &newPhone}
	_, err := Update(db, "barbearia-test", 1, 10, "owner", input)
	if !errors.Is(err, ErrDuplicatePhone) {
		t.Errorf("expected ErrDuplicatePhone, got: %v", err)
	}
}

// TestUpdateCustomer_EmptyBody verifica que body sem campos reconhecidos retorna erro.
func TestUpdateCustomer_EmptyBody(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(1, 1, "Carlos Souza", "11999990004"))

	input := models.UpdateCustomerInput{} // nenhum campo preenchido
	_, err := Update(db, "barbearia-test", 1, 10, "owner", input)
	if err == nil {
		t.Error("expected error for empty body, got nil")
	}
}

// TestUpdateCustomer_NotFound verifica que ID inexistente retorna ErrCustomerNotFound.
func TestUpdateCustomer_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	newName := "Novo Nome"
	input := models.UpdateCustomerInput{Name: &newName}
	_, err := Update(db, "barbearia-test", 999, 10, "owner", input)
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Errorf("expected ErrCustomerNotFound, got: %v", err)
	}
}
