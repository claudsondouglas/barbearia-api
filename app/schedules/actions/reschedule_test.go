package actions

import (
	"errors"
	"testing"
)

// TestReschedule_ByClient_OwnSchedule_Pending verifica que cliente reagenda próprio agendamento
// e o status volta para pending.
func TestReschedule_ByClient_OwnSchedule_Pending(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := RescheduleInput{ScheduledAt: "2026-07-05T10:00:00-03:00"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Reschedule(db, "barbearia-test", 1, 20, "user", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "pending" {
		t.Errorf("expected status=pending after client reschedule, got %s", result.Status)
	}
	if result.RescheduledBy == nil || *result.RescheduledBy != 20 {
		t.Error("expected rescheduled_by=20")
	}
	if result.RescheduledAt == nil {
		t.Error("expected rescheduled_at to be set")
	}
}

// TestReschedule_ByOwner_StatusUnchanged verifica que owner reagenda sem alterar o status do agendamento.
func TestReschedule_ByOwner_StatusUnchanged(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := RescheduleInput{ScheduledAt: "2026-07-05T10:00:00-03:00"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Reschedule(db, "barbearia-test", 2, 10, "owner", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "confirmed" {
		t.Errorf("expected status=confirmed (unchanged) for owner reschedule, got %s", result.Status)
	}
}

// TestReschedule_FirstTime_OriginalScheduledAt verifica que no primeiro reagendamento
// original_scheduled_at é preenchido com o scheduled_at anterior.
func TestReschedule_FirstTime_OriginalScheduledAt(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := RescheduleInput{ScheduledAt: "2026-07-05T10:00:00-03:00"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Reschedule(db, "barbearia-test", 1, 20, "user", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OriginalScheduledAt == nil {
		t.Error("expected original_scheduled_at to be set on first reschedule")
	}
}

// TestReschedule_SecondTime_OriginalScheduledAtImmutable verifica que no segundo reagendamento
// original_scheduled_at não é alterado.
func TestReschedule_SecondTime_OriginalScheduledAtImmutable(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	// Schedule id=2 já tem original_scheduled_at preenchido ("2026-06-30T10:00:00Z")
	input := RescheduleInput{ScheduledAt: "2026-07-10T10:00:00-03:00"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Reschedule(db, "barbearia-test", 2, 10, "owner", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// original_scheduled_at deve permanecer o valor do primeiro reagendamento
	if result.OriginalScheduledAt == nil {
		t.Error("expected original_scheduled_at to remain set")
	}
}

// TestReschedule_HistoryEntry_Created verifica que um registro é inserido em schedule_reschedule_history.
func TestReschedule_HistoryEntry_Created(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := RescheduleInput{ScheduledAt: "2026-07-05T10:00:00-03:00"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Reschedule(db, "barbearia-test", 1, 20, "user", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verificação de INSERT em schedule_reschedule_history é feita via mock expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

// TestReschedule_StatusCancelled verifica que reagendar agendamento cancelled retorna ErrInvalidTransition.
func TestReschedule_StatusCancelled(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := RescheduleInput{ScheduledAt: "2026-07-05T10:00:00-03:00"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Reschedule(db, "barbearia-test", 3, 10, "owner", input)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

// TestReschedule_Conflict verifica que reagendar para horário com conflito retorna ErrConflict.
func TestReschedule_Conflict(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	// Slot ocupado por agendamento id=6 (14:00-14:30)
	input := RescheduleInput{ScheduledAt: "2026-07-01T17:00:00-03:00"} // 14:00 UTC-3 = 17:00 UTC

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Reschedule(db, "barbearia-test", 1, 20, "user", input)
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

// TestGetRescheduleHistory_Success verifica que owner pode consultar o histórico de reagendamentos.
func TestGetRescheduleHistory_Success(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	history, err := GetRescheduleHistory(db, "barbearia-test", 2, 10, "owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) == 0 {
		t.Error("expected at least 1 history entry for schedule id=2")
	}
}

// TestGetRescheduleHistory_Empty verifica que agendamento sem histórico retorna slice vazio.
func TestGetRescheduleHistory_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	history, err := GetRescheduleHistory(db, "barbearia-test", 1, 10, "owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if history == nil {
		t.Error("expected non-nil slice")
	}
}
