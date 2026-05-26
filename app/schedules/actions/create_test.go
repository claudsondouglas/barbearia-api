package actions

import (
	"errors"
	"testing"

	"barbearia-api/models"
)

// TestCreateSchedule_Client_PendingDefault verifica que cliente cria agendamento com status pending por padrão.
func TestCreateSchedule_Client_PendingDefault(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := models.CreateScheduleInput{
		ServiceID:      1,
		ProfessionalID: 12,
		ScheduledAt:    "2026-06-10T15:00:00-03:00",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Create(db, "barbearia-test", 20, "user", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "pending" {
		t.Errorf("expected status=pending, got %s", result.Status)
	}
}

// TestCreateSchedule_Owner_WalkIn verifica que owner pode criar agendamento com status completed (walk-in).
func TestCreateSchedule_Owner_WalkIn(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := models.CreateScheduleInput{
		ServiceID:      1,
		ProfessionalID: 12,
		ScheduledAt:    "2026-05-01T10:00:00-03:00",
		Status:         "completed",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Create(db, "barbearia-test", 10, "owner", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("expected status=completed, got %s", result.Status)
	}
}

// TestCreateSchedule_Unauthenticated verifica que Create retorna ErrForbidden quando role não é reconhecida.
func TestCreateSchedule_Unauthenticated(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := models.CreateScheduleInput{
		ServiceID:      1,
		ProfessionalID: 12,
		ScheduledAt:    "2026-06-10T15:00:00-03:00",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Create(db, "barbearia-test", 0, "", input)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// TestCreateSchedule_ProfessionalNotMember verifica que Create retorna ErrProfessionalNotMember
// quando o professional_id não é membro ativo da organização.
func TestCreateSchedule_ProfessionalNotMember(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := models.CreateScheduleInput{
		ServiceID:      1,
		ProfessionalID: 99,
		ScheduledAt:    "2026-06-10T15:00:00-03:00",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Create(db, "barbearia-test", 20, "user", input)
	if !errors.Is(err, ErrProfessionalNotMember) {
		t.Errorf("expected ErrProfessionalNotMember, got %v", err)
	}
}

// TestCreateSchedule_PastScheduledAt verifica que Create retorna ErrPastScheduledAt
// quando scheduled_at está no passado e o solicitante não é owner/admin.
func TestCreateSchedule_PastScheduledAt(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := models.CreateScheduleInput{
		ServiceID:      1,
		ProfessionalID: 12,
		ScheduledAt:    "2020-01-01T10:00:00-03:00",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Create(db, "barbearia-test", 20, "user", input)
	if !errors.Is(err, ErrPastScheduledAt) {
		t.Errorf("expected ErrPastScheduledAt, got %v", err)
	}
}

// TestCreateSchedule_Conflict_Pending verifica que Create retorna ErrConflict
// quando há sobreposição com agendamento de status pending do mesmo profissional.
func TestCreateSchedule_Conflict_Pending(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := models.CreateScheduleInput{
		ServiceID:      1,
		ProfessionalID: 12,
		ScheduledAt:    "2026-06-10T13:00:00Z",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Create(db, "barbearia-test", 20, "user", input)
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

// TestCreateSchedule_Conflict_Confirmed verifica que Create retorna ErrConflict
// quando há sobreposição com agendamento de status confirmed.
func TestCreateSchedule_Conflict_Confirmed(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	input := models.CreateScheduleInput{
		ServiceID:      1,
		ProfessionalID: 12,
		ScheduledAt:    "2026-06-10T13:15:00Z",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Create(db, "barbearia-test", 20, "user", input)
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

// TestCreateSchedule_MissingProfessionalId verifica que a ausência de professional_id é detectável.
// (Validação de binding — no nível de actions, professional_id=0 deve retornar erro.)
func TestCreateSchedule_MissingProfessionalId(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	// professional_id=0 representa campo ausente
	input := models.CreateScheduleInput{
		ServiceID:   1,
		ScheduledAt: "2026-06-10T15:00:00-03:00",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Create(db, "barbearia-test", 20, "user", input)
	if err == nil {
		t.Error("expected error for missing professional_id, got nil")
	}
}
