package actions

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestAdd_NewUser verifica que Add insere um novo membro e cria 7 business hours.
func TestAdd_NewUser(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização pelo slug
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Busca o usuário solicitante (owner check)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(userRows(10, "Owner", "owner@test.com"))

	// Busca o usuário alvo
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(20).
		WillReturnRows(userRows(20, "New Member", "newmember@test.com"))

	// Busca membro existente (retorna vazio — novo membro)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "created_at", "deleted_at"}))

	// INSERT do novo org_member
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))
	mock.ExpectCommit()

	// INSERT de 7 member_business_hours
	for i := 0; i < 7; i++ {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "member_business_hours"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(i + 1)))
		mock.ExpectCommit()
	}

	err := Add(db, "barber-shop", 10, 20)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

// TestAdd_ReactivateSoftDeleted verifica que Add reativa um membro soft-deleted.
func TestAdd_ReactivateSoftDeleted(t *testing.T) {
	db, mock := newTestDB(t)

	deletedAt := time.Now().Add(-24 * time.Hour)

	// Busca a organização
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Busca o usuário solicitante
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(userRows(10, "Owner", "owner@test.com"))

	// Busca o usuário alvo
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(20).
		WillReturnRows(userRows(20, "Old Member", "oldmember@test.com"))

	// Busca membro existente (soft-deleted)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(memberRows(50, 1, 20, deletedAt))

	// UPDATE para reativar (deleted_at = NULL)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "org_members"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// DELETE dos business hours antigos
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "member_business_hours"`)).
		WillReturnResult(sqlmock.NewResult(0, 7))
	mock.ExpectCommit()

	// INSERT de 7 novos business hours
	for i := 0; i < 7; i++ {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "member_business_hours"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(i + 1)))
		mock.ExpectCommit()
	}

	err := Add(db, "barber-shop", 10, 20)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

// TestAdd_AlreadyActiveMember verifica que Add retorna ErrAlreadyActiveMember para membro ativo.
func TestAdd_AlreadyActiveMember(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Busca o usuário solicitante
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(userRows(10, "Owner", "owner@test.com"))

	// Busca o usuário alvo
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(20).
		WillReturnRows(userRows(20, "Active Member", "active@test.com"))

	// Membro já ativo (deleted_at = nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(memberRows(50, 1, 20, nil))

	err := Add(db, "barber-shop", 10, 20)
	if !errors.Is(err, ErrAlreadyActiveMember) {
		t.Errorf("expected ErrAlreadyActiveMember, got: %v", err)
	}
}

// TestAdd_OrgNotFound verifica que Add retorna ErrOrgNotFound para slug inválido.
func TestAdd_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}))

	err := Add(db, "nonexistent", 10, 20)
	if !errors.Is(err, ErrOrgNotFound) {
		t.Errorf("expected ErrOrgNotFound, got: %v", err)
	}
}

// TestAdd_UserNotFound verifica que Add retorna ErrUserNotFound quando o usuário alvo não existe.
func TestAdd_UserNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Busca o usuário solicitante
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(userRows(10, "Owner", "owner@test.com"))

	// Usuário alvo não encontrado
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(99).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}))

	err := Add(db, "barber-shop", 10, 99)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

// TestAdd_Forbidden verifica que Add retorna ErrForbidden quando o solicitante não é owner/admin.
func TestAdd_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização (ownerID = 10)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Solicitante é user comum (ID=99, role=user — não é owner nem admin)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(99).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(99, "Regular User", "regular@test.com", "hashed", "user", nil, nil))

	err := Add(db, "barber-shop", 99, 20)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}
