package actions

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ---------------------------------------------------------------------------
// GenerateSlug unit tests
// ---------------------------------------------------------------------------

func TestGenerateSlug_Simple(t *testing.T) {
	got := GenerateSlug("Barbearia Teste")
	want := "barbearia-teste"
	if got != want {
		t.Errorf("GenerateSlug(%q) = %q, want %q", "Barbearia Teste", got, want)
	}
}

func TestGenerateSlug_AccentsAndSpecial(t *testing.T) {
	got := GenerateSlug("Barbearia do João")
	want := "barbearia-do-joao"
	if got != want {
		t.Errorf("GenerateSlug(%q) = %q, want %q", "Barbearia do João", got, want)
	}
}

func TestGenerateSlug_UpperCase(t *testing.T) {
	got := GenerateSlug("BARBER SHOP")
	want := "barber-shop"
	if got != want {
		t.Errorf("GenerateSlug(%q) = %q, want %q", "BARBER SHOP", got, want)
	}
}

func TestGenerateSlug_MultipleSpaces(t *testing.T) {
	got := GenerateSlug("  My   Barber  ")
	want := "my-barber"
	if got != want {
		t.Errorf("GenerateSlug(%q) = %q, want %q", "  My   Barber  ", got, want)
	}
}

func TestGenerateSlug_OnlyLowercaseAndHyphens(t *testing.T) {
	got := GenerateSlug("Barbearia & Beleza #1!")
	for _, r := range got {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			t.Errorf("GenerateSlug produced invalid char %q in %q", r, got)
		}
	}
	if len(got) > 0 && (got[0] == '-' || got[len(got)-1] == '-') {
		t.Errorf("GenerateSlug has leading/trailing hyphen: %q", got)
	}
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestCreate_InvalidTimezone(t *testing.T) {
	db, _ := newTestDB(t)
	input := CreateOrgInput{
		Name:         "Org TZ",
		Phone:        "(11) 99999-9999",
		Email:        "org@test.com",
		Street:       "Rua A",
		Number:       "1",
		Neighborhood: "Centro",
		City:         "SP",
		State:        "SP",
		ZipCode:      "00000-000",
		Timezone:     "Mars/Olympus",
	}
	_, err := Create(db, 1, input)
	if err != ErrInvalidTimezone {
		t.Errorf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestCreate_DefaultTimezone(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organizations"`)).
		WithArgs("barbearia-teste").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "organizations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	for i := 0; i < 7; i++ {
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "member_business_hours"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(i + 1)))
	}
	mock.ExpectCommit()

	input := CreateOrgInput{
		Name:         "Barbearia Teste",
		Phone:        "(11) 99999-9999",
		Email:        "org@test.com",
		Street:       "Rua A",
		Number:       "1",
		Neighborhood: "Centro",
		City:         "São Paulo",
		State:        "SP",
		ZipCode:      "01310-100",
	}
	org, err := Create(db, 1, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Timezone != "America/Sao_Paulo" {
		t.Errorf("expected default timezone America/Sao_Paulo, got %s", org.Timezone)
	}
}

func TestCreate_SlugUniqueness(t *testing.T) {
	db, mock := newTestDB(t)

	// First slug taken, second is free.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organizations"`)).
		WithArgs("barber-shop").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organizations"`)).
		WithArgs("barber-shop-2").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "organizations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	for i := 0; i < 7; i++ {
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "member_business_hours"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(i + 1)))
	}
	mock.ExpectCommit()

	input := CreateOrgInput{
		Name:         "Barber Shop",
		Phone:        "(11) 99999-9999",
		Email:        "org@test.com",
		Street:       "Rua A",
		Number:       "1",
		Neighborhood: "Centro",
		City:         "SP",
		State:        "SP",
		ZipCode:      "00000-000",
	}
	org, err := Create(db, 1, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Slug != "barber-shop-2" {
		t.Errorf("expected slug barber-shop-2, got %s", org.Slug)
	}
}

// ---------------------------------------------------------------------------
// FindBySlug tests
// ---------------------------------------------------------------------------

