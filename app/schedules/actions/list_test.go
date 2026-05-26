package actions

import (
	"errors"
	"testing"

	"barbearia-api/models"
)

// TestListSchedules_Owner verifica que owner lista todos os agendamentos da org.
func TestListSchedules_Owner(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := List(db, "barbearia-test", 10, "owner", models.ListSchedulesFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// TestListSchedules_Professional verifica que profissional vê apenas os próprios agendamentos.
func TestListSchedules_Professional(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := List(db, "barbearia-test", 12, "professional", models.ListSchedulesFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, s := range result {
		if s.ProfessionalID != 12 {
			t.Errorf("expected professional_id=12, got %d", s.ProfessionalID)
		}
	}
}

// TestListSchedules_Client_Forbidden verifica que cliente recebe ErrForbidden ao tentar listar.
func TestListSchedules_Client_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := List(db, "barbearia-test", 20, "user", models.ListSchedulesFilter{})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// TestListSchedules_FilterByStatus verifica que o filtro por status funciona corretamente.
func TestListSchedules_FilterByStatus(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := List(db, "barbearia-test", 10, "owner", models.ListSchedulesFilter{Status: "pending"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, s := range result {
		if s.Status != "pending" {
			t.Errorf("expected status=pending, got %s", s.Status)
		}
	}
}

// TestListSchedules_Empty verifica que List retorna slice vazio quando não há agendamentos.
func TestListSchedules_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := List(db, "barbearia-test", 10, "owner", models.ListSchedulesFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil slice, got nil")
	}
}
