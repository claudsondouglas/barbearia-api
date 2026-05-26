package actions

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func hhmm(h, m int) time.Time {
	return time.Date(2026, 7, 1, h, m, 0, 0, time.UTC)
}

// ---------------------------------------------------------------------------
// TestGenerateSlots
// ---------------------------------------------------------------------------

func TestGenerateSlots_ExactFit(t *testing.T) {
	// 60-min window, 30-min slots → 2 slots (09:00, 09:30).
	open := hhmm(9, 0)
	close := hhmm(10, 0)
	slots := GenerateSlots(open, close, 30)
	if len(slots) != 2 {
		t.Errorf("expected 2 slots, got %d: %v", len(slots), slots)
	}
}

func TestGenerateSlots_SingleSlot(t *testing.T) {
	// 30-min window, 30-min duration → 1 slot.
	open := hhmm(9, 0)
	close := hhmm(9, 30)
	slots := GenerateSlots(open, close, 30)
	if len(slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(slots))
	}
}

func TestGenerateSlots_Multiple(t *testing.T) {
	// 09:00–12:00, 30-min → 6 slots (09:00, 09:30, 10:00, 10:30, 11:00, 11:30).
	// 12:00 is excluded because 12:00 + 30min > 12:00.
	open := hhmm(9, 0)
	close := hhmm(12, 0)
	slots := GenerateSlots(open, close, 30)
	if len(slots) != 6 {
		t.Errorf("expected 6 slots, got %d: %v", len(slots), slots)
	}
	if !slots[0].Equal(hhmm(9, 0)) {
		t.Errorf("first slot should be 09:00, got %v", slots[0])
	}
	if !slots[5].Equal(hhmm(11, 30)) {
		t.Errorf("last slot should be 11:30, got %v", slots[5])
	}
}

func TestGenerateSlots_60MinDuration(t *testing.T) {
	// 09:00–12:00, 60-min → 3 slots (09:00, 10:00, 11:00).
	open := hhmm(9, 0)
	close := hhmm(12, 0)
	slots := GenerateSlots(open, close, 60)
	if len(slots) != 3 {
		t.Errorf("expected 3 slots, got %d", len(slots))
	}
}

func TestGenerateSlots_EndIsExclusive(t *testing.T) {
	// Last slot at t+duration == window_close is included.
	// t+duration > window_close is excluded.
	open := hhmm(9, 0)
	close := hhmm(9, 30)
	// 30-min duration: slot at 09:00 ends at 09:30 == close → included.
	slots := GenerateSlots(open, close, 30)
	if len(slots) != 1 {
		t.Errorf("expected 1 slot (end inclusive), got %d", len(slots))
	}

	// 31-min duration: slot at 09:00 ends at 09:31 > 09:30 → excluded.
	slots = GenerateSlots(open, close, 31)
	if len(slots) != 0 {
		t.Errorf("expected 0 slots (end exclusive when over), got %d", len(slots))
	}
}

func TestGenerateSlots_WindowSmallerThanDuration(t *testing.T) {
	// 20-min window, 30-min duration → 0 slots.
	open := hhmm(9, 0)
	close := hhmm(9, 20)
	slots := GenerateSlots(open, close, 30)
	if len(slots) != 0 {
		t.Errorf("expected 0 slots, got %d", len(slots))
	}
}

func TestGenerateSlots_ZeroWindow(t *testing.T) {
	// Window of 0 minutes → 0 slots.
	open := hhmm(9, 0)
	slots := GenerateSlots(open, open, 30)
	if len(slots) != 0 {
		t.Errorf("expected 0 slots for zero window, got %d", len(slots))
	}
}

// ---------------------------------------------------------------------------
// TestOverlaps
// ---------------------------------------------------------------------------

func TestOverlaps_ExactMatch(t *testing.T) {
	// [09:00–09:30] with [09:00–09:30] → true.
	if !Overlaps(hhmm(9, 0), hhmm(9, 30), hhmm(9, 0), hhmm(9, 30)) {
		t.Error("exact match should overlap")
	}
}

func TestOverlaps_PartialLeft(t *testing.T) {
	// slot [09:00–09:30], appt [08:45–09:15] → true.
	if !Overlaps(hhmm(9, 0), hhmm(9, 30), hhmm(8, 45), hhmm(9, 15)) {
		t.Error("partial left overlap should return true")
	}
}

func TestOverlaps_PartialRight(t *testing.T) {
	// slot [09:00–09:30], appt [09:15–09:45] → true.
	if !Overlaps(hhmm(9, 0), hhmm(9, 30), hhmm(9, 15), hhmm(9, 45)) {
		t.Error("partial right overlap should return true")
	}
}

