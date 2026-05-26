package actions

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestDeleteService_Success verifica que owner pode deletar (soft delete) um serviço sem agendamentos ativos.
func TestDeleteService_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(userRows(10, "Owner", "owner@test.com", "user"))

	// Serviço existe
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}).AddRow(1, 1, "Corte", nil, 30.00, 30, true, time.Now(), time.Now(), nil))

	// Sem agendamentos ativos
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "service_id", "status"}))

	// Soft delete
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "services"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := Delete(db, "barbearia-test", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestDeleteService_WithActiveSchedules verifica que serviço com agendamentos
// pending/confirmed retorna ErrServiceHasActiveSchedules.
func TestDeleteService_WithActiveSchedules(t *testing.T) {
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

	// Agendamento pendente encontrado
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "service_id", "status"}).
			AddRow(1, 1, "pending"))

	err := Delete(db, "barbearia-test", 1, 10)
	if !errors.Is(err, ErrServiceHasActiveSchedules) {
		t.Errorf("expected ErrServiceHasActiveSchedules, got: %v", err)
	}
}

// TestDeleteService_NotFound verifica que ID inexistente retorna ErrServiceNotFound.
func TestDeleteService_NotFound(t *testing.T) {
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

	err := Delete(db, "barbearia-test", 999, 10)
	if !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("expected ErrServiceNotFound, got: %v", err)
	}
}

// TestDeleteService_Forbidden verifica que usuário comum recebe ErrForbidden.
func TestDeleteService_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(orgRows(1, "barbearia-test", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(12)).
		WillReturnRows(userRows(12, "Regular", "regular@test.com", "user"))

	err := Delete(db, "barbearia-test", 1, 12)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}
