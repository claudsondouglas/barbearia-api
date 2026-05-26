package actions

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestListServices_Public_ReturnsActiveOnly verifica que usuário não autenticado
// recebe apenas serviços ativos de uma org existente.
func TestListServices_Public_ReturnsActiveOnly(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca org pelo slug
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	// Retorna 2 serviços ativos (o inativo não aparece)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}).
			AddRow(1, 1, "Corte", nil, 30.00, 30, true, time.Now(), time.Now(), nil).
			AddRow(3, 1, "Sobrancelha", nil, 10.00, 10, true, time.Now(), time.Now(), nil))

	services, err := List(db, "barbearia-test", nil, false, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}
	for _, s := range services {
		if !s.Active {
			t.Errorf("expected only active services, got inactive service id=%d", s.ID)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestListServices_Empty verifica que org sem serviços ativos retorna slice vazio sem erro.
func TestListServices_Empty(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}))

	services, err := List(db, "barbearia-test", nil, false, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(services) != 0 {
		t.Errorf("expected empty slice, got %d", len(services))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestListServices_OrgNotFound verifica que slug inexistente retorna ErrOrgNotFound.
func TestListServices_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}))

	_, err := List(db, "nonexistent", nil, false, 20, 0)
	if !errors.Is(err, ErrOrgNotFound) {
		t.Errorf("expected ErrOrgNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
