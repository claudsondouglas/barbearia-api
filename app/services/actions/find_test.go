package actions

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestFindService_Public_Active verifica que usuário não autenticado consegue
// ver um serviço ativo.
func TestFindService_Public_Active(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}).AddRow(1, 1, "Corte", nil, 30.00, 30, true, time.Now(), time.Now(), nil))

	svc, err := Find(db, "barbearia-test", 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc == nil || svc.ID != 1 {
		t.Errorf("expected service id=1, got %v", svc)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestFindService_Public_Inactive verifica que usuário não autenticado recebe
// ErrServiceNotFound para serviço inativo.
func TestFindService_Public_Inactive(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	// Serviço existe mas está inativo — a query não retorna nada para usuário público
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}))

	_, err := Find(db, "barbearia-test", 2, nil)
	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("expected ErrServiceNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestFindService_NotFound verifica que ID inexistente retorna ErrServiceNotFound.
func TestFindService_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}))

	_, err := Find(db, "barbearia-test", 999, nil)
	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("expected ErrServiceNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
