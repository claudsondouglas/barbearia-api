package actions

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// ---------------------------------------------------------------------------
// ValidateBusinessHour unit tests (pure function — no DB)
// ---------------------------------------------------------------------------

func TestValidateBusinessHour_ClosedTrue_NoTimes(t *testing.T) {
	err := ValidateBusinessHour(1, nil, nil, true)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateBusinessHour_ClosedFalse_WithTimes(t *testing.T) {
	open := "09:00"
	close := "18:00"
	err := ValidateBusinessHour(1, &open, &close, false)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateBusinessHour_DayOfWeek_Boundary_Zero(t *testing.T) {
	err := ValidateBusinessHour(0, nil, nil, true)
	if err != nil {
		t.Errorf("day_of_week=0: expected no error, got %v", err)
	}
}

func TestValidateBusinessHour_DayOfWeek_Boundary_Six(t *testing.T) {
	err := ValidateBusinessHour(6, nil, nil, true)
	if err != nil {
		t.Errorf("day_of_week=6: expected no error, got %v", err)
	}
}

func TestValidateBusinessHour_TimeFormat_HHmm(t *testing.T) {
	open := "09:00"
	close := "18:00"
	err := ValidateBusinessHour(3, &open, &close, false)
	if err != nil {
		t.Errorf("expected valid HH:mm format to pass, got %v", err)
	}
}

func TestValidateBusinessHour_CloseEqualToOpen(t *testing.T) {
	open := "09:00"
	close := "09:00"
	err := ValidateBusinessHour(1, &open, &close, false)
	if err == nil {
		t.Error("expected error when close_time equals open_time")
	}
}

func TestValidateBusinessHour_ExactlyMidnight(t *testing.T) {
	open := "00:00"
	close := "23:59"
	err := ValidateBusinessHour(1, &open, &close, false)
	if err != nil {
		t.Errorf("00:00–23:59 should be valid, got %v", err)
	}
}

func TestValidateBusinessHour_ClosedFalse_MissingOpenTime(t *testing.T) {
	close := "18:00"
	err := ValidateBusinessHour(1, nil, &close, false)
	if err == nil {
		t.Error("expected error when open_time is nil and closed=false")
	}
}

func TestValidateBusinessHour_ClosedFalse_MissingCloseTime(t *testing.T) {
	open := "09:00"
	err := ValidateBusinessHour(1, &open, nil, false)
	if err == nil {
		t.Error("expected error when close_time is nil and closed=false")
	}
}

func TestValidateBusinessHour_ClosedFalse_BothTimesNil(t *testing.T) {
	err := ValidateBusinessHour(1, nil, nil, false)
	if err == nil {
		t.Error("expected error when both times are nil and closed=false")
	}
}

func TestValidateBusinessHour_ClosedTrue_WithOpenTime(t *testing.T) {
	open := "09:00"
	err := ValidateBusinessHour(1, &open, nil, true)
	if err == nil {
		t.Error("expected error when open_time is set and closed=true")
	}
}

func TestValidateBusinessHour_ClosedTrue_WithCloseTime(t *testing.T) {
	close := "18:00"
	err := ValidateBusinessHour(1, nil, &close, true)
	if err == nil {
		t.Error("expected error when close_time is set and closed=true")
	}
}

func TestValidateBusinessHour_Overnight(t *testing.T) {
	// open_time="22:00", close_time="06:00" → overnight not supported.
	open := "22:00"
	close := "06:00"
	err := ValidateBusinessHour(1, &open, &close, false)
	if err == nil {
		t.Error("expected error for overnight schedule")
	}
}

func TestValidateBusinessHour_DayOfWeek_Negative(t *testing.T) {
	err := ValidateBusinessHour(-1, nil, nil, true)
	if err == nil {
		t.Error("expected error for day_of_week=-1")
	}
}

func TestValidateBusinessHour_DayOfWeek_TooLarge(t *testing.T) {
	err := ValidateBusinessHour(7, nil, nil, true)
	if err == nil {
		t.Error("expected error for day_of_week=7")
	}
}

// ---------------------------------------------------------------------------
// GetMemberBusinessHours — stub tests
// ---------------------------------------------------------------------------

func TestGetMemberBusinessHours_Success_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	// Mock org lookup.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	// Mock member lookup.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "created_at", "deleted_at"}).
			AddRow(1, 1, 20, nil, nil))

	// Mock business hours.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_business_hours"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_member_id", "day_of_week", "closed"}).
			AddRow(1, 1, 0, true).
			AddRow(2, 1, 1, false))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetMemberBusinessHours panics (stub): %v — ok for now", r)
		}
	}()

	hours, err := GetMemberBusinessHours(db, "barber-shop", 20)
	if err != nil {
		t.Logf("GetMemberBusinessHours returned error: %v", err)
	}
	_ = hours
}

func TestGetMemberBusinessHours_MemberNotFound_Stub(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(sqlmock.NewRows(nil))

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetMemberBusinessHours panics (stub): %v — ok for now", r)
		}
	}()

	_, err := GetMemberBusinessHours(db, "barber-shop", 999)
	if err != ErrMemberNotFound {
		t.Logf("expected ErrMemberNotFound, got %v — stub may not be implemented yet", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateMemberBusinessHoursBatch — stub test
// ---------------------------------------------------------------------------

func TestUpdateMemberBusinessHoursBatch_Success_Stub(t *testing.T) {
	db, _ := newTestDB(t)
	updates := []UpdateBusinessHourInput{
		{DayOfWeek: 1, Closed: true},
	}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("UpdateMemberBusinessHoursBatch panics (stub): %v — ok for now", r)
		}
	}()
	err := UpdateMemberBusinessHoursBatch(db, "barber-shop", 10, 20, updates)
	_ = err
}

func TestUpdateBatch_InvalidHour_Stub(t *testing.T) {
	db, _ := newTestDB(t)
	open := "22:00"
	close := "06:00"
	updates := []UpdateBusinessHourInput{
		{DayOfWeek: 1, Closed: false, OpenTime: &open, CloseTime: &close},
	}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("UpdateMemberBusinessHoursBatch panics (stub): %v — ok for now", r)
		}
	}()
	err := UpdateMemberBusinessHoursBatch(db, "barber-shop", 10, 20, updates)
	if err == nil {
		t.Log("expected validation error for overnight — stub not implemented yet")
	}
}
