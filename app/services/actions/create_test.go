package actions

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestCreateService_Success verifica que owner pode criar um serviço com campos válidos.
func TestCreateService_Success(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca org
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	// Verifica se solicitante é owner/admin
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(userRows(10, "Owner", "owner@test.com", "user"))

	// INSERT do serviço
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "services"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))
	mock.ExpectCommit()

	input := models.CreateServiceInput{
		Name:        "Corte",
		Price:       30.00,
		DurationMin: 30,
	}

	svc, err := Create(db, "barbearia-test", 10, input)
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

// TestCreateService_PriceZero verifica que price=0 retorna ErrInvalidPrice.
func TestCreateService_PriceZero(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(userRows(10, "Owner", "owner@test.com", "user"))

	input := models.CreateServiceInput{
		Name:        "Corte",
		Price:       0,
		DurationMin: 30,
	}

	_, err := Create(db, "barbearia-test", 10, input)
	if !errors.Is(err, ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got: %v", err)
	}
}

// TestCreateService_DurationAboveMax verifica que duration_min > 240 retorna ErrInvalidDuration.
func TestCreateService_DurationAboveMax(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(userRows(10, "Owner", "owner@test.com", "user"))

	input := models.CreateServiceInput{
		Name:        "Corte",
		Price:       30.00,
		DurationMin: 241,
	}

	_, err := Create(db, "barbearia-test", 10, input)
	if !errors.Is(err, ErrInvalidDuration) {
		t.Errorf("expected ErrInvalidDuration, got: %v", err)
	}
}

// TestCreateService_Forbidden verifica que usuário comum recebe ErrForbidden.
func TestCreateService_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)

	// org owner é user 10, solicitante é user 12 (sem vínculo)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(12)).
		WillReturnRows(userRows(12, "Regular", "regular@test.com", "user"))

	input := models.CreateServiceInput{
		Name:        "Corte",
		Price:       30.00,
		DurationMin: 30,
	}

	_, err := Create(db, "barbearia-test", 12, input)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

// TestCreateService_OrgNotFound verifica que slug inválido retorna ErrOrgNotFound.
func TestCreateService_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}))

	input := models.CreateServiceInput{
		Name:        "Corte",
		Price:       30.00,
		DurationMin: 30,
	}

	_, err := Create(db, "nonexistent", 10, input)
	if !errors.Is(err, ErrOrgNotFound) {
		t.Errorf("expected ErrOrgNotFound, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
