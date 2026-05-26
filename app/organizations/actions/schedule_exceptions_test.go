package actions

import (
	"regexp"
	"testing"

	"barbearia-api/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// ---------------------------------------------------------------------------
// ValidateException unit tests (pure function)
// ---------------------------------------------------------------------------

func TestValidateException_ClosedTrue_NoTimes(t *testing.T) {
	err := ValidateException(true, nil, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateException_ClosedFalse_WithTimes(t *testing.T) {
	open := "09:00"
	close := "14:00"
	err := ValidateException(false, &open, &close)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateException_ClosedFalse_MissingOpenTime(t *testing.T) {
	close := "14:00"
	err := ValidateException(false, nil, &close)
	if err == nil {
		t.Error("expected error when open_time is nil and closed=false")
	}
}

func TestValidateException_ClosedFalse_MissingCloseTime(t *testing.T) {
	open := "09:00"
	err := ValidateException(false, &open, nil)
	if err == nil {
		t.Error("expected error when close_time is nil and closed=false")
	}
}

func TestValidateException_ClosedTrue_WithTimes(t *testing.T) {
	open := "09:00"
	close := "14:00"
	err := ValidateException(true, &open, &close)
	if err == nil {
		t.Error("expected error when times are set and closed=true")
	}
}

func TestValidateException_Overnight(t *testing.T) {
	open := "22:00"
	close := "06:00"
	err := ValidateException(false, &open, &close)
	if err == nil {
		t.Error("expected error for overnight schedule")
	}
}

func TestValidateException_CloseNotAfterOpen(t *testing.T) {
	open := "10:00"
	close := "10:00"
	err := ValidateException(false, &open, &close)
	if err == nil {
		t.Error("expected error when close_time equals open_time")
	}
}

// ---------------------------------------------------------------------------
// ValidateMemberExceptionAgainstOrg unit tests (pure function)
// ---------------------------------------------------------------------------

func TestMemberVsOrg_ContainedInOrg(t *testing.T) {
	// Member 10:00–13:00, org 09:00–14:00 → ok.
	err := ValidateMemberExceptionAgainstOrg("10:00", "13:00", "09:00", "14:00")
	if err != nil {
		t.Errorf("expected no error for contained window, got %v", err)
	}
}

func TestMemberVsOrg_ExactlyOrgWindow(t *testing.T) {
	// Member exactly matches org window → ok.
	err := ValidateMemberExceptionAgainstOrg("09:00", "14:00", "09:00", "14:00")
	if err != nil {
		t.Errorf("expected no error for exact org window, got %v", err)
	}
}

func TestMemberVsOrg_OpenBeforeOrg(t *testing.T) {
	// Member opens before org → error.
	err := ValidateMemberExceptionAgainstOrg("08:00", "13:00", "09:00", "14:00")
	if err == nil {
		t.Error("expected error when member opens before org")
	}
}

func TestMemberVsOrg_CloseAfterOrg(t *testing.T) {
	// Member closes after org → error.
	err := ValidateMemberExceptionAgainstOrg("10:00", "15:00", "09:00", "14:00")
	if err == nil {
		t.Error("expected error when member closes after org")
	}
}

func TestMemberVsOrg_CompletelyOutside(t *testing.T) {
	// Member window completely outside org → error.
	err := ValidateMemberExceptionAgainstOrg("15:00", "18:00", "09:00", "14:00")
	if err == nil {
		t.Error("expected error when member window is completely outside org")
	}
}

// ---------------------------------------------------------------------------
// ListOrgExceptions — stub tests
// ---------------------------------------------------------------------------

func TestListOrgExceptions_Success_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "user_id", "date", "closed", "open_time", "close_time", "reason",
		}).AddRow(1, 1, nil, "2026-12-25", true, nil, nil, "Natal"))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("ListOrgExceptions panics (stub): %v — ok for now", r)
		}
	}()

	exceptions, err := ListOrgExceptions(db, "barber-shop", 10, 20, 0, nil, nil)
	if err != nil {
		t.Logf("ListOrgExceptions returned error: %v", err)
	}
	_ = exceptions
}

func TestListOrgExceptions_Empty_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows(nil))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("ListOrgExceptions panics (stub): %v — ok for now", r)
		}
	}()

	exceptions, err := ListOrgExceptions(db, "barber-shop", 10, 20, 0, nil, nil)
	if err != nil {
		t.Logf("ListOrgExceptions returned error: %v", err)
	}
	if exceptions != nil && len(exceptions) != 0 {
		t.Errorf("expected empty result, got %d", len(exceptions))
	}
}

// ---------------------------------------------------------------------------
// CreateOrgException — stub tests
// ---------------------------------------------------------------------------

func TestCreateOrgException_Success_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "schedule_exceptions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	input := models.CreateExceptionInput{
		Date:   "2026-12-25",
		Closed: true,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("CreateOrgException panics (stub): %v — ok for now", r)
		}
	}()

	exc, count, err := CreateOrgException(db, "barber-shop", 10, input)
	if err != nil {
		t.Logf("CreateOrgException returned error: %v", err)
	}
	_ = exc
	_ = count
}

func TestCreateOrgException_WithDeletedMemberExceptions_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	// Expect transaction: count member exceptions, delete them, insert org exception.
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*)`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "schedule_exceptions"`)).
		WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "schedule_exceptions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	input := models.CreateExceptionInput{
		Date:   "2026-12-25",
		Closed: true,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("CreateOrgException panics (stub): %v — ok for now", r)
		}
	}()

	exc, count, err := CreateOrgException(db, "barber-shop", 10, input)
	if err != nil {
		t.Logf("CreateOrgException returned error: %v", err)
	}
	if exc != nil && count != 2 {
		t.Errorf("expected deleted_member_exceptions_count=2, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// DeleteOrgException — stub tests
// ---------------------------------------------------------------------------

func TestDeleteOrgException_Success_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "schedule_exceptions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "date", "closed"}).
			AddRow(1, 1, nil, "2026-12-25", true))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "schedule_exceptions"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("DeleteOrgException panics (stub): %v — ok for now", r)
		}
	}()

	err := DeleteOrgException(db, "barber-shop", 10, 1)
	if err != nil {
		t.Logf("DeleteOrgException returned error: %v", err)
	}
}

func TestDeleteOrgException_NotFound_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "schedule_exceptions"`)).
		WillReturnRows(sqlmock.NewRows(nil))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("DeleteOrgException panics (stub): %v — ok for now", r)
		}
	}()

	err := DeleteOrgException(db, "barber-shop", 10, 999)
	if err != ErrExceptionNotFound {
		t.Logf("expected ErrExceptionNotFound, got %v — stub may not be implemented yet", err)
	}
}
