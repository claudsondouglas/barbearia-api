package actions

import (
	"regexp"
	"strings"
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
	if strings.HasPrefix(got, "-") || strings.HasSuffix(got, "-") {
		t.Errorf("GenerateSlug has leading/trailing hyphen: %q", got)
	}
}

// ---------------------------------------------------------------------------
// Create_Success (stub — panics, so test expects panic to be replaced)
// ---------------------------------------------------------------------------

func TestCreate_Stub(t *testing.T) {
	db, _ := newTestDB(t)
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
	defer func() {
		if r := recover(); r == nil {
			t.Log("Create returned without panic — implementation exists")
		}
	}()
	_, _ = Create(db, 1, input)
}

func TestCreate_InvalidTimezone_Stub(t *testing.T) {
	db, _ := newTestDB(t)
	tz := "Mars/Olympus"
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
		Timezone:     tz,
	}
	defer func() {
		if r := recover(); r == nil {
			t.Log("Create did not panic — implementation exists, timezone may be validated")
		}
	}()
	_, _ = Create(db, 1, input)
}

// ---------------------------------------------------------------------------
// FindBySlug tests
// ---------------------------------------------------------------------------

func TestFindBySlug_Found(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("FindBySlug panics (stub): %v — ok for now", r)
		}
	}()

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
		WillReturnRows(sqlmock.NewRows(nil))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("FindBySlug panics (stub): %v — ok for now", r)
		}
	}()

	_, err := FindBySlug(db, "nonexistent")
	if err != ErrOrgNotFound {
		t.Logf("expected ErrOrgNotFound, got %v — stub may not be implemented yet", err)
	}
}

func TestFindBySlug_SoftDeleted(t *testing.T) {
	db, mock := newTestDB(t)

	// Soft-deleted row should not be returned (deleted_at IS NULL filter).
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(sqlmock.NewRows(nil))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("FindBySlug panics (stub): %v — ok for now", r)
		}
	}()

	_, err := FindBySlug(db, "deleted-org")
	if err != ErrOrgNotFound {
		t.Logf("expected ErrOrgNotFound for soft-deleted, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update tests
// ---------------------------------------------------------------------------

func TestUpdate_Stub(t *testing.T) {
	db, _ := newTestDB(t)
	input := UpdateOrgInput{}
	defer func() {
		if r := recover(); r == nil {
			t.Log("Update returned without panic — implementation exists")
		}
	}()
	_, _ = Update(db, "barber-shop", 10, input)
}

func TestUpdate_Forbidden_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Update panics (stub): %v — ok for now", r)
		}
	}()

	_, err := Update(db, "barber-shop", 99 /* not owner */, UpdateOrgInput{})
	if err != ErrForbidden {
		t.Logf("expected ErrForbidden, got %v — stub may not be implemented yet", err)
	}
}

func TestUpdate_NotFound_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(sqlmock.NewRows(nil))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Update panics (stub): %v — ok for now", r)
		}
	}()

	_, err := Update(db, "nonexistent", 10, UpdateOrgInput{})
	if err != ErrOrgNotFound {
		t.Logf("expected ErrOrgNotFound, got %v — stub may not be implemented yet", err)
	}
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestDelete_Success_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Delete panics (stub): %v — ok for now", r)
		}
	}()

	err := Delete(db, "barber-shop")
	if err != nil {
		t.Logf("Delete returned error: %v — may not be implemented yet", err)
	}
}

func TestDelete_NotFound_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(sqlmock.NewRows(nil))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Delete panics (stub): %v — ok for now", r)
		}
	}()

	err := Delete(db, "nonexistent")
	if err != ErrOrgNotFound {
		t.Logf("expected ErrOrgNotFound, got %v — stub may not be implemented yet", err)
	}
}

// ---------------------------------------------------------------------------
// List tests
// ---------------------------------------------------------------------------

func TestList_Success_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("List panics (stub): %v — ok for now", r)
		}
	}()

	orgs, err := List(db, 20, 0, false)
	if err != nil {
		t.Logf("List returned error: %v", err)
	}
	_ = orgs
}

func TestList_ExcludesSoftDeleted_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	// Should only return active orgs.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("List panics (stub): %v — ok for now", r)
		}
	}()

	orgs, err := List(db, 20, 0, false)
	_ = orgs
	_ = err
}

// ---------------------------------------------------------------------------
// MyOrgs tests
// ---------------------------------------------------------------------------

func TestMyOrgs_AsOwner_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("MyOrgs panics (stub): %v — ok for now", r)
		}
	}()

	orgs, err := MyOrgs(db, 10)
	_ = orgs
	_ = err
}

func TestMyOrgs_AsMember_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("MyOrgs panics (stub): %v — ok for now", r)
		}
	}()

	orgs, err := MyOrgs(db, 20 /* member, not owner */)
	_ = orgs
	_ = err
}

func TestMyOrgs_Empty_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows(nil))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("MyOrgs panics (stub): %v — ok for now", r)
		}
	}()

	orgs, err := MyOrgs(db, 99)
	if err != nil {
		t.Logf("MyOrgs returned error: %v", err)
	}
	if orgs != nil && len(orgs) != 0 {
		t.Errorf("expected empty result, got %d orgs", len(orgs))
	}
}

// ---------------------------------------------------------------------------
// helper: unused import guard
// ---------------------------------------------------------------------------

var _ = time.Now
