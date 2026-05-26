package actions

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestList_Success verifica que List retorna slice com 3 membros ativos.
func TestList_Success(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização pelo slug
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Query de membros com JOIN em users
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "email", "created_at"}).
			AddRow(1, 10, "Alice", "alice@test.com", time.Now()).
			AddRow(2, 20, "Bob", "bob@test.com", time.Now()).
			AddRow(3, 30, "Carol", "carol@test.com", time.Now()))

	members, err := List(db, "barber-shop")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("expected 3 members, got: %d", len(members))
	}
}

// TestList_ExcludesSoftDeleted verifica que membros soft-deleted não são retornados.
func TestList_ExcludesSoftDeleted(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Query com WHERE deleted_at IS NULL retorna apenas 2 membros ativos
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "email", "created_at"}).
			AddRow(1, 10, "Alice", "alice@test.com", time.Now()).
			AddRow(2, 20, "Bob", "bob@test.com", time.Now()))

	members, err := List(db, "barber-shop")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members (soft-deleted excluded), got: %d", len(members))
	}
}

// TestList_Empty verifica que List retorna slice vazio sem erro quando não há membros.
func TestList_Empty(t *testing.T) {
	db, mock := newTestDB(t)

	// Busca a organização
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(orgRows(1, "barber-shop", 10))

	// Query retorna vazio
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "email", "created_at"}))

	members, err := List(db, "barber-shop")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if members == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(members) != 0 {
		t.Errorf("expected 0 members, got: %d", len(members))
	}
}

// TestList_OrgNotFound verifica que List retorna ErrOrgNotFound para slug inválido.
func TestList_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}))

	_, err := List(db, "nonexistent")
	if !errors.Is(err, ErrOrgNotFound) {
		t.Errorf("expected ErrOrgNotFound, got: %v", err)
	}
}
