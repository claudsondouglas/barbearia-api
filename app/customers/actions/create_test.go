package actions

import (
	"errors"
	"regexp"
	"testing"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestCreateCustomer_Owner_Success verifica que owner (role "user") cria customer com sucesso.
func TestCreateCustomer_Owner_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	// Verifica duplicidade de phone na org (nenhum encontrado)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "phone"}))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	input := models.CreateCustomerInput{Name: "Carlos Souza", Phone: "11999990004"}
	// userID=10 é o owner da org (orgRows retorna owner_id=10); role é "user" (padrão no banco)
	customer, err := Create(db, "barbearia-test", 10, "user", input)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if customer == nil {
		t.Fatal("expected customer, got nil")
	}
}

// TestCreateCustomer_DuplicatePhone verifica que phone duplicado na mesma org retorna ErrDuplicatePhone.
func TestCreateCustomer_DuplicatePhone(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	// Phone já existe na org
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(customerRows(1, 1, "Outro Cliente", "11999990004"))

	input := models.CreateCustomerInput{Name: "Novo Cliente", Phone: "11999990004"}
	_, err := Create(db, "barbearia-test", 10, "user", input)
	if !errors.Is(err, ErrDuplicatePhone) {
		t.Errorf("expected ErrDuplicatePhone, got: %v", err)
	}
}

// TestCreateCustomer_Forbidden verifica que usuário que não é owner nem admin recebe ErrForbidden.
func TestCreateCustomer_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)

	// org com owner_id=10; requestingUserID=20 (não é owner)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	input := models.CreateCustomerInput{Name: "Cliente", Phone: "11999990099"}
	_, err := Create(db, "barbearia-test", 20, "user", input)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

// TestCreateCustomer_OrgNotFound verifica que slug inexistente retorna ErrOrgNotFound.
func TestCreateCustomer_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nao-existe").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id"}))

	input := models.CreateCustomerInput{Name: "Cliente", Phone: "11999990099"}
	_, err := Create(db, "nao-existe", 10, "owner", input)
	if !errors.Is(err, ErrOrgNotFound) {
		t.Errorf("expected ErrOrgNotFound, got: %v", err)
	}
}

// TestCreateCustomer_UserIdIgnored verifica que user_id no body é ignorado (sempre nil no banco).
func TestCreateCustomer_UserIdIgnored(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "phone"}))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	// O input não tem campo user_id — mesmo que o JSON viesse com ele, Create não o usa
	input := models.CreateCustomerInput{Name: "Carlos", Phone: "11999990010"}
	customer, err := Create(db, "barbearia-test", 10, "user", input)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if customer == nil {
		t.Fatal("expected customer, got nil")
	}
	if customer.UserID != nil {
		t.Errorf("expected user_id=nil, got: %v", customer.UserID)
	}
}
