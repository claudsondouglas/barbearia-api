package actions

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestRemove_Success verifica que Remove realiza o soft-delete do membro com sucesso.
func TestRemove_Success(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização (ownerID=10, requestingUser=10 é o owner)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Busca o usuário solicitante (owner)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(userRows(10, "Owner", "owner@test.com"))

	// Membro alvo ativo (targetUserID=20, diferente do owner=10)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(memberRows(50, 1, 20, nil))

	// Busca schedules ativos (retorna vazio — sem pendências)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "schedules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_member_id", "status"}))

	// DELETE dos business hours do membro
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "member_business_hours"`)).
		WillReturnResult(sqlmock.NewResult(0, 7))
	mock.ExpectCommit()

	// DELETE das schedule_exceptions do membro
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "schedule_exceptions"`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Soft-delete do org_member (UPDATE deleted_at)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "org_members"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := Remove(db, "barber-shop", 10, 20)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

// TestRemove_OwnerCannotBeRemoved verifica que o dono da org não pode ser removido.
func TestRemove_OwnerCannotBeRemoved(t *testing.T) {
	db, mock := newTestDB(t)

	// ownerID=10, targetUserID=10 (mesmo usuário)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	err := Remove(db, "barber-shop", 10, 10)
	if !errors.Is(err, ErrOwnerCannotBeRemoved) {
		t.Errorf("expected ErrOwnerCannotBeRemoved, got: %v", err)
	}
}

// TestRemove_MemberHasActiveSchedules verifica que membro com schedules pendentes não pode ser removido.
func TestRemove_MemberHasActiveSchedules(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Busca o usuário solicitante
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(userRows(10, "Owner", "owner@test.com"))

	// Membro alvo ativo
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(memberRows(50, 1, 20, nil))

	// Schedules com status pending
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "schedules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_member_id", "status"}).
			AddRow(1, 50, "pending"))

	err := Remove(db, "barber-shop", 10, 20)
	if !errors.Is(err, ErrMemberHasActiveSchedules) {
		t.Errorf("expected ErrMemberHasActiveSchedules, got: %v", err)
	}
}

// TestRemove_MemberNotFound verifica que Remove retorna ErrMemberNotFound quando membro não existe.
func TestRemove_MemberNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Busca o usuário solicitante
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(userRows(10, "Owner", "owner@test.com"))

	// Membro não encontrado
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "created_at", "deleted_at"}))

	err := Remove(db, "barber-shop", 10, 20)
	if !errors.Is(err, ErrMemberNotFound) {
		t.Errorf("expected ErrMemberNotFound, got: %v", err)
	}
}

// TestRemove_OrgNotFound verifica que Remove retorna ErrOrgNotFound para slug inválido.
func TestRemove_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}))

	err := Remove(db, "nonexistent", 10, 20)
	if !errors.Is(err, ErrOrgNotFound) {
		t.Errorf("expected ErrOrgNotFound, got: %v", err)
	}
}

// TestRemove_Forbidden verifica que Remove retorna ErrForbidden quando solicitante não tem permissão.
func TestRemove_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)

	// ownerID=10, solicitante=99 (não é owner nem admin)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(99).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(99, "Regular User", "regular@test.com", "hashed", "user", nil, nil))

	err := Remove(db, "barber-shop", 99, 20)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}