func TestFindBySlug_Found(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	org, err := FindBySlug(db, "barber-shop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Slug != "barber-shop" {
		t.Errorf("expected slug barber-shop, got %s", org.Slug)
	}
}

func TestFindBySlug_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("nonexistent", 1).
		WillReturnRows(sqlmock.NewRows(nil))

	_, err := FindBySlug(db, "nonexistent")
	if err != ErrOrgNotFound {
		t.Errorf("expected ErrOrgNotFound, got %v", err)
	}
}

func TestFindBySlug_SoftDeleted(t *testing.T) {
	db, mock := newTestDB(t)

	// GORM automatically adds deleted_at IS NULL; soft-deleted rows are excluded.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("deleted-org", 1).
		WillReturnRows(sqlmock.NewRows(nil))

	_, err := FindBySlug(db, "deleted-org")
	if err != ErrOrgNotFound {
		t.Errorf("expected ErrOrgNotFound for soft-deleted org, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update tests
// ---------------------------------------------------------------------------

func TestUpdate_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("nonexistent", 1).
		WillReturnRows(sqlmock.NewRows(nil))

	_, err := Update(db, "nonexistent", 10, "user", UpdateOrgInput{})
	if err != ErrOrgNotFound {
		t.Errorf("expected ErrOrgNotFound, got %v", err)
	}
}

func TestUpdate_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	_, err := Update(db, "barber-shop", 99, "user", UpdateOrgInput{})
	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdate_AdminCanUpdateAnyOrg(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	phone := "(11) 88888-8888"
	org, err := Update(db, "barber-shop", 99, "admin", UpdateOrgInput{Phone: &phone})
	if err != nil {
		t.Fatalf("admin should be allowed, got error: %v", err)
	}
	if org.Phone != phone {
		t.Errorf("expected phone %s, got %s", phone, org.Phone)
	}
}

func TestUpdate_InvalidTimezone(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	tz := "Mars/Olympus"
	_, err := Update(db, "barber-shop", 10, "user", UpdateOrgInput{Timezone: &tz})
	if err != ErrInvalidTimezone {
		t.Errorf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestUpdate_NoFields(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	org, err := Update(db, "barber-shop", 10, "user", UpdateOrgInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Slug != "barber-shop" {
		t.Errorf("expected unchanged slug, got %s", org.Slug)
	}
}

func TestUpdate_OwnerCanUpdate(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	phone := "(11) 77777-7777"
	org, err := Update(db, "barber-shop", 10, "user", UpdateOrgInput{Phone: &phone})
	if err != nil {
		t.Fatalf("owner should be allowed, got error: %v", err)
	}
	if org.Phone != phone {
		t.Errorf("expected phone %s, got %s", phone, org.Phone)
	}
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestDelete_NotFound(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("nonexistent", 1).
		WillReturnRows(sqlmock.NewRows(nil))

	err := Delete(db, "nonexistent")
	if err != ErrOrgNotFound {
		t.Errorf("expected ErrOrgNotFound, got %v", err)
	}
}

func TestDelete_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := Delete(db, "barber-shop"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// List tests
// ---------------------------------------------------------------------------

func TestList_Active(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	orgs, err := List(db, 20, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orgs) != 1 {
		t.Errorf("expected 1 org, got %d", len(orgs))
	}
}

func TestList_Deleted(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows(nil))

	orgs, err := List(db, 20, 0, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orgs) != 0 {
		t.Errorf("expected 0 orgs, got %d", len(orgs))
	}
}

// ---------------------------------------------------------------------------
// MyOrgs tests
// ---------------------------------------------------------------------------

func TestMyOrgs_ReturnsOrgs(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	orgs, err := MyOrgs(db, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orgs) != 1 {
		t.Errorf("expected 1 org, got %d", len(orgs))
	}
}

func TestMyOrgs_Empty(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows(nil))

	orgs, err := MyOrgs(db, 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orgs) != 0 {
		t.Errorf("expected empty result, got %d orgs", len(orgs))
	}
}

// ---------------------------------------------------------------------------
// helper: unused import guard
// ---------------------------------------------------------------------------

var _ = time.Now
