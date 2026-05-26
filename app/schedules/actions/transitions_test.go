package actions

import (
	"errors"
	"testing"
)

// --- Confirm ---

// TestConfirm_ByProfessional_Success verifica que o profissional responsável pode confirmar agendamento pending.
func TestConfirm_ByProfessional_Success(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Confirm(db, "barbearia-test", 1, 12, "professional")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "confirmed" {
		t.Errorf("expected status=confirmed, got %s", result.Status)
	}
	if result.ConfirmedBy == nil || *result.ConfirmedBy != 12 {
		t.Error("expected confirmed_by=12")
	}
	if result.ConfirmedAt == nil {
		t.Error("expected confirmed_at to be set")
	}
}

// TestConfirm_ByClient_Forbidden verifica que cliente não pode confirmar agendamento.
func TestConfirm_ByClient_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Confirm(db, "barbearia-test", 1, 20, "user")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// TestConfirm_AlreadyConfirmed_InvalidTransition verifica que confirmar agendamento já confirmado retorna ErrInvalidTransition.
func TestConfirm_AlreadyConfirmed_InvalidTransition(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Confirm(db, "barbearia-test", 2, 10, "owner")
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

// --- Cancel ---

// TestCancel_ByClient_Own_Pending verifica que cliente pode cancelar próprio agendamento pending.
func TestCancel_ByClient_Own_Pending(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Cancel(db, "barbearia-test", 1, 20, "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "cancelled" {
		t.Errorf("expected status=cancelled, got %s", result.Status)
	}
	if result.CancelledBy == nil || *result.CancelledBy != 20 {
		t.Error("expected cancelled_by=20")
	}
	if result.CancelledAt == nil {
		t.Error("expected cancelled_at to be set")
	}
}

// TestCancel_ByProfessional_Own_Confirmed verifica que profissional pode cancelar próprio agendamento confirmed.
func TestCancel_ByProfessional_Own_Confirmed(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Cancel(db, "barbearia-test", 2, 12, "professional")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "cancelled" {
		t.Errorf("expected status=cancelled, got %s", result.Status)
	}
}

// TestCancel_AlreadyCancelled_InvalidTransition verifica que cancelar agendamento já cancelado retorna ErrInvalidTransition.
func TestCancel_AlreadyCancelled_InvalidTransition(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Cancel(db, "barbearia-test", 3, 10, "owner")
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

// --- Complete ---

// TestComplete_ByProfessional_Confirmed_Success verifica que profissional pode concluir agendamento confirmed.
func TestComplete_ByProfessional_Confirmed_Success(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Complete(db, "barbearia-test", 2, 12, "professional")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("expected status=completed, got %s", result.Status)
	}
	if result.CompletedBy == nil || *result.CompletedBy != 12 {
		t.Error("expected completed_by=12")
	}
	if result.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

// TestComplete_ByClient_Forbidden verifica que cliente não pode concluir agendamento.
func TestComplete_ByClient_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Complete(db, "barbearia-test", 2, 20, "user")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// TestComplete_FromPending_InvalidTransition verifica que concluir agendamento pending retorna ErrInvalidTransition.
func TestComplete_FromPending_InvalidTransition(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Complete(db, "barbearia-test", 1, 10, "owner")
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

// --- NoShow ---

// TestNoShow_ByOwner_Confirmed_Success verifica que owner pode marcar agendamento confirmed como no_show.
func TestNoShow_ByOwner_Confirmed_Success(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := NoShow(db, "barbearia-test", 2, 10, "owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "no_show" {
		t.Errorf("expected status=no_show, got %s", result.Status)
	}
	if result.NoShowBy == nil || *result.NoShowBy != 10 {
		t.Error("expected no_show_by=10")
	}
	if result.NoShowAt == nil {
		t.Error("expected no_show_at to be set")
	}
}

// TestNoShow_ByClient_Forbidden verifica que cliente não pode marcar no-show.
func TestNoShow_ByClient_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := NoShow(db, "barbearia-test", 2, 20, "user")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// TestNoShow_FromPending_InvalidTransition verifica que marcar no-show em agendamento pending retorna ErrInvalidTransition.
func TestNoShow_FromPending_InvalidTransition(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := NoShow(db, "barbearia-test", 1, 10, "owner")
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}