func TestOverlaps_Contained(t *testing.T) {
	// slot [09:00–09:30] inside appt [08:00–10:00] → true.
	if !Overlaps(hhmm(9, 0), hhmm(9, 30), hhmm(8, 0), hhmm(10, 0)) {
		t.Error("contained overlap should return true")
	}
}

func TestOverlaps_Contains(t *testing.T) {
	// slot [09:00–10:00] contains appt [09:15–09:45] → true.
	if !Overlaps(hhmm(9, 0), hhmm(10, 0), hhmm(9, 15), hhmm(9, 45)) {
		t.Error("contains overlap should return true")
	}
}

func TestOverlaps_AdjacentBefore(t *testing.T) {
	// slot [09:00–09:30], appt [08:30–09:00] → false (touching, not overlapping).
	if Overlaps(hhmm(9, 0), hhmm(9, 30), hhmm(8, 30), hhmm(9, 0)) {
		t.Error("adjacent before should not overlap")
	}
}

func TestOverlaps_AdjacentAfter(t *testing.T) {
	// slot [09:00–09:30], appt [09:30–10:00] → false.
	if Overlaps(hhmm(9, 0), hhmm(9, 30), hhmm(9, 30), hhmm(10, 0)) {
		t.Error("adjacent after should not overlap")
	}
}

func TestOverlaps_CompletelyBefore(t *testing.T) {
	// slot [09:00–09:30], appt [08:00–08:45] → false.
	if Overlaps(hhmm(9, 0), hhmm(9, 30), hhmm(8, 0), hhmm(8, 45)) {
		t.Error("completely before should not overlap")
	}
}

func TestOverlaps_CompletelyAfter(t *testing.T) {
	// slot [09:00–09:30], appt [10:00–10:30] → false.
	if Overlaps(hhmm(9, 0), hhmm(9, 30), hhmm(10, 0), hhmm(10, 30)) {
		t.Error("completely after should not overlap")
	}
}

// ---------------------------------------------------------------------------
// TestDiscardPastSlots
// ---------------------------------------------------------------------------

func TestDiscardPastSlots_AllFuture(t *testing.T) {
	// now=09:00; slots [09:00, 09:30, 10:00] → all kept (09:00 == now is kept).
	now := hhmm(9, 0)
	slots := []time.Time{hhmm(9, 0), hhmm(9, 30), hhmm(10, 0)}
	result := DiscardPastSlots(slots, now)
	if len(result) != 3 {
		t.Errorf("expected 3 slots, got %d", len(result))
	}
}

func TestDiscardPastSlots_SomeDiscarded(t *testing.T) {
	// now=09:15; slots [09:00, 09:30, 10:00] → [09:30, 10:00].
	now := hhmm(9, 15)
	slots := []time.Time{hhmm(9, 0), hhmm(9, 30), hhmm(10, 0)}
	result := DiscardPastSlots(slots, now)
	if len(result) != 2 {
		t.Errorf("expected 2 slots, got %d", len(result))
	}
	if !result[0].Equal(hhmm(9, 30)) {
		t.Errorf("first kept slot should be 09:30, got %v", result[0])
	}
}

func TestDiscardPastSlots_AllDiscarded(t *testing.T) {
	// now=23:00; morning slots → empty.
	now := hhmm(23, 0)
	slots := []time.Time{hhmm(9, 0), hhmm(9, 30), hhmm(10, 0)}
	result := DiscardPastSlots(slots, now)
	if len(result) != 0 {
		t.Errorf("expected 0 slots, got %d", len(result))
	}
}

func TestDiscardPastSlots_ExactNow(t *testing.T) {
	// Slot whose start == now is kept (spec: "slot cujo início é exatamente igual ao minuto atual é mantido").
	now := hhmm(9, 30)
	slots := []time.Time{hhmm(9, 0), hhmm(9, 30), hhmm(10, 0)}
	result := DiscardPastSlots(slots, now)
	if len(result) != 2 {
		t.Errorf("expected 2 slots (09:30 and 10:00), got %d", len(result))
	}
	if !result[0].Equal(hhmm(9, 30)) {
		t.Errorf("first kept slot should be 09:30, got %v", result[0])
	}
}

func TestDiscardPastSlots_EmptyInput(t *testing.T) {
	now := hhmm(9, 0)
	result := DiscardPastSlots(nil, now)
	if len(result) != 0 {
		t.Errorf("expected 0 slots from nil input, got %d", len(result))
	}
}
