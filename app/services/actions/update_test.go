package actions

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestUpdateService_Name verifica que owner pode atualizar apenas o nome de um serviço.
func TestUpdateService_Name(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(userRows(10, "Owner", "owner@test.com", "user"))

	// Busca o serviço atual
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}).AddRow(1, 1, "Corte", nil, 30.00, 30, true, time.Now(), time.Now(), nil))

	// UPDATE
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "services"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	newName := "Corte Degradê"
	input := models.UpdateServiceInput{Name: &newName}

	svc, err := Update(db, "barbearia-test", 1, 10, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected service, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestUpdateService_PriceZero verifica que price=0 retorna ErrInvalidPrice.
func TestUpdateService_PriceZero(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(userRows(10, "Owner", "owner@test.com", "user"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}).AddRow(1, 1, "Corte", nil, 30.00, 30, true, time.Now(), time.Now(), nil))

	zeroPrice := 0.0
	input := models.UpdateServiceInput{Price: &zeroPrice}

	_, err := Update(db, "barbearia-test", 1, 10, input)
	if !errors.Is(err, ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got: %v", err)
	}
}

// TestUpdateService_DeletedService verifica que atualizar serviço deletado retorna ErrServiceDeleted.
func TestUpdateService_DeletedService(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(userRows(10, "Owner", "owner@test.com", "user"))

	deletedAt := time.Now().Add(-24 * time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}).AddRow(3, 1, "Sobrancelha", nil, 10.00, 10, false, time.Now(), time.Now(), deletedAt))

	newName := "Sobrancelha Nova"
	input := models.UpdateServiceInput{Name: &newName}

	_, err := Update(db, "barbearia-test", 3, 10, input)
	if !errors.Is(err, ErrServiceDeleted) {
		t.Errorf("expected ErrServiceDeleted, got: %v", err)
	}
}

// TestUpdateService_NotFound verifica que ID inexistente retorna ErrServiceNotFound.
func TestUpdateService_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(userRows(10, "Owner", "owner@test.com", "user"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}))

	newName := "Inexistente"
	input := models.UpdateServiceInput{Name: &newName}

	_, err := Update(db, "barbearia-test", 999, 10, input)
	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("expected ErrServiceNotFound, got: %v", err)
	}
}

// TestUpdateService_Forbidden verifica que usuário comum recebe ErrForbidden.
func TestUpdateService_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(12)).
		WillReturnRows(userRows(12, "Regular", "regular@test.com", "user"))

	newName := "Corte"
	input := models.UpdateServiceInput{Name: &newName}

	_, err := Update(db, "barbearia-test", 1, 12, input)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}
