package scheduleshttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Confirm ---

// TestConfirm_200 verifica que PATCH /confirm com profissional autenticado retorna 200.
func TestConfirm_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 12, "professional")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/confirm", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestConfirm_401 verifica que PATCH /confirm sem autenticação retorna 401.
func TestConfirm_401(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/confirm", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestConfirm_403_Client verifica que PATCH /confirm com token de cliente retorna 403.
func TestConfirm_403_Client(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/confirm", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestConfirm_422_InvalidTransition verifica que PATCH /confirm em agendamento já confirmado retorna 422.
func TestConfirm_422_InvalidTransition(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/2/confirm", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Cancel ---

// TestCancel_200 verifica que PATCH /cancel com cliente autenticado e próprio agendamento retorna 200.
func TestCancel_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/cancel", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCancel_401 verifica que PATCH /cancel sem autenticação retorna 401.
func TestCancel_401(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/cancel", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestCancel_422_AlreadyCancelled verifica que PATCH /cancel em agendamento já cancelado retorna 422.
func TestCancel_422_AlreadyCancelled(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/3/cancel", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Complete ---

// TestComplete_200 verifica que PATCH /complete com profissional em agendamento confirmed retorna 200.
func TestComplete_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 12, "professional")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/2/complete", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestComplete_403_Client verifica que PATCH /complete com token de cliente retorna 403.
func TestComplete_403_Client(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/2/complete", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestComplete_422_FromPending verifica que PATCH /complete em agendamento pending retorna 422.
func TestComplete_422_FromPending(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/complete", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- NoShow ---

// TestNoShow_200 verifica que PATCH /no-show com owner em agendamento confirmed retorna 200.
func TestNoShow_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/2/no-show", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestNoShow_403_Client verifica que PATCH /no-show com token de cliente retorna 403.
func TestNoShow_403_Client(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/2/no-show", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestNoShow_422_FromPending verifica que PATCH /no-show em agendamento pending retorna 422.
func TestNoShow_422_FromPending(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/schedules/1/no-show", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}
