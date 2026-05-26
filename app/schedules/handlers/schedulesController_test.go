package scheduleshttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Create ---

// TestCreate_201 verifica que POST /organizations/:slug/schedules com dados válidos retorna 201.
func TestCreate_201(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	body := bytes.NewBufferString(`{
		"service_id": 1,
		"professional_id": 12,
		"scheduled_at": "2026-06-10T15:00:00-03:00"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/schedules", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreate_401 verifica que POST sem autenticação retorna 401.
func TestCreate_401(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouter(h)

	body := bytes.NewBufferString(`{
		"service_id": 1,
		"professional_id": 12,
		"scheduled_at": "2026-06-10T15:00:00-03:00"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/schedules", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestCreate_409_Conflict verifica que POST com conflito de horário retorna 409.
func TestCreate_409_Conflict(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	body := bytes.NewBufferString(`{
		"service_id": 1,
		"professional_id": 12,
		"scheduled_at": "2026-06-10T13:00:00Z"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/schedules", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreate_422_PastScheduledAt verifica que POST com horário no passado retorna 422.
func TestCreate_422_PastScheduledAt(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	body := bytes.NewBufferString(`{
		"service_id": 1,
		"professional_id": 12,
		"scheduled_at": "2020-01-01T10:00:00-03:00"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/schedules", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreate_400_MissingFields verifica que POST sem campos obrigatórios retorna 400.
func TestCreate_400_MissingFields(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/schedules", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- List ---

// TestList_200 verifica que GET /organizations/:slug/schedules com owner autenticado retorna 200.
func TestList_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestList_401 verifica que GET sem autenticação retorna 401.
func TestList_401(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestList_403_Client verifica que GET com token de cliente retorna 403.
func TestList_403_Client(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Find ---

// TestFind_200 verifica que GET /organizations/:slug/schedules/:id com owner retorna 200.
func TestFind_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules/1", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestFind_403_OtherClient verifica que GET com token de outro cliente retorna 403.
func TestFind_403_OtherClient(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 21, "user")

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules/1", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestFind_404 verifica que GET com ID inexistente retorna 404.
func TestFind_404(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 10, "owner")

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/schedules/9999", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- MySchedules ---

// TestMySchedules_200 verifica que GET /my/schedules com autenticação retorna 200.
func TestMySchedules_200(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouterWithAuth(h, 20, "user")

	req := httptest.NewRequest(http.MethodGet, "/my/schedules", nil)
	w := httptest.NewRecorder()

	defer func() { recover() }()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestMySchedules_401 verifica que GET /my/schedules sem autenticação retorna 401.
func TestMySchedules_401(t *testing.T) {
	h, _ := newHandler(t)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/my/schedules", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
