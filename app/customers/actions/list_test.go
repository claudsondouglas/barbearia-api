package actions

import (
	"errors"
	"regexp"
	"testing"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestListCustomers_Success verifica que owner lista customers com sucesso.
func TestListCustomers_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(
			customerRows(1, 1, "Carlos Souza", "11999990004").
				AddRow(2, 1, nil, "Cliente Manual", "11999990099", nil, nil, nil),
		)

	customers, err := List(db, "barbearia-test", 10, "owner", models.ListCustomersFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(customers) == 0 {
		t.Error("expected at least one customer")
	}
}

// TestListCustomers_Empty verifica que org sem customers retorna slice vazio sem erro.
func TestListCustomers_Empty(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "customers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	customers, err := List(db, "barbearia-test", 10, "owner", models.ListCustomersFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if customers == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(customers) != 0 {
		t.Errorf("expected 0 customers, got %d", len(customers))
	}
}

// TestListCustomers_Search_ByName verifica que busca por nome gera query com ILIKE.
func TestListCustomers_Search_ByName(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	// Espera query com ILIKE para busca por nome
	mock.ExpectQuery(`ILIKE`).
		WillReturnRows(customerRows(1, 1, "Carlos Souza", "11999990004"))

	customers, err := List(db, "barbearia-test", 10, "owner", models.ListCustomersFilter{Query: "carlos", Limit: 20})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(customers) == 0 {
		t.Error("expected at least one customer matching 'carlos'")
	}
}

// TestListCustomers_Forbidden verifica que usuário comum recebe ErrForbidden.
func TestListCustomers_Forbidden(t *testing.T) {
	db, _ := newTestDB(t)

	_, err := List(db, "barbearia-test", 20, "user", models.ListCustomersFilter{})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

// TestListCustomers_OrgNotFound verifica que slug inexistente retorna ErrOrgNotFound.
func TestListCustomers_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nao-existe").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id"}))

	_, err := List(db, "nao-existe", 10, "owner", models.ListCustomersFilter{})
	if !errors.Is(err, ErrOrgNotFound) {
		t.Errorf("expected ErrOrgNotFound, got: %v", err)
	}
}
