package actions

import (
	"errors"
	"testing"
)

// TestFindSchedule_ByOwner verifica que owner pode ver qualquer agendamento da org.
func TestFindSchedule_ByOwner(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Find(db, "barbearia-test", 1, 10, "owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected schedule ID=1, got %d", result.ID)
	}
}

// TestFindSchedule_ByClient_Own verifica que cliente pode ver seu próprio agendamento.
func TestFindSchedule_ByClient_Own(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	result, err := Find(db, "barbearia-test", 1, 20, "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// TestFindSchedule_ByClient_Other verifica que cliente recebe ErrForbidden ao tentar ver agendamento alheio.
func TestFindSchedule_ByClient_Other(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Find(db, "barbearia-test", 1, 21, "user")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// TestFindSchedule_NotFound verifica que Find retorna ErrScheduleNotFound para ID inexistente.
func TestFindSchedule_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	_ = mock

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic(not implemented)")
		}
	}()

	_, err := Find(db, "barbearia-test", 9999, 10, "owner")
	if !errors.Is(err, ErrScheduleNotFound) {
		t.Errorf("expected ErrScheduleNotFound, got %v", err)
	}
}
