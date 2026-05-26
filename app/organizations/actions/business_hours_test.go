package actions

import (
	"testing"

	"barbearia-api/models"
)

// ---------------------------------------------------------------------------
// AggregateBusinessHours unit tests (pure function — no DB)
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

func TestAggregate_NoMembers(t *testing.T) {
	result := AggregateBusinessHours(nil)
	if len(result) != 7 {
		t.Fatalf("expected 7 days, got %d", len(result))
	}
	for _, d := range result {
		if !d.Closed {
			t.Errorf("day %d: expected closed=true, got false", d.DayOfWeek)
		}
		if d.OpenTime != nil || d.CloseTime != nil {
			t.Errorf("day %d: expected nil times when closed, got open=%v close=%v", d.DayOfWeek, d.OpenTime, d.CloseTime)
		}
	}
}

func TestAggregate_AllClosed(t *testing.T) {
	hours := make([]models.MemberBusinessHour, 14)
	for i := 0; i < 7; i++ {
		hours[i] = models.MemberBusinessHour{OrgMemberID: 1, DayOfWeek: i, Closed: true}
		hours[i+7] = models.MemberBusinessHour{OrgMemberID: 2, DayOfWeek: i, Closed: true}
	}
	result := AggregateBusinessHours(hours)
	if len(result) != 7 {
		t.Fatalf("expected 7 days, got %d", len(result))
	}
	for _, d := range result {
		if !d.Closed {
			t.Errorf("day %d: expected closed=true", d.DayOfWeek)
		}
	}
}

func TestAggregate_SingleMember(t *testing.T) {
	// Only day 1 (Monday) is open for member 1.
	hours := []models.MemberBusinessHour{
		{OrgMemberID: 1, DayOfWeek: 0, Closed: true},
		{OrgMemberID: 1, DayOfWeek: 1, Closed: false, OpenTime: strPtr("09:00"), CloseTime: strPtr("18:00")},
		{OrgMemberID: 1, DayOfWeek: 2, Closed: true},
		{OrgMemberID: 1, DayOfWeek: 3, Closed: true},
		{OrgMemberID: 1, DayOfWeek: 4, Closed: true},
		{OrgMemberID: 1, DayOfWeek: 5, Closed: true},
		{OrgMemberID: 1, DayOfWeek: 6, Closed: true},
	}
	result := AggregateBusinessHours(hours)
	if len(result) != 7 {
		t.Fatalf("expected 7 days, got %d", len(result))
	}
	// Day 1 should be open.
	if result[1].Closed {
		t.Error("day 1: expected closed=false")
	}
	if result[1].OpenTime == nil || *result[1].OpenTime != "09:00" {
		t.Errorf("day 1: expected open_time=09:00, got %v", result[1].OpenTime)
	}
	if result[1].CloseTime == nil || *result[1].CloseTime != "18:00" {
		t.Errorf("day 1: expected close_time=18:00, got %v", result[1].CloseTime)
	}
	// All other days should be closed.
	for _, d := range result {
		if d.DayOfWeek == 1 {
			continue
		}
		if !d.Closed {
			t.Errorf("day %d: expected closed=true", d.DayOfWeek)
		}
	}
}

func TestAggregate_TwoMembers_MinMax(t *testing.T) {
	// Member A: day 2 — 08:00–17:00
	// Member B: day 2 — 10:00–19:00
	// Expected: day 2 — open_time=08:00, close_time=19:00
	hours := []models.MemberBusinessHour{
		{OrgMemberID: 1, DayOfWeek: 2, Closed: false, OpenTime: strPtr("08:00"), CloseTime: strPtr("17:00")},
		{OrgMemberID: 2, DayOfWeek: 2, Closed: false, OpenTime: strPtr("10:00"), CloseTime: strPtr("19:00")},
	}
	result := AggregateBusinessHours(hours)
	day := result[2]
	if day.Closed {
		t.Fatal("day 2: expected closed=false")
	}
	if day.OpenTime == nil || *day.OpenTime != "08:00" {
		t.Errorf("day 2: expected open_time=08:00 (MIN), got %v", day.OpenTime)
	}
	if day.CloseTime == nil || *day.CloseTime != "19:00" {
		t.Errorf("day 2: expected close_time=19:00 (MAX), got %v", day.CloseTime)
	}
}

func TestAggregate_TwoMembers_OneClosedOneOpen(t *testing.T) {
	// Member A: day 3 — closed=true
	// Member B: day 3 — closed=false, 09:00–17:00
	// Expected: day 3 open with member B's hours.
	hours := []models.MemberBusinessHour{
		{OrgMemberID: 1, DayOfWeek: 3, Closed: true},
		{OrgMemberID: 2, DayOfWeek: 3, Closed: false, OpenTime: strPtr("09:00"), CloseTime: strPtr("17:00")},
	}
	result := AggregateBusinessHours(hours)
	day := result[3]
	if day.Closed {
		t.Fatal("day 3: expected closed=false because member B is open")
	}
	if day.OpenTime == nil || *day.OpenTime != "09:00" {
		t.Errorf("day 3: expected open_time=09:00, got %v", day.OpenTime)
	}
	if day.CloseTime == nil || *day.CloseTime != "17:00" {
		t.Errorf("day 3: expected close_time=17:00, got %v", day.CloseTime)
	}
}

func TestAggregate_ClosedDayNoTimeFields(t *testing.T) {
	hours := []models.MemberBusinessHour{
		{OrgMemberID: 1, DayOfWeek: 0, Closed: true},
	}
	result := AggregateBusinessHours(hours)
	if result[0].OpenTime != nil {
		t.Errorf("day 0: open_time should be nil when closed, got %v", *result[0].OpenTime)
	}
	if result[0].CloseTime != nil {
		t.Errorf("day 0: close_time should be nil when closed, got %v", *result[0].CloseTime)
	}
}

func TestAggregate_DayOrderIsCorrect(t *testing.T) {
	result := AggregateBusinessHours(nil)
	for i, d := range result {
		if d.DayOfWeek != i {
			t.Errorf("position %d has day_of_week=%d, want %d", i, d.DayOfWeek, i)
		}
	}
}
