package scheduleshttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Reschedule ---

// TestReschedule_200 verifica que PATCH /reschedule com dados válidos e cliente autenticado retorna 200.
func TestReschedule_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	body := bytes.NewBufferString(`{"scheduled_at": "2026-07-05T10:00:00-03:00"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/reschedule", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestReschedule_401 verifica que PATCH /reschedule sem autenticação retorna 401.
func TestReschedule_401(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouter(h)

	body := bytes.NewBufferString(`{"scheduled_at": "2026-07-05T10:00:00-03:00"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/reschedule", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestReschedule_403_OtherClient verifica que PATCH /reschedule com token de outro cliente retorna 403.
func TestReschedule_403_OtherClient(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 21, "user")

	body := bytes.NewBufferString(`{"scheduled_at": "2026-07-05T10:00:00-03:00"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/reschedule", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestReschedule_409_Conflict verifica que PATCH /reschedule com horário em conflito retorna 409.
func TestReschedule_409_Conflict(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	body := bytes.NewBufferString(`{"scheduled_at": "2026-07-01T17:00:00-03:00"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/reschedule", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestReschedule_422_InvalidTransition verifica que PATCH /reschedule em agendamento cancelled retorna 422.
func TestReschedule_422_InvalidTransition(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	body := bytes.NewBufferString(`{"scheduled_at": "2026-07-05T10:00:00-03:00"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/3/reschedule", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestReschedule_400_MissingScheduledAt verifica que PATCH /reschedule sem scheduled_at retorna 400.
func TestReschedule_400_MissingScheduledAt(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/reschedule", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- RescheduleHistory ---

// TestRescheduleHistory_200 verifica que GET /reschedule-history com owner retorna 200.
func TestRescheduleHistory_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules/2/reschedule-history", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestRescheduleHistory_401 verifica que GET /reschedule-history sem autenticação retorna 401.
func TestRescheduleHistory_401(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules/2/reschedule-history", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestRescheduleHistory_403_OtherClient verifica que GET /reschedule-history com outro cliente retorna 403.
func TestRescheduleHistory_403_OtherClient(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 21, "user")

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules/1/reschedule-history", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestRescheduleHistory_404 verifica que GET /reschedule-history com schedule inexistente retorna 404.
func TestRescheduleHistory_404(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules/9999/reschedule-history", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}
